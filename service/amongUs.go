// service/amongUs.go
package service

import (
	"fmt"
	"go-backend/common"
	"math/rand"
	"sync"
)

type gameSession struct {
	gid        string
	players    map[string]*Player
	imposster  uint
	fool       uint
	impTask    string
	roundTasks []string

	// mu        sync.RWMutex
}

type taskInfo struct {
	T    string
	Tinf string
}

// var allTasks = []string{"最多人頭", "最少人頭", "最多死亡", "最少死亡", "kda1.5以下", "kda8.5以上", "cs最多", "cs最少", "戳仔", "符文智多星"}
var allTasks = []string{"最多人頭", "最少人頭", "最多死亡", "最少死亡", "cs最多", "cs最少", "戳仔", "符文智多星", "呆呆鳥"}
var info = map[string]*taskInfo{
	"最多人頭": {T: "最多人頭", Tinf: "以最高人頭結束遊戲"},
	"最少人頭": {T: "最少人頭", Tinf: "以最少人頭結束遊戲"},
	"最多死亡": {T: "最多死亡", Tinf: "最多死亡"},
	"最少死亡": {T: "最少死亡", Tinf: "最少死亡"},
	// "kda2以下":   {T: "kda1.5以下", Tinf: "kda維持在1.5以下結束遊戲，(K+A)/D小於2"},
	// "kda8.5以上": {T: "kda8.5以上", Tinf: "kda維持在8.5以上結束遊戲，(K+A)/D大於8.5"},
	"cs最多":  {T: "cs最多", Tinf: "cs最多"},
	"cs最少":  {T: "cs最少", Tinf: "cs最少"},
	"呆呆鳥":   {T: "呆呆鳥", Tinf: "你的任務是假的，你必須被淘汰才算獲勝(呆呆鳥)"},
	"戳仔":    {T: "戳仔", Tinf: "送你隊友下去"},
	"符文智多星": {T: "符文智多星", Tinf: "只能使用以下符文並達成條件結束遊戲:\n致命節奏觸發時間少於20秒\n靈魂收割觸發次數小於10次\n相位衝擊觸發時間少於15秒\n不死之握觸發次數少於8次\n先發制人偷取金錢少於250 "},
}

type Player struct {
	ID      string
	Role    string
	Task    string
	TaskInf string
	// mu   sync.RWMutex
}

type GameManager struct {
	games map[string]*gameSession
	mu    sync.RWMutex
}

func NewGameManager() *GameManager {
	return &GameManager{}
}

func (gs *gameSession) GetId() string {
	return gs.gid
}

func (gm *GameManager) List() []string {
	gm.mu.RLock()
	defer gm.mu.RUnlock()

	gamesId := make([]string, 0, len(gm.games))

	for id := range gm.games {
		gamesId = append(gamesId, id)
	}

	return gamesId
}

func (gm *GameManager) Create(num string) (*gameSession, error) {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	if gm.games == nil {
		gm.games = make(map[string]*gameSession)
	}

	n := common.NumPlayer
	imp := uint(rand.Intn(n + 1)) // 0 到 n 之間
	f := -1
	c := uint(rand.Intn(999))
	if c <= uint(common.FoolChance) {
		f = (int(imp + 1))
		if imp == 4 {
			f = 0
		}
	}

	rt := getRandomTasks()

	g := &gameSession{
		gid:        common.GetRandomString(8),
		players:    make(map[string]*Player),
		imposster:  imp,
		fool:       uint(f),
		impTask:    "",
		roundTasks: rt,
	}
	// 呆呆鵝 首場100 跟 內鬼同任務
	gm.games[g.gid] = g
	return g, nil
}

func (gm *GameManager) EndGame(gid string) error {
	gm.mu.Lock()
	defer gm.mu.Unlock()
	if gid == "ALL" {
		gm.games = make(map[string]*gameSession) // del all
		return nil
	}

	if _, ok := gm.games[gid]; !ok {
		return fmt.Errorf("game %s not found", gid)
	}

	delete(gm.games, gid)
	return nil
}

func (gm *GameManager) Join(player, gid string) (string, string, string, []string, error) {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	game, ok := gm.games[gid]
	if !ok {
		return "", "", "", []string{}, fmt.Errorf("game %s not found", gid)
	}

	// 已存在就直接回傳
	if p, ok := game.players[player]; ok {
		return p.Role, p.Task, p.TaskInf, game.roundTasks, nil
	}

	if len(game.players) >= common.NumPlayer {
		return "", "", "", []string{}, fmt.Errorf("game %s is full", gid)
	}

	role := "Crewmate"
	task := ""
	taskInf := "ejecting all Impostors."

	if len(game.players) == int(game.imposster) {
		role = "Impostor"
		// 隨機一個任務
		if game.impTask == "" {
			tasks := game.roundTasks
			r := rand.Intn(len(tasks))
			task = tasks[r]
			taskInf = info[task].Tinf

			game.impTask = task

		} else {
			task = game.impTask
			taskInf = info[task].Tinf
		}

	}
	common.SysDebug(fmt.Sprintf("%d/%d/imp:%d", len(game.players), int(game.fool), int(game.imposster)))

	if len(game.players) == int(game.fool) {
		// 把 map 轉 slice 再取 index
		role = "隱藏職業:呆呆鳥"
		if game.impTask != "" {
			task = game.impTask + "(same task as Impostor)"
		} else {
			tasks := game.roundTasks
			r := rand.Intn(len(tasks))
			task = tasks[r]
		}

		taskInf = "你的任務是假的，你必須被淘汰才算獲勝(呆呆鳥)，目前沒人知道這個職業，你如果主動詢問就會失去優勢"
	}

	game.players[player] = &Player{
		ID:      player,
		Role:    role,
		Task:    task,
		TaskInf: taskInf,
	}

	return fmt.Sprintf("Player:%s, your Role:%s", player, role), task, taskInf, game.roundTasks, nil
}

func (gm *GameManager) ListPlayers(gid string) ([]*Player, error) {
	gm.mu.RLock()
	defer gm.mu.RUnlock()

	game, ok := gm.games[gid]
	if !ok {
		return nil, fmt.Errorf("game %s not found", gid)
	}

	players := make([]*Player, 0, len(game.players))
	for _, p := range game.players {
		players = append(players, p)
	}
	return players, nil
}

func getRandomTasks() []string {

	rand.Shuffle(len(allTasks), func(i, j int) {
		allTasks[i], allTasks[j] = allTasks[j], allTasks[i]
	})

	return allTasks[:5]
}

func GetPlayer(players []Player, idx int) (Player, bool) {
	if idx >= 0 && idx < len(players) {
		return players[idx], true
	}
	return Player{}, false
}
