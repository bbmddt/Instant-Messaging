package main

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

type Server struct {
	Ip   string
	Port int

	//線上用戶的列表
	OnlineMap map[string]*User
	mapLock   sync.RWMutex

	//訊息廣播的channel
	Message chan string
}

//創建一個server的物件
func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}

	return server
}

//監聽Message廣播訊息channel的goroutine，一有訊息就發送给全部的線上User
func (this *Server) ListenMessager() {
	for {
		msg := <-this.Message

		//將msg發送給全部的線上User
		this.mapLock.Lock()
		for _, cli := range this.OnlineMap {
			cli.C <- msg
		}
		this.mapLock.Unlock()
	}
}

//廣播訊息的方法
func (this *Server) BroadCast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg

	this.Message <- sendMsg
}

func (this *Server) Handler(conn net.Conn) {
	//當前連接的任務
	//fmt.Println("成功建立連結")

	user := NewUser(conn, this)

	user.Online()

	//監控用戶是否活躍的channel
	isLive := make(chan bool)

	//接收客戶端發送的消息
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)
			if n == 0 {
				user.Offline()
				return
			}

			if err != nil && err != io.EOF {
				fmt.Println("Conn Read err:", err.Error())
				return
			}

			//提取用戶的訊息(去除'\n')
			msg := string(buf[:n-1])

			//用戶針對msg進行訊息處理
			user.DoMessage(msg)

			//用戶發送任意訊息，即表示活躍中
			isLive <- true
		}
	}()

	//當前handler阻塞
	for {
		select {
		case <-isLive:
			//觸發此case即表示用戶活躍中，
			//不執行任何動作並進入下個for循環，以重置下個case的計時器。
		case <-time.After(time.Second * 10):
			//已閒置逾時，將當前的User強制下線

			user.SendMsg("閒置過久...已將您登出!")

			//銷毀踢除用戶的資源
			close(user.C)

			//關閉連線
			conn.Close()

			//退出當前Handler
			return
		}
	}
}

//啟動伺服器介面
func (this *Server) Start() {
	//socket listen
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this.Ip, this.Port))
	if err != nil {
		fmt.Println("net.Listen err:", err)
		return
	}
	//close listen socket
	defer listener.Close()

	//啓動監聽Message的goroutine
	go this.ListenMessager()

	for {
		//accept
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listener accept err:", err)
			continue
		}

		//do handler
		go this.Handler(conn)
	}
}
