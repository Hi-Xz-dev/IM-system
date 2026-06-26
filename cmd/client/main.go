package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
)

type Client struct {
	ServerIp   string
	ServerPort int
	Name       string
	conn       net.Conn
	flag       int
}

func NewClient(serverIp string, serverPort int) *Client {
	client := &Client{
		ServerIp:   serverIp,
		ServerPort: serverPort,
		flag:       999,
	}
	conn, err := net.Dial("tcp", net.JoinHostPort(serverIp, strconv.Itoa(serverPort)))
	if err != nil {
		fmt.Println("net.Dial error:", err)
		return nil
	}
	client.conn = conn
	return client
}

// 发送命令到服务端
func (c *Client) send(cmd string) {
	_, err := c.conn.Write([]byte(cmd + "\n"))
	if err != nil {
		fmt.Println("conn Write err:", err)
	}
}

// 主菜单
func (c *Client) menu() bool {
	var flag int
	fmt.Println("\n===== 主菜单 =====")
	fmt.Println("1.公聊模式")
	fmt.Println("2.私聊模式")
	fmt.Println("3.更新用户名")
	fmt.Println("4.房间功能")
	fmt.Println("5.帮助")
	fmt.Println("0.退出")

	fmt.Scanln(&flag)

	if flag >= 0 && flag <= 5 {
		c.flag = flag
		return true
	}
	fmt.Println(">>>>>请输入合法范围内的数字<<<<<<")
	return false
}

// 房间子菜单
func (c *Client) roomMenu() {
	for {
		var flag int
		fmt.Println("\n===== 房间功能 =====")
		fmt.Println("1.创建房间")
		fmt.Println("2.加入房间")
		fmt.Println("3.离开房间")
		fmt.Println("4.查看房间列表")
		fmt.Println("5.当前房间")
		fmt.Println("6.房间成员")
		fmt.Println("7.房间聊天")
		fmt.Println("0.返回")

		fmt.Scanln(&flag)

		var input string
		switch flag {
		case 1:
			fmt.Println(">>>请输入房间名:")
			fmt.Scanln(&input)
			c.send("create|" + input)
		case 2:
			fmt.Println(">>>请输入房间名:")
			fmt.Scanln(&input)
			c.send("join|" + input)
		case 3:
			c.send("leave")
		case 4:
			c.send("rooms")
		case 5:
			c.send("where")
		case 6:
			c.send("members")
		case 7:
			c.roomChat()
		case 0:
			return
		default:
			fmt.Println(">>>>>请输入合法范围内的数字<<<<<<")
		}
	}
}

// 房间聊天
func (c *Client) roomChat() {
	fmt.Println(">>>>>请输入聊天内容【群聊】, exit退出。")
	var input string
	for {
		fmt.Scanln(&input)
		if input == "exit" {
			return
		}
		if len(input) != 0 {
			c.send("room|" + input)
		}
		fmt.Println(">>>>>请输入聊天内容【群聊】, exit退出。")
	}
}

// 处理服务端返回消息
func (c *Client) DealResponse() {
	io.Copy(os.Stdout, c.conn)
}

// 更新用户名
func (c *Client) UpdataName() {
	fmt.Println(">>>请输入用户名:")
	fmt.Scanln(&c.Name)
	c.send("rename|" + c.Name)
}

// 私聊模式
func (c *Client) PrivateChat() {
	c.send("who")
	fmt.Println(">>>>>请输入聊天对象[用户名], exit退出。")
	var remoteName string
	fmt.Scanln(&remoteName)

	for remoteName != "exit" {
		fmt.Println(">>>>>请输入聊天内容, exit退出。")
		var chatMsg string
		fmt.Scanln(&chatMsg)

		for chatMsg != "exit" {
			if len(chatMsg) != 0 {
				c.send("to|" + remoteName + "|" + chatMsg)
			}
			fmt.Println(">>>>>请输入聊天内容, exit退出。")
			fmt.Scanln(&chatMsg)
		}
		c.send("who")
		fmt.Println(">>>>>请输入聊天对象[用户名], exit退出。")
		fmt.Scanln(&remoteName)
	}
}

// 公聊模式
func (c *Client) PublicChat() {
	fmt.Println(">>>>>请输入聊天内容【公聊】, exit退出。")
	var chatMsg string
	fmt.Scanln(&chatMsg)
	for chatMsg != "exit" {
		if len(chatMsg) != 0 {
			c.send(chatMsg)
		}
		fmt.Println(">>>>>请输入聊天内容【公聊】, exit退出。")
		fmt.Scanln(&chatMsg)
	}
}

// 帮助
func (c *Client) Help() {
	fmt.Println("\n===== 命令列表 =====")
	fmt.Println("who              查看在线用户")
	fmt.Println("rename|名字      修改昵称")
	fmt.Println("to|用户|消息     私聊")
	fmt.Println("rooms            查看房间")
	fmt.Println("create|房间      创建房间")
	fmt.Println("join|房间        加入房间")
	fmt.Println("leave            离开房间")
	fmt.Println("room|消息        房间聊天")
	fmt.Println("members          房间成员")
	fmt.Println("where            当前房间")
	fmt.Println("help             帮助")
	fmt.Println("quit             退出")
	fmt.Println("任何文本          公聊消息")
}

// 主循环
func (c *Client) Run() {
	for c.flag != 0 {
		for c.menu() != true {
		}
		switch c.flag {
		case 1:
			c.PublicChat()
		case 2:
			c.PrivateChat()
		case 3:
			c.UpdataName()
		case 4:
			c.roomMenu()
		case 5:
			c.Help()
		case 0:
			c.send("quit")
			return
		}
	}
}

var serverIp string
var serverPort int

func init() {
	flag.StringVar(&serverIp, "ip", "127.0.0.1", "设置服务器IP地址(默认127.0.0.1)")
	flag.IntVar(&serverPort, "port", 8080, "设置服务器端口地址(默认8080)")
}

func main() {
	flag.Parse()
	client := NewClient(serverIp, serverPort)
	if client == nil {
		fmt.Println(">>>>>链接服务器失败.....")
		return
	}
	go client.DealResponse()
	fmt.Println(">>>>>链接服务器成功.....")
	client.Run()
}
