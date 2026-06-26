package main

import "IM-system/server"
func main(){
	s := server.NewServer("127.0.0.1",8080)//创建服务器
	s.Start()
}