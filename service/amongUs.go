// service/among_us.go
package service

import (
	"errors"
	"fmt"
	"go-backend/common"
	"math/rand"
	"sync"
)

/***************
 * Domain types
 ***************/

type TaskInfo struct {
	Key  string // 任務代號
	Info string // 任務說明
}

type Player struct {
	ID      string
	Role    string
	Task    string
	TaskInf string
}

// GameSession 為單場遊戲的狀態容器
type GameSession struct {
	id         string
	players    map[string]*Player
	impostorIx int // 0..NumPlayer-1
	foolIx     int // -1 表示沒有
	impTask    string
	roundTasks []string
	// 用不到就先不開放，若日後要支援 per-session 併發，再啟用
	// mu sync.RWMutex
}

func (gs *GameSession) ID() string { return gs.id }

/*******************
 * Configuration
 *******************/

const (
	roundTaskCount = 5
)

// 若要新增/調整任務，修改這裡即可
var (
	allTasks = []string{"最多人頭", "最少人頭", "最多死亡", "最少死亡", "cs最多", "cs最少", "戳仔", "符文智多星"}

	taskInfoMap = map[string]TaskInfo{
		"最多人頭":  {Key: "最多人頭", Info: "以最高人頭結束遊戲"},
		"最少人頭":  {Key: "最少人頭", Info: "以最少人頭結束遊戲"},
		"最多死亡":  {Key: "最多死亡", Info: "最多死亡"},
		"最少死亡":  {Key: "最少死亡", Info: "最少死亡"},
		"cs最多":  {Key: "cs最多", Info: "cs最多"},
		"cs最少":  {Key: "cs最少", Info: "cs最少"},
		"戳仔":    {Key: "戳仔", Info: "送你隊友下去"},
		"符文智多星": {Key: "符文智多星", Info: "只能使用以下符文並達成條件結束遊戲:\n致命節奏觸發時間少於20秒\n靈魂收割觸發次數小於10次\n相位衝擊觸發時間少於15秒\n不死之握觸發次數少於8次\n先發制人偷取金錢少於250"},
	}
)

/*******************
 * Game manager
 *******************/

type GameManager struct {
	games map[string]*GameSession
	mu    sync.RWMutex
}

func NewGameManager() *GameManager {
	return &GameManager{
		games: make(map[string]*GameSession),
	}
}

// List 回傳目前所有遊戲 ID
func (gm *GameManager) List() []string {
	gm.mu.RLock()
	defer gm.mu.RUnlock()
	ids := make([]string, 0, len(gm.games))
	for id := range gm.games {
		ids = append(ids, id)
	}
	return ids
}

// Create 建立新遊戲；num 參數未使用，保留相容性
func (gm *GameManager) Create(num string) (*GameSession, error) {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	n := common.NumPlayer
	if n <= 0 {
		return nil, errors.New("NumPlayer must be > 0")
	}

	impostorIx := rand.Intn(n) // [0, n)
	foolIx := -1
	// 以 0..999 與 FoolChance 比較，沿用舊邏輯但明確
	if rand.Intn(1000) < common.FoolChance {
		foolIx = (impostorIx + 1) % n
		if n == 1 {
			foolIx = -1 // 邏輯保護：單人局不產生呆呆鳥
		}
	}

	roundTasks := pickTasks(allTasks, roundTaskCount)
	impTask := roundTasks[rand.Intn(len(roundTasks))]

	g := &GameSession{
		id:         common.GetRandomString(8),
		players:    make(map[string]*Player),
		impostorIx: impostorIx,
		foolIx:     foolIx,
		impTask:    impTask,
		roundTasks: roundTasks,
	}
	gm.games[g.id] = g
	return g, nil
}

// EndGame 結束指定遊戲；gid == "ALL" 代表清空
func (gm *GameManager) EndGame(gid string) error {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	if gid == "ALL" {
		gm.games = make(map[string]*GameSession)
		return nil
	}
	if _, ok := gm.games[gid]; !ok {
		return fmt.Errorf("game %s not found", gid)
	}
	delete(gm.games, gid)
	return nil
}

// Join 讓玩家加入，並回傳角色與任務資訊
func (gm *GameManager) Join(playerID, gid string) (roleMsg, task, taskInf string, roundTasks []string, err error) {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	game, ok := gm.games[gid]
	if !ok {
		return "", "", "", nil, fmt.Errorf("game %s not found", gid)
	}

	// 已加入則直回
	if p, ok := game.players[playerID]; ok {
		return fmt.Sprintf("Player:%s, your Role:%s", playerID, p.Role), p.Task, p.TaskInf, game.roundTasks, nil
	}

	if len(game.players) >= common.NumPlayer {
		return "", "", "", nil, fmt.Errorf("game %s is full", gid)
	}

	// 以加入順序決定角色：第 k 個加入的人，其索引為 currentLen
	currentLen := len(game.players)
	role := "Crewmate"
	task = ""
	taskInf = "ejecting all Impostors."

	if currentLen == game.impostorIx {
		role = "Impostor"
		task = game.impTask
		taskInf = safeTaskInfo(task)
	} else if currentLen == game.foolIx {
		role = "隱藏職業:呆呆鳥"
		task = game.impTask + "(same task as Impostor)"
		taskInf = "你的任務是假的，你必須被淘汰才算獲勝(呆呆鳥)"
	}

	game.players[playerID] = &Player{
		ID:      playerID,
		Role:    role,
		Task:    task,
		TaskInf: taskInf,
	}

	common.SysDebug(fmt.Sprintf("%d joined; impostor=%d, fool=%d", len(game.players), game.impostorIx, game.foolIx))

	return fmt.Sprintf("Player:%s, your Role:%s", playerID, role), task, taskInf, game.roundTasks, nil
}

func (gm *GameManager) ListPlayers(gid string) ([]*Player, error) {
	gm.mu.RLock()
	defer gm.mu.RUnlock()

	game, ok := gm.games[gid]
	if !ok {
		return nil, fmt.Errorf("game %s not found", gid)
	}

	out := make([]*Player, 0, len(game.players))
	for _, p := range game.players {
		out = append(out, p)
	}
	return out, nil
}

func (gm *GameManager) GameInfo(gid string) (id string, roundTasks []string, impostorTask string, err error) {
	gm.mu.RLock()
	defer gm.mu.RUnlock()

	game, ok := gm.games[gid]
	if !ok {
		return "", nil, "", fmt.Errorf("game %s not found", gid)
	}
	return game.id, game.roundTasks, game.impTask, nil
}

/*******************
 * Utilities
 *******************/

// pickTasks 隨機挑選 count 個不重複任務
func pickTasks(pool []string, count int) []string {
	if count <= 0 {
		return []string{}
	}
	tmp := append([]string(nil), pool...) // 複製避免汙染原 slice
	rand.Shuffle(len(tmp), func(i, j int) { tmp[i], tmp[j] = tmp[j], tmp[i] })
	if count > len(tmp) {
		count = len(tmp)
	}
	out := make([]string, count)
	copy(out, tmp[:count])
	return out
}

func safeTaskInfo(task string) string {
	if ti, ok := taskInfoMap[task]; ok {
		return ti.Info
	}
	return ""
}

// GetPlayerSafe 以索引安全取得玩家指標
func GetPlayerSafe(players []*Player, idx int) (*Player, bool) {
	if idx >= 0 && idx < len(players) {
		return players[idx], true
	}
	return nil, false
}
