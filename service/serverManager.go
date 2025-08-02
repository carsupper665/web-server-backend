//service/serverManager.go

package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"sync"
	"time"
)

const (
	// MaxServersPerOwner 限制每個 owner 同時開啟的最大伺服器數量
	MaxServersPerOwner = 3
)

type Server struct {
	sid          string
	oid          string
	workDir      string
	maxMem       string
	minMem       string
	port         string
	cmd          *exec.Cmd
	stdin        io.Writer
	stdout       io.Reader
	logBuffer    *bytes.Buffer
	serverStatus string
	exp          time.Time
	sdc          func(string)
	args         []string
	mu           sync.Mutex
}

func NewServer(sid, oid, workDir, maxMem, minMem string, portStr string, callback func(string), args []string) *Server {
	return &Server{
		sid:          sid,
		oid:          oid,
		workDir:      workDir,
		maxMem:       maxMem,
		minMem:       minMem,
		port:         portStr,
		serverStatus: "stopped",
		sdc:          callback,
		args:         args,
		logBuffer:    &bytes.Buffer{},
	}
}

func (s *Server) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.serverStatus == "running" {
		return errors.New("server already running")
	}
	// 建立命令參數
	cmdArgs := []string{
		"-Xms" + s.minMem,
		"-Xmx" + s.maxMem,
		"-jar", "server.jar",
		"--port", s.port,
	}
	cmdArgs = append(cmdArgs, s.args...)
	cmd := exec.CommandContext(context.Background(), "java", cmdArgs...)
	cmd.Dir = s.workDir
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	cmd.Stderr = cmd.Stdout

	s.cmd = cmd
	s.stdin = stdin
	s.stdout = stdout
	s.logBuffer.Reset()

	if err := cmd.Start(); err != nil {
		return err
	}
	s.serverStatus = "running"

	go s.captureLogs()
	go s.waitAndCleanup()
	return nil
}

func (s *Server) captureLogs() {
	io.Copy(s.logBuffer, s.stdout)
}

func (s *Server) waitAndCleanup() {
	s.cmd.Wait()
	s.mu.Lock()
	s.serverStatus = "stopped"
	s.exp = time.Now().Add(5 * time.Minute)
	s.mu.Unlock()
	if s.sdc != nil {
		s.sdc(s.sid)
	}
}

func (s *Server) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cmd == nil || s.serverStatus != "running" {
		return errors.New("server not running")
	}
	if err := s.cmd.Process.Kill(); err != nil {
		return err
	}
	s.serverStatus = "stopped"
	s.exp = time.Now().Add(5 * time.Minute)
	return nil
}

func (s *Server) Restart() error {
	if err := s.Stop(); err != nil {
		return err
	}
	return s.Start()
}

func (s *Server) Status() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.serverStatus
}

func (s *Server) ReadLatestLog() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	data := s.logBuffer.String()
	if len(data) > 1024 {
		return data[len(data)-1024:]
	}
	return data
}

// SendCommand 發送指令到伺服器 stdin
func (s *Server) SendCommand(cmd string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.serverStatus != "running" {
		return errors.New("server not running")
	}
	_, err := io.WriteString(s.stdin, cmd+"\n")
	return err
}

func (s *Server) SetProperty(key, value string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.serverStatus == "running" {
		return errors.New("必須先停止 server 才能修改 server.properties")
	}
	return UpdateProperty(s.workDir, key, value)
}

// ---------------- ServerManager ----------------

type ServerManager struct {
	servers        map[string]*Server
	availablePorts []int
	usingPorts     map[int]string //port -> server ID
	mu             sync.RWMutex
}

func NewServerManager() *ServerManager {
	sm := &ServerManager{
		servers: make(map[string]*Server),
	}
	go sm.cleanupExpired()
	return sm
}

func (sm *ServerManager) IsExist(sid string) bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	_, ok := sm.servers[sid]
	return ok
}

func (sm *ServerManager) countByOwner(oid string) int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	count := 0
	for _, srv := range sm.servers {
		if srv.oid == oid {
			count++
		}
	}
	return count
}

func (sm *ServerManager) allocatePort() (int, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	if len(sm.availablePorts) == 0 {
		return 0, errors.New("no available ports")
	}
	port := sm.availablePorts[0]
	sm.availablePorts = sm.availablePorts[1:]
	sm.usingPorts[port] = ""
	return port, nil
}

// assignPortToServer 將 port 綁定給 sid
func (sm *ServerManager) assignPortToServer(port int, sid string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.usingPorts[port] = sid
}

// releasePort 釋放回 pool
func (sm *ServerManager) releasePort(portStr string) {
	var port int
	fmt.Sscanf(portStr, "%d", &port)
	sm.mu.Lock()
	defer sm.mu.Unlock()
	delete(sm.usingPorts, port)
	sm.availablePorts = append(sm.availablePorts, port)
}

func (sm *ServerManager) StartServer(sid, oid, workDir, maxMem, minMem string, args []string) (*Server, error) {
	if sm.countByOwner(oid) >= MaxServersPerOwner {
		return nil, fmt.Errorf("owner %s has reached max server limit of %d", oid, MaxServersPerOwner)
	}

	sm.mu.Lock()
	if _, exists := sm.servers[sid]; exists {
		return nil, fmt.Errorf("server with ID %s already exists", sid)
	}
	sm.mu.Unlock()

	var allocatedPort int
	p, err := sm.allocatePort()
	if err != nil {
		return nil, err
	}
	allocatedPort = p
	portStr := fmt.Sprintf("%d", p)

	srv := NewServer(sid, oid, workDir, maxMem, minMem, portStr, sm.shutDownServerCallback, args)
	sm.assignPortToServer(allocatedPort, sid)

	sm.mu.Lock()
	sm.servers[sid] = srv
	sm.mu.Unlock()

	if err := srv.Start(); err != nil {
		sm.mu.Lock()
		delete(sm.servers, sid)
		sm.releasePort(portStr)
		sm.mu.Unlock()
		return nil, err
	}
	return srv, nil
}

func (sm *ServerManager) StopServer(sid string) error {
	sm.mu.RLock()
	srv, exists := sm.servers[sid]
	sm.mu.RUnlock()
	if !exists {
		return fmt.Errorf("no server with ID %s", sid)
	}
	if err := srv.Stop(); err != nil {
		return err
	}
	// 不能重複釋放PORT 必須等到過期回收 在釋放
	// var p int
	// fmt.Sscanf(srv.port, "%d", &p)
	// sm.releasePort(p)
	return nil
}

func (sm *ServerManager) RestartServer(sid string) error {
	sm.mu.RLock()
	srv, exists := sm.servers[sid]
	sm.mu.RUnlock()
	if !exists {
		return fmt.Errorf("no server with ID %s", sid)
	}
	return srv.Restart()
}

func (sm *ServerManager) GetServerStatus(sid string) (string, error) {
	sm.mu.RLock()
	srv, exists := sm.servers[sid]
	sm.mu.RUnlock()
	if !exists {
		return "", fmt.Errorf("no server with ID %s", sid)
	}
	return srv.Status(), nil
}

func (sm *ServerManager) shutDownServerCallback(sid string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	// shut down server時必須釋放port
	srv := sm.servers[sid]
	sm.releasePort(srv.port)
	delete(sm.servers, sid)
}

func (sm *ServerManager) cleanupExpired() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		now := time.Now()
		sm.mu.Lock()
		for sid, srv := range sm.servers {
			if srv.Status() == "stopped" && srv.exp.Before(now) {
				sm.releasePort(srv.port)
				delete(sm.servers, sid)
			}
		}
		sm.mu.Unlock()
	}
}
