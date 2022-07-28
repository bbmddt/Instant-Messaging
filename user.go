package main

import "net"

type User struct {
	Name string
	Addr string
	C    chan string
	conn net.Conn

	server *Server
}

//創建一個用戶的API
func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()

	user := &User{
		Name: userAddr,
		Addr: userAddr,
		C:    make(chan string),
		conn: conn,

		server: server,
	}

	//啟動監聽當前user channel訊息的goroutine
	go user.ListenMessage()

	return user
}

//用戶的上線業務
func (this *User) Online() {

	//用戶上線,將用戶加入到onlineMap中
	this.server.mapLock.Lock()
	this.server.OnlineMap[this.Name] = this
	this.server.mapLock.Unlock()

	//廣播當前用戶上線消息
	this.server.BroadCast(this, "已上線")
}

//用戶的下線業務
func (this *User) Offline() {

	//用戶下線,將用戶從onlineMap中刪除
	this.server.mapLock.Lock()
	delete(this.server.OnlineMap, this.Name)
	this.server.mapLock.Unlock()

	//廣播當前用戶上線消息
	this.server.BroadCast(this, "下線")

}

//用戶處理訊息的業務
func (this *User) DoMessage(msg string) {
	this.server.BroadCast(this, msg)
}

//監聽當前User channel的方法,一有訊息就直接發送給對端客戶端
func (this *User) ListenMessage() {
	for {
		msg := <-this.C

		this.conn.Write([]byte(msg + "\n"))
	}
}
