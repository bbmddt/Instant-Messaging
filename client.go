package main

import (
	"fmt"
	"net"
)

type Client struct {
	ServerIP   string
	ServerPort int
	Name       string
	conn       net.Conn
}

func NewClient(serverIP string, serverPort int) *Client {
	//創建Clinet端物件
	client := &Client{
		ServerIP:   serverIP,
		ServerPort: serverPort,
	}

	//連接server
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", serverIP, serverPort))
	if err != nil {
		fmt.Println("net.Dial Error:", err)
		return nil
	}

	client.conn = conn

	//返回client物件
	return client
}

func main() {
	client := NewClient("127.0.0.1", 8888)
	if client == nil {
		fmt.Println("連接伺服器失敗...")
		return
	}

	fmt.Println("伺服器已連接成功!")

	//啟動client的業務
	select {}
}
