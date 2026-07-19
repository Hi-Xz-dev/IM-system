package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strconv"
)

type Client struct {
	ServerIP   string
	ServerPort int
	Name       string
	conn       net.Conn
	flag       int
	Token      string
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`

	Data struct {
		UserID   int64  `json:"UserID"`
		Username string `json:"Username"`
		Nickname string `json:"Nickname"`
		Token    string `json:"Token"`
	} `json:"data"`
}

func NewClient(ip string, port int) *Client {

	return &Client{
		ServerIP:   ip,
		ServerPort: port,
	}
}

func (c *Client) Connect() error {

	conn, err := net.Dial(
		"tcp",
		net.JoinHostPort(
			c.ServerIP,
			strconv.Itoa(c.ServerPort),
		),
	)

	if err != nil {
		return err
	}

	c.conn = conn

	return nil
}

func (c *Client) Login() error {

	var username string
	var password string

	fmt.Print("username:")
	fmt.Scanln(&username)

	fmt.Print("password:")
	fmt.Scanln(&password)
	//序列化
	body, err := json.Marshal(
		LoginRequest{
			Username: username,
			Password: password,
		},
	)

	if err != nil {
		return err
	}

	resp, err := http.Post(
		"http://127.0.0.1:8081/login",
		"application/json",
		bytes.NewBuffer(body),
	)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	var result LoginResponse

	err = json.NewDecoder(resp.Body).
		Decode(&result)

	if err != nil {
		return err
	}

	if result.Code != 0 {
		return fmt.Errorf("%s", result.Msg)
	}

	c.Token = result.Data.Token
	c.Name = result.Data.Nickname

	fmt.Println(
		"login success:",
		c.Name,
	)

	return nil
}

func (c *Client) Auth() {

	msg := fmt.Sprintf(
		"auth|%s\n",
		c.Token,
	)

	c.conn.Write(
		[]byte(msg),
	)
}

func (c *Client) Run() {

	for {

		var msg string

		fmt.Scanln(&msg)

		c.conn.Write(
			[]byte(msg + "\n"),
		)
	}
}

func (c *Client) Receive() {

	io.Copy(
		os.Stdout,
		c.conn,
	)

}

func main() {

	client := NewClient(
		"127.0.0.1",
		8080,
	)

	err := client.Login()

	if err != nil {
		fmt.Println(err)
		return
	}

	err = client.Connect()

	if err != nil {
		fmt.Println(err)
		return
	}

	client.Auth()

	go client.Receive()

	client.Run()
}
