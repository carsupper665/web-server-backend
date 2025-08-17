//service/serverManager.go

package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"go-backend/common"
	"io"
	"os/exec"
	"sync"
	"time"
)

const (
	// MaxServersPerOwner 限制每個 owner 同時開啟的最大伺服器數量
	MaxServersPerOwner = 3
)

var ErrAlreadyRunning = errors.New("server already running")
var ErrNotFound = errors.New("Server Not Found.")
var ErrMaxReached = errors.New("User has reached the maximum number of servers")

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
	mu           sync.RWMutex
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
		return ErrAlreadyRunning
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
	s.exp = time.Now().Add(3 * time.Minute)
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
	s.exp = time.Now().Add(3 * time.Minute)
	s.mu.Unlock()
}

func (s *Server) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cmd == nil || s.serverStatus != "running" {
		return errors.New("server not running")
	}

	_, _ = io.WriteString(s.stdin, "stop\n")

	timeout := 30 * time.Second
	done := make(chan error, 1)
	go func() {
		done <- s.cmd.Wait()
	}()

	select {
	case <-time.After(timeout):
		// 超時，強制 kill
		if s.cmd.Process != nil {
			_ = s.cmd.Process.Kill()
		}
		<-done // 等候 goroutine 結束（可以忽略 error）
	case err := <-done:
		if err != nil {
			common.SysError(err.Error())
		}
	}

	s.serverStatus = "stopped"
	s.exp = time.Now().Add(3 * time.Minute)
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

func (s *Server) Port() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.port
}

func (s *Server) ID() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.sid
}

func (s *Server) ReadLatestLog() string {
	data := s.logBuffer.String()
	if len(data) > (1024 * 8) {
		s.mu.Lock()
		defer s.mu.Unlock()
		return data[len(data)-(1024*8):]
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

func (s *Server) ShutDown() error {
	s.mu.RLock()
	status := s.serverStatus
	callback := s.sdc
	sid := s.sid
	s.mu.RUnlock()
	if status == "running" {
		if err := s.Stop(); err != nil {
			return err
		}
	}
	callback(sid)
	return nil
}

// ---------------- ServerManager ----------------

type ServerManager struct {
	servers        map[string]*Server
	availablePorts []int
	usingPorts     map[int]string //port -> server ID
	mu             sync.RWMutex
}

func NewServerManager(ports []int) *ServerManager {
	sm := &ServerManager{
		servers:        make(map[string]*Server),
		availablePorts: ports,
		usingPorts:     make(map[int]string),
	}
	go sm.cleanupExpired()
	return sm
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

func (sm *ServerManager) releasePortWithOutLock(portStr string) {
	var port int
	fmt.Sscanf(portStr, "%d", &port)
	delete(sm.usingPorts, port)
	sm.availablePorts = append(sm.availablePorts, port)
}

func (sm *ServerManager) StartServer(sid, oid, workDir, maxMem, minMem string, args []string) (*Server, error) {
	if sm.countByOwner(oid) >= MaxServersPerOwner {
		return nil, ErrMaxReached
	}

	sm.mu.Lock()
	if s, exists := sm.servers[sid]; exists {
		err := s.Start()
		if err != nil && errors.Is(err, ErrAlreadyRunning) {
			sm.mu.Unlock()
			common.SysDebug("server already running sid: " + sid)
			return s, nil
		} else if err != nil {
			panic("Unknow error:" + err.Error())
		}
		sm.mu.Unlock()
		common.SysDebug("Server is running: " + sid)
		return s, nil // Server Running successfully
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
	common.SysDebug("Server Start: " + sid)
	return srv, nil
}

func (sm *ServerManager) SendCommand(sid string, cmd string) error {
	sm.mu.RLock()
	srv, exists := sm.servers[sid]
	sm.mu.RUnlock()

	if !exists {
		return ErrNotFound
	}

	if err := srv.SendCommand(cmd); err != nil {
		return err
	}

	return nil
}

func (sm *ServerManager) StopServer(sid string) error {
	sm.mu.RLock()
	srv, exists := sm.servers[sid]
	sm.mu.RUnlock()
	if !exists {
		return ErrNotFound
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
		return ErrNotFound
	}
	return srv.Restart()
}

func (sm *ServerManager) GetServerStatus(sid string) (string, error) {
	sm.mu.RLock()
	srv, exists := sm.servers[sid]
	sm.mu.RUnlock()
	if !exists {
		return "stopped", nil
	}
	return srv.Status(), nil
}

func (sm *ServerManager) shutDownServerCallback(sid string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	// shut down server時必須釋放port
	srv := sm.servers[sid]
	sm.releasePortWithOutLock(srv.port)
	delete(sm.servers, sid)
}

func (sm *ServerManager) ReadLatestLog(sid string) (string, error) {
	sm.mu.RLock()
	srv, exists := sm.servers[sid]
	sm.mu.RUnlock()

	if !exists {
		return "", ErrNotFound
	}

	return srv.ReadLatestLog(), nil
}

func (sm *ServerManager) cleanupExpired() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		now := time.Now()
		sm.mu.Lock()
		for sid, srv := range sm.servers {
			s := srv.Status()
			isExp := srv.exp.Before(now)
			if s == "stopped" && isExp {
				sm.releasePortWithOutLock(srv.port)
				delete(sm.servers, sid)
				common.SysLog(fmt.Sprintf("Server: %s del, port: %s", sid, srv.port))
			}
		}
		sm.mu.Unlock()
	}
}
