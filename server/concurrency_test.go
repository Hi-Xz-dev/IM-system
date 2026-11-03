package server

import (
	"net"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"IM-system/internal/logger"
)

type lifecycleTestListener struct{ closed atomic.Bool }

func (l *lifecycleTestListener) Accept() (net.Conn, error) { return nil, net.ErrClosed }
func (l *lifecycleTestListener) Close() error              { l.closed.Store(true); return nil }
func (l *lifecycleTestListener) Addr() net.Addr            { return nil }

func TestConcurrentListenerRegistrationAndShutdown(t *testing.T) {
	for range 50 {
		s := NewServer("127.0.0.1", 0, nil, nil, nil)
		listener := &lifecycleTestListener{}
		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()
			// 与 Start 一样，关闭已经开始时不再发布新监听器。
			if !s.registerListener(listener) {
				_ = listener.Close()
			}
		}()
		go func() { defer wg.Done(); s.Shutdown() }()
		wg.Wait()
		if !s.isShuttingDown() || !listener.closed.Load() {
			t.Fatal("shutdown must not miss a concurrently registered listener")
		}
		if s.registerListener(&lifecycleTestListener{}) {
			t.Fatal("must reject listener registration after shutdown")
		}
	}
}

func TestConcurrentActivityAndInactiveScan(t *testing.T) {
	s := NewServer("127.0.0.1", 0, nil, nil, nil)
	u := newRoomTestSession(1)
	u.UpdateActiveTime()
	s.Online(u)
	if len(s.inactiveUsers(time.Now().Add(11*time.Minute))) != 1 {
		t.Fatal("scan must identify an inactive session")
	}
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		for range 1000 {
			u.UpdateActiveTime()
		}
	}()
	go func() {
		defer wg.Done()
		for range 1000 {
			if len(s.inactiveUsers(time.Now())) != 0 {
				t.Error("active session must not be marked inactive")
			}
		}
	}()
	wg.Wait()
}

func TestConcurrentPrivateChatAndSessionChurn(t *testing.T) {
	logger.Init()
	s := NewServer("127.0.0.1", 0, nil, nil, nil)
	sender, target := newRoomTestSession(1), newRoomTestSession(2)
	s.Online(sender)
	s.Online(target)
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		for range 20 {
			s.PrivateChat(sender, 2, "hello")
		}
	}()
	go func() {
		defer wg.Done()
		for range 20 {
			session := newRoomTestSession(2)
			s.Online(session)
			s.Offline(session)
		}
	}()
	wg.Wait()
	if sessions := s.OnlineUsers[2]; len(sessions) != 1 || sessions[0] != target {
		t.Fatal("session churn must preserve the original target connection")
	}
}
