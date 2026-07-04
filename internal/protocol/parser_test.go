package protocol

import (
	"reflect"
	"testing"

	"IM-system/internal/domain"
)

//测试函数 函数名为Test开头
//这一类测试叫单元测试，因为它只测一个很小的函数，不需要启动 TCP、不需要启动 Gin、不需要连接数据库。
//也算做 回归测试 以后修改的时候，验证功能有没有被破坏

func TestParse(t *testing.T) {
	//定义一个匿名结构体切片，每一项都是一个测试用例
	tests := []struct {
		name     string
		input    string
		wantType domain.CommandType
		wantArgs []string
		wantRaw  string
	}{
		{
			name:     "who command",
			input:    "who",
			wantType: domain.CmdWho,
			wantArgs: []string{},
			wantRaw:  "who",
		},
		{
			name:     "rename command",
			input:    "rename|Tom",
			wantType: domain.CmdRename,
			wantArgs: []string{"Tom"},
			wantRaw:  "rename|Tom",
		},
		{
			name:     "private chat command",
			input:    "to|Jack|hello",
			wantType: domain.CmdPrivate,
			wantArgs: []string{"Jack", "hello"},
			wantRaw:  "to|Jack|hello",
		},
		{
			name:     "empty command",
			input:    "",
			wantType: domain.CmdUnknown,
			wantArgs: nil,
			wantRaw:  "",
		},
		{
			name:     "show rooms command",
			input:    "rooms",
			wantType: domain.CmdRooms,
			wantArgs: []string{},
			wantRaw:  "rooms",
		},
		{
			name:     "Create room command",
			input:    "create|golang",
			wantType: domain.CmdCreate,
			wantArgs: []string{"golang"},
			wantRaw:  "create|golang",
		},
		{
			name:     "join room command",
			input:    "join|golang",
			wantType: domain.CmdJoin,
			wantArgs: []string{"golang"},
			wantRaw:  "join|golang",
		},
		{
			name:     "leave room command",
			input:    "leave",
			wantType: domain.CmdLeave,
			wantArgs: []string{},
			wantRaw:  "leave",
		},
		{
			name:     "help command",
			input:    "help",
			wantType: domain.CmdHelp,
			wantArgs: []string{},
			wantRaw:  "help",
		},
		{
			name:     "where command",
			input:    "where",
			wantType: domain.CmdWhere,
			wantArgs: []string{},
			wantRaw:  "where",
		},
		{
			name:     "members command",
			input:    "members",
			wantType: domain.CmdMembers,
			wantArgs: []string{},
			wantRaw:  "members",
		},
		{
			name:     "public command",
			input:    "hello everyone",
			wantType: domain.CmdPublic,
			wantArgs: []string{},
			wantRaw:  "hello everyone",
		},
	}
	//t *testing.T 报告测试结果
	//Fatalf 相当于 Printf return 但会返回 测试失败
	//t.Error / t.Errorf 记录错误。 继续执行后面的判断。
	//%v 打印任意类型
	for _, tt := range tests {
		//把每一条测试数据变成一个独立的子测试（Subtest)
		t.Run(tt.name, func(t *testing.T) {
			got := Parse(tt.input)

			if got.Type != tt.wantType {
				t.Fatalf("Type: want %v, got %v", tt.wantType, got.Type)
			}
			//递归比较内容（深度比较）
			if !reflect.DeepEqual(got.Args, tt.wantArgs) {
				t.Fatalf("Args: want %v, got %v", tt.wantArgs, got.Args)
			}

			if got.Raw != tt.wantRaw {
				t.Fatalf("Raw: want %v, got %v", tt.wantRaw, got.Raw)
			}
		})
	}
}

// 性能测试
func BenchmarkParse(b *testing.B) {
	input := "to|TOm|hello"
	for i := 0; i < b.N; i++ {
		_ = Parse(input) //防止编译器优化（Compiler Optimization）
	}
}
