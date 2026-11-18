package server

import (
	"net"
	"sync"
	"testing"
	"time"

	"IM-system/internal/logger"
)

func TestShutdownClosesUnauthenticatedTCP(t *testing.T) {
	logger.Init()
	s := NewServer("", 0, nil, nil, nil, nil)
	serverConn, client := net.Pipe()
	defer client.Close()
	if !s.startTCPHandler(serverConn) {
		t.Fatal("connection rejected before shutdown")
	}
	finished := make(chan struct{})
	go func() {
		s.Shutdown()
		close(finished)
	}()
	select {
	case <-finished:
	case <-time.After(time.Second):
		t.Fatal("shutdown did not release unauthenticated TCP read")
	}
	s.lifecycleMu.Lock()
	remaining := len(s.tcpConnections)
	s.lifecycleMu.Unlock()
	if remaining != 0 {
		t.Fatalf("shutdown returned with %d tracked connections", remaining)
	}
	_ = client.SetReadDeadline(time.Now().Add(time.Second))
	if _, err := client.Read(make([]byte, 1)); err == nil {
		t.Fatal("client connection remains open")
	}
}

func TestTCPRegistrationAndShutdownRace(t *testing.T) {
	logger.Init()
	for range 30 {
		s := NewServer("", 0, nil, nil, nil, nil)
		serverConn, client := net.Pipe()
		var wg sync.WaitGroup
		wg.Add(2)
		go func() { defer wg.Done(); s.startTCPHandler(serverConn) }()
		go func() { defer wg.Done(); s.Shutdown() }()
		wg.Wait()
		s.lifecycleMu.Lock()
		remaining := len(s.tcpConnections)
		s.lifecycleMu.Unlock()
		_ = client.Close()
		if remaining != 0 {
			t.Fatal("connection escaped shutdown")
		}
		lateServer, lateClient := net.Pipe()
		if s.startTCPHandler(lateServer) {
			t.Fatal("accepted TCP connection after shutdown")
		}
		_ = lateClient.SetReadDeadline(time.Now().Add(time.Second))
		_, err := lateClient.Read(make([]byte, 1))
		_ = lateClient.Close()
		if err == nil {
			t.Fatal("rejected TCP connection was not closed")
		}
	}
}
