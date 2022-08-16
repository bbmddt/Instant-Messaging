package main

import (
	"flag"
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

var serverIP string
var serverPort int

// 命令行參數定義
// -ip 127.0.0.1 -port 8888
func init() {
	flag.StringVar(&serverIP, "ip", "127.0.0.1", "設置伺服器IP位址 (預設:127.0.0.1)")
	flag.IntVar(&serverPort, "port", 8888, "設置伺服器連接埠 (預設:8888)")
}

func main() {
	//命令行解析
	flag.Parse()

	client := NewClient(serverIP, serverPort)
	if client == nil {
		fmt.Println("連接伺服器失敗...")
		return
	}

	fmt.Println("伺服器已連接成功!")

	//啟動client的業務
	select {}
}
