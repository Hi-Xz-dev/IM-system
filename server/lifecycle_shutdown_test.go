package server

import (
	"net"
	"sync"
	"testing"
	"time"
)

func TestStartReturnsListenError(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	s := NewServer("127.0.0.1", listener.Addr().(*net.TCPAddr).Port, nil, nil, nil, nil)
	if err := s.Start(); err == nil {
		t.Fatal("occupied TCP port must return an error")
	}
}

func TestBackgroundWorkersStopOnShutdown(t *testing.T) {
	s := NewServer("", 0, nil, nil, nil, nil)
	s.startBackgroundWorkers()
	stopped := make(chan struct{})
	go func() {
		s.Shutdown()
		close(stopped)
	}()
	select {
	case <-stopped:
	case <-time.After(time.Second):
		t.Fatal("background workers did not stop promptly")
	}
	s.Shutdown() // 重复关闭不应重复 close(done) 而 panic。
}

func TestBackgroundStartAndShutdownRace(t *testing.T) {
	for range 50 {
		s := NewServer("", 0, nil, nil, nil, nil)
		var wg sync.WaitGroup
		wg.Add(2)
		go func() { defer wg.Done(); s.startBackgroundWorkers() }()
		go func() { defer wg.Done(); s.Shutdown() }()
		wg.Wait()
		s.startBackgroundWorkers() // 关闭后不得再次启动后台任务。
		s.workerWG.Wait()
		select {
		case <-s.done:
		default:
			t.Fatal("shutdown signal was not closed")
		}
	}
}

func TestShutdownUnblocksBroadcastWithFullQueue(t *testing.T) {
	s := NewServer("", 0, nil, nil, nil, nil)
	u := newRoomTestSession(1)
	s.Online(u)
	for len(s.Message) < cap(s.Message) {
		s.Message <- "occupied"
	}
	done := make(chan error, 1)
	go func() { done <- s.BroadcastSystemMessage(u, "hello") }()
	s.Shutdown()
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("broadcast should fail when shutdown interrupts enqueue")
		}
	case <-time.After(time.Second):
		t.Fatal("broadcast remained blocked after shutdown")
	}
}
