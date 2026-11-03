package user

import (
	"sync"
	"testing"
)

func TestConcurrentActivityReadWrite(t *testing.T) {
	u := NewUser(nil, 1, "test", "test")
	if u.LastActiveTime() == 0 {
		t.Fatal("new session must have an initial activity time")
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
			if u.LastActiveTime() == 0 {
				t.Error("activity time unexpectedly zero")
			}
		}
	}()
	wg.Wait()
}
