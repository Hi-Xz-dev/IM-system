package user

import (
	"errors"
	"sync"
	"testing"
	"time"
)

// 模拟网络写入卡住；Close 必须能解除阻塞。
type stalledConnection struct {
	started chan struct{}
	release chan struct{}
	written chan struct{}
	closed  chan struct{}
	once    sync.Once
}

func (c *stalledConnection) Write([]byte) error {
	c.started <- struct{}{}
	select {
	case <-c.closed:
		return errors.New("closed")
	case <-c.release:
		c.written <- struct{}{}
		return nil
	}
}

func (c *stalledConnection) Close() error {
	c.once.Do(func() { close(c.closed) })
	return nil
}

func newStalledUser() (*User, *stalledConnection) {
	c := &stalledConnection{
		started: make(chan struct{}, 10), release: make(chan struct{}),
		written: make(chan struct{}, 10), closed: make(chan struct{}),
	}
	u := NewUser(c, 1, "slow", "test")
	u.C = make(chan string, 1)
	u.queueFullTimeout = 100 * time.Millisecond
	return u, c
}

func awaitSignal(t *testing.T, ch <-chan struct{}) {
	t.Helper()
	select {
	case <-ch:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for connection event")
	}
}

func TestSustainedCongestionClosesOnlySlowConnection(t *testing.T) {
	u, c := newStalledUser()
	t.Cleanup(u.Close)
	disconnect := make(chan *User, 1)
	go u.ListenMessage(disconnect)
	if err := u.SendMsg("first"); err != nil {
		t.Fatal(err)
	}
	awaitSignal(t, c.started)
	if err := u.SendMsg("second"); err != nil {
		t.Fatal(err)
	}
	if err := u.SendMsg("third"); !errors.Is(err, ErrSendQueueFull) {
		t.Fatalf("expected queue full, got %v", err)
	}
	// 慢连接入队失败不会阻塞另一个连接。
	healthy := NewUser(&fakeConnection{messages: make(chan string, 1)}, 2, "healthy", "test")
	t.Cleanup(healthy.Close)
	if err := healthy.SendMsg("hello"); err != nil {
		t.Fatal(err)
	}
	awaitSignal(t, c.closed)
	select {
	case got := <-disconnect:
		if got != u {
			t.Fatal("wrong session disconnected")
		}
	case <-time.After(time.Second):
		t.Fatal("writer did not exit after congestion timeout")
	}
	if err := u.SendMsg("after close"); !errors.Is(err, ErrConnectionClosed) {
		t.Fatalf("expected closed error, got %v", err)
	}
	if err := healthy.SendMsg("still online"); err != nil {
		t.Fatal(err)
	}
}

func TestTransientCongestionRecovers(t *testing.T) {
	u, c := newStalledUser()
	t.Cleanup(u.Close)
	disconnect := make(chan *User, 1)
	go u.ListenMessage(disconnect)
	if err := u.SendMsg("first"); err != nil {
		t.Fatal(err)
	}
	awaitSignal(t, c.started)
	if err := u.SendMsg("second"); err != nil {
		t.Fatal(err)
	}
	if err := u.SendMsg("third"); !errors.Is(err, ErrSendQueueFull) {
		t.Fatal(err)
	}
	close(c.release)
	awaitSignal(t, c.written)
	awaitSignal(t, c.written)
	// 超过原拥塞截止时间，确认成功写入后旧计时器不会误关连接。
	select {
	case <-c.closed:
		t.Fatal("transient congestion must not close a recovered session")
	case <-time.After(2 * u.queueFullTimeout):
	}
	if err := u.SendMsg("recovered"); err != nil {
		t.Fatal(err)
	}
	awaitSignal(t, c.written)
}

func TestConcurrentSendAndCloseStopsWriter(t *testing.T) {
	u, _ := newStalledUser()
	disconnect := make(chan *User, 1)
	go u.ListenMessage(disconnect)
	var wg sync.WaitGroup
	for range 10 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = u.SendMsg("concurrent")
			u.Close()
		}()
	}
	wg.Wait()
	select {
	case <-disconnect:
	case <-time.After(time.Second):
		t.Fatal("close must stop the writer even without queued messages")
	}
}

func TestCloseStopsIdleWriterWithFullDisconnectQueue(t *testing.T) {
	u, c := newStalledUser()
	disconnect := make(chan *User, 1)
	disconnect <- &User{}
	exited := make(chan struct{})
	go func() {
		u.ListenMessage(disconnect)
		close(exited)
	}()
	u.Close()
	u.Close()
	awaitSignal(t, exited)
	awaitSignal(t, c.closed)
}
