package server

import (
	"testing"

	"IM-system/user"
)

func TestOnlineOffline(t *testing.T) {
	s := NewServer("127.0.0.1", 8080)

	u := &user.User{
		Name: "Tom",
		Addr: "127.0.0.1:10001",
		C:    make(chan string, 100),
	}

	s.Online(u)

	if _, ok := s.OnlineUsers["Tom"]; !ok {
		t.Fatalf("expected user Tom online")
	}
	//测试过程中 发现业务逻辑和网络资源释放有一点耦合
	s.Offline(u)

	if _, ok := s.OnlineUsers["Tom"]; ok {
		t.Fatalf("expected user Tom offline")
	}
}
func TestOfflineDoubleCall(t *testing.T) {
	s := NewServer("127.0.0.1", 8080)

	u := &user.User{
		Name: "Tom",
		Addr: "127.0.0.1:10001",
		C:    make(chan string, 100),

	}

	s.Online(u)
	//Offline 调用两次，程序不崩，最终用户仍然是下线状态。
	//但它不能完全证明“第二次没有执行完整下线逻辑”。
	s.Offline(u)
	if _, ok := s.OnlineUsers["Tom"]; ok {
		t.Fatalf("expected user Tom offline")
	}
	if !u.IsClosed {
		t.Fatalf("expected IsClosed=true")
	}
	// 这里检查：
	// 1. OnlineUsers 里面没有 Tom
	// 2. u.IsClosed 是 true
}
