package websocket

import (
	"fmt"
	"net/http"
	"node/database"

	"github.com/gorilla/websocket"
)

var clients = make(map[*websocket.Conn]bool)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// 접속한 웹소켓 클라이언트, clients에 저장
func WebsocketHandler(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer ws.Close()

	clients[ws] = true
	fmt.Println("클라이언트 접속", ws.RemoteAddr())

	for {
		var msg string
		err := ws.ReadJSON(&msg)
		if err != nil {
			fmt.Println("클라이언트 접속 종료", ws.RemoteAddr())
			delete(clients, ws)
			break
		}
	}
}

// tcpserver가 파일 업로드를 수행한 후, 호출
// websocket 클라이언트 들에게 파일리스트를 전달하는
func SendMessages(files []database.File) {

	for client := range clients {
		err := client.WriteJSON(files)
		if err != nil {
			fmt.Println(err)
			client.Close()
			delete(clients, client)
		}
	}
}
