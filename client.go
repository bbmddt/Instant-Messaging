package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
)

type Client struct {
	ServerIP   string
	ServerPort int
	Name       string
	conn       net.Conn
	flag       int //當前Client模式
}

func NewClient(serverIP string, serverPort int) *Client {
	//創建Clinet端物件
	client := &Client{
		ServerIP:   serverIP,
		ServerPort: serverPort,
		flag:       -1,
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

// 選項模式
func (c *Client) menu() bool {
	var flag int
	fmt.Println("1 => 公開頻道")
	fmt.Println("2 => 私人訊息")
	fmt.Println("3 => 修改用戶名稱")
	fmt.Println("0 => 離開")

	fmt.Scanln(&flag)
	if flag >= 0 && flag <= 3 {
		c.flag = flag
		return true
	} else {
		fmt.Println("===輸入錯誤! 請重新輸入合法數字===")
		return false
	}

}

// 公開頻道(選項模式1)
func (c *Client) PublicChat() {

	var chatMsg string

	//提示用戶輸入訊息
	fmt.Println("===請輸入訊息內容，exit退出。")
	fmt.Scanln(&chatMsg)

	for chatMsg != "exit" {
		//發送給伺服器

		//內容不為空則發送
		if len(chatMsg) != 0 {
			sendMsg := chatMsg + "\n"
			_, err := c.conn.Write([]byte(sendMsg))
			if err != nil {
				fmt.Println("conn.Write err:", err)
				break
			}
		}

		chatMsg = ""
		fmt.Println("===請輸入訊息內容，exit退出。")
		fmt.Scanln(&chatMsg)
	}
}

// 私人訊息(選項模式2)

//查詢在線用戶
func (c *Client) CheckUsers() {
	sendMsg := "who\n"
	_, err := c.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn.Write err:", err)
		return
	}
}

//私訊處理
func (c *Client) PrivateChat() {

	var remoteName string
	var chatMsg string

	//先查詢
	c.CheckUsers()
	fmt.Println("===請輸入私訊對象[名稱]，exit退出")
	fmt.Scanln(&remoteName)

	for remoteName != "exit" {
		fmt.Println("===請輸入訊息內容，exit退出。")
		fmt.Scanln(&chatMsg)

		for chatMsg != "exit" {

			//內容不為空則發送
			if len(chatMsg) != 0 {
				sendMsg := `to"` + remoteName + `"` + chatMsg + "\n\n"
				_, err := c.conn.Write([]byte(sendMsg))
				if err != nil {
					fmt.Println("conn.Write err:", err)
					break
				}
			}

			chatMsg = ""
			fmt.Println("===請輸入訊息內容，exit退出。")
			fmt.Scanln(&chatMsg)
		}

		c.CheckUsers()
		fmt.Println("===請輸入私訊對象[名稱]，exit退出")
		fmt.Scanln(&remoteName)
	}
}

// 重新命名處理(選項模式3)
func (c *Client) ReName() bool {

	fmt.Println("===請輸入用戶名稱:")
	fmt.Scanln(&c.Name)

	sendMsg := "rename/" + c.Name + "\n"
	_, err := c.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn.Write err:", err)
		return false
	}

	return true
}

//處理server回應的訊息，直接顯示到標準輸出即可
func (c *Client) ServerResp() {
	//一旦c.conn有資料，就直接copy到stdout標準輸出上，永久阻塞保持監聽
	io.Copy(os.Stdout, c.conn)
}

func (c *Client) Run() {
	for c.flag != 0 {
		for c.menu() != true {
		}

		//依照輸入的數字切換不同的業務模式
		switch c.flag {
		case 1:
			//公開頻道
			c.PublicChat()
		case 2:
			//私人訊息
			c.PrivateChat()
		case 3:
			//修改用戶名稱
			c.ReName()
		}
	}
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

	//另外開啟一個goroutine來處理Server回應的訊息
	go client.ServerResp()

	fmt.Println("伺服器已連接成功!")

	//啟動client的業務
	client.Run()
}
