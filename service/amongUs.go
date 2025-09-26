// service/amongUs.go
package service

import (
	"fmt"
	"go-backend/common"
	"math/rand"
	"sync"
)

type gameSession struct {
	gid       string
	players   map[string]*Player
	imposster uint
	// mu        sync.RWMutex
}

type Player struct {
	ID   string
	Role string
	Task string
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

	n := common.GetEnvOrDefault("num", 5)

	g := &gameSession{
		gid:       common.GetRandomString(8),
		players:   make(map[string]*Player),
		imposster: uint(rand.Intn(n + 1)), // 0 到 n 之間
	}

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

func (gm *GameManager) Join(player, gid string) (string, error) {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	game, ok := gm.games[gid]
	if !ok {
		return "", fmt.Errorf("game %s not found", gid)
	}

	// 已存在就直接回傳
	if p, ok := game.players[player]; ok {
		return p.Role, nil
	}

	role := "Crewmate"
	task := ""

	if len(game.players) == int(game.imposster) {
		role = "Impostor"
		// 隨機一個任務
		tasks := []string{"人頭", "死亡"}
		task = tasks[rand.Intn(len(tasks))]
	}

	game.players[player] = &Player{
		ID:   player,
		Role: role,
		Task: task,
	}

	return role + task, nil
}
