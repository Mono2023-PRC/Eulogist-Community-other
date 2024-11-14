package ws_interface

import (
	"fmt"
	"net/http"

	"github.com/pterm/pterm"

	"github.com/gorilla/websocket"
)

type Client struct {
	conn *websocket.Conn
}

type Message struct {
	Type    string `json:"type"`
	Content any    `json:"Content"`
}

type PacketMessage struct {
	ID      uint32 `json:"id"`
	Content string `json:"content"`
}

type SetListenPacketsMessage struct {
	PacketIDs []uint32 `json:"packet_ids"`
}

const ws_port = 10132

var clis = make(map[*Client]bool)
var broadcaster = make(chan Message)
var receiver = make(chan Message)
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func BroadcastMessageToWS(msg Message) {
	broadcaster <- msg
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		pterm.Error.Println("客户端连接预处理出错:", err)
		return
	}
	defer conn.Close()

	client := &Client{conn: conn}
	clis[client] = true

	pterm.Info.Println("客户端", conn.RemoteAddr().String(), "已连接到赞颂者")

	for {
		var msg Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			pterm.Warning.Println("客户端", conn.RemoteAddr().String(), "连接中断:", err)
			delete(clis, client)
			break
		}

		receiver <- msg
	}
}

func handleBroadcasts() {
	for {
		msg := <-broadcaster

		for client := range clis {
			if client.conn != nil {
				err := client.conn.WriteJSON(msg)
				if err != nil {
					pterm.Error.Println(err)
					client.conn.Close()
					delete(clis, client)
				}
			}
		}
	}
}

func StartWSServer() {
	http.HandleFunc("/", handleConnections)
	go handleBroadcasts()

	pterm.Info.Println("赞颂者在", ws_port, "开放 WebSocket 接口")
	pterm.Error.Println(http.ListenAndServe(fmt.Sprintf(":%d", ws_port), nil))
}
