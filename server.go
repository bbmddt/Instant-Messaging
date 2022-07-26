package main

import (
	"fmt"
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

//創建一個server的介面
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

	user := NewUser(conn)

	//用戶上線,將用戶加入到onlineMap中
	this.mapLock.Lock()
	this.OnlineMap[user.Name] = user
	this.mapLock.Unlock()

	//廣播當前用戶上線消息
	this.BroadCast(user, "已上線")

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
