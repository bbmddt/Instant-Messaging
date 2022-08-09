package main

import (
	"fmt"
	"io"
	"net"
	"sync"
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
				fmt.Println("Conn Read err:", err)
				return
			}

			//提取用戶的訊息(去除'\n')
			msg := string(buf[:n-1])

			//用戶針對msg進行訊息處理
			user.DoMessage(msg)
		}
	}()

	//當前handler阻塞
	select {}
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
