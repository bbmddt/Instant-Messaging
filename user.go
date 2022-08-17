package main

import (
	"net"
	"strings"
)

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
func (u *User) Online() {

	//用戶上線,將用戶加入到onlineMap中
	u.server.mapLock.Lock()
	u.server.OnlineMap[u.Name] = u
	u.server.mapLock.Unlock()

	//廣播當前用戶上線消息
	u.server.BroadCast(u, "已上線...")
}

//用戶的下線業務
func (u *User) Offline() {

	//用戶下線,將用戶從onlineMap中刪除
	u.server.mapLock.Lock()
	delete(u.server.OnlineMap, u.Name)
	u.server.mapLock.Unlock()

	//廣播當前用戶下線消息
	u.server.BroadCast(u, "已下線...")

}

//發送訊息給當前User對應的客戶端
func (u *User) SendMsg(msg string) {
	u.conn.Write([]byte(msg))
}

//用戶處理訊息的業務
func (u *User) DoMessage(msg string) {
	//查詢當前線上用戶清單
	if msg == "who" {

		u.server.mapLock.Lock()
		for _, user := range u.server.OnlineMap {
			onlineMsg := "[" + user.Addr + "]" + user.Name + ":" + "上線中...\n"
			u.SendMsg(onlineMsg)
		}
		u.server.mapLock.Unlock()

		//修改用戶名稱
	} else if len(msg) > 7 && msg[:7] == "rename/" {
		//訊息格式: rename/金乘五
		//newName := strings.Split(msg, "/")[1]
		//上為原取newName方法，優化用slice取
		newName := msg[7:]

		//判斷名字是否與當前線上用戶同名
		_, isExist := u.server.OnlineMap[newName]
		if isExist {
			u.SendMsg("已存在相同名稱...\n")
		} else {
			u.server.mapLock.Lock()
			delete(u.server.OnlineMap, u.Name)
			u.server.OnlineMap[newName] = u
			u.server.mapLock.Unlock()

			u.Name = newName
			u.SendMsg("已更新名稱為:" + u.Name + "\n")
		}

		//私訊功能
	} else if len(msg) > 4 && msg[:3] == `to"` {
		//訊息格式: to"鄧只騎"

		//1. 獲取對方的用戶名稱
		remoteName := strings.Split(msg, `"`)[1]
		if remoteName == "" {
			u.SendMsg(`訊息格式錯誤，請使用\to"鄧只騎"hello\格式。` + "\n")
			return
		}

		//2. 根據用戶名稱，得到對方User物件
		remoteUser, ok := u.server.OnlineMap[remoteName]
		if !ok {
			u.SendMsg("查無此用戶名...\n")
			return
		}

		//3. 獲取訊息內容，通過對方的User物件將內容發送過去
		content := strings.Split(msg, `"`)[2]
		if content == "" {
			u.SendMsg("無訊息內容，請重新輸入...\n")
			return
		}

		remoteUser.SendMsg(u.Name + " 私訊說:" + content)

	} else {
		u.server.BroadCast(u, msg)
	}
}

//監聽當前User channel的方法,一有訊息就直接發送給對端客戶端
func (u *User) ListenMessage() {
	for {
		msg := <-u.C

		u.conn.Write([]byte(msg + "\n"))
	}
}
