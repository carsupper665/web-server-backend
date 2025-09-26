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
	mu        sync.RWMutex
}

type Player struct {
	ID   string
	Role string
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

func (gm *GameManager) Create() (*gameSession, error) {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	if gm.games == nil {
		gm.games = make(map[string]*gameSession)
	}

	g := &gameSession{
		gid:       common.GetRandomString(8),
		players:   make(map[string]*Player),
		imposster: uint(rand.Intn(6)),
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

	if _, ok := gm.games[gid]; !ok {
		return "", fmt.Errorf("game %s not found", gid)
	}

	game := gm.games[gid]

	if p, ok := game.players[player]; ok {
		return p.Role, nil
	}

	role := "Crewmate"

	if len(game.players) == int(game.imposster)-1 {
		role = "Impostor"
	}

	game.players[player] = &Player{
		ID:   player,
		Role: role,
	}

	return role, nil

}
