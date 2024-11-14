package ws_interface

import (
	raknet_wrapper "Eulogist/core/raknet/wrapper"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pterm/pterm"

	"github.com/gorilla/websocket"
)

const ws_port = 10132

var clis = make(map[*WS_Client]bool)
var broadcaster = make(chan Message)
var receiver = make(chan Message)
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// 向所有 WebSocketClient 广播消息
func BroadcastMessageToWS(msg Message) {
	broadcaster <- msg
}

// 处理来自单个 WebSocketClient 的连接
func handleConnections(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		pterm.Error.Println("客户端连接预处理出错:", err)
		return
	}
	defer conn.Close()

	client := &WS_Client{conn: conn, ready: false}
	clis[client] = true

	pterm.Info.Println("客户端", conn.RemoteAddr().String(), "已连接到赞颂者")

	if botDatasReady {
		sendBotBasicIDAndSetClientReady(*client)
	}
	sendUpdateUQ(*client)

	// 读取消息
	for {
		var msg Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			pterm.Warning.Println("客户端", conn.RemoteAddr().String(), "连接中断:", err)
			delete(clis, client)
			break
		}
		// 暂时将所有消息统一处理
		receiver <- msg
	}
}

// 统一处理来自所有 WebSocketClient 的通信消息
func HandleWSClientsMessages(
	writePacketsToServer func(packets []raknet_wrapper.MinecraftPacket),
) {
	for {
		msg := <-receiver
		content, ok := msg.Content.(map[string]interface{})
		if !ok {
			pterm.Error.Println("无效 WS 信息:", msg.Type)
			continue
		}
		switch msg.Type {
		case WSMSG_SERVER_PACKET:
			pkID, ok := content["ID"].(float64)
			if !ok {
				pterm.Error.Println("数据包解析错误: 无效ID:", content["ID"])
				continue
			}
			pk := pool[uint32(pkID)]()
			pkt_str, ok := content["Content"].(string)
			if !ok {
				pterm.Error.Println("数据包解析错误: 无效数据包体:", content["Content"])
				continue
			}
			err := json.Unmarshal([]byte(pkt_str), &pk)
			if err != nil {
				pterm.Error.Println("数据包解析错误: 无效数据包结构:", err)
				continue
			}
			pk1 := raknet_wrapper.MinecraftPacket{
				Packet: pk,
			}
			writePacketsToServer([]raknet_wrapper.MinecraftPacket{pk1})

		case WSMSG_SET_SERVER_LISTEN_PACKETS:
			pkIDs, ok := convertToIntArr(content["PacketsID"])
			if !ok {
				pterm.Error.Println("无法识别监听数据包请求:", content["PacketsID"])
				continue
			}
			// pterm.Info.Println("设置监听服务端的数据包:", pkIDs)
			setServerToClientListenPackets(pkIDs)
		case WSMSG_SET_CLIENT_LISTEN_PACKETS:
			pkIDs, ok := convertToIntArr(content["PacketsID"])
			if !ok {
				pterm.Error.Println("无法识别监听数据包请求:", content["PacketsID"])
				continue
			}
			// pterm.Info.Println("设置监听客户端的数据包:", pkIDs)
			setClientToServerListenPackets(pkIDs)
		default:
			pterm.Warning.Println("无效的消息类型:", msg.Type)
		}
	}
}

func handleBroadcasts() {
	// 处理来自赞颂者内部向所有 WSClient 广播的消息
	for {
		msg := <-broadcaster
		for client := range clis {
			err := client.sendJson(msg)
			if err != nil {
				pterm.Error.Println(err)
				client.conn.Close()
				delete(clis, client)
			}

		}
	}
}

func handoutBotBasicInfo() {
	// 分发赞颂者机器人自身的基本信息, 如玩家名, UQ等
	for cli := range clis {
		if !cli.ready {
			sendBotBasicIDAndSetClientReady(*cli)
		}
	}
}

func sendBotBasicIDAndSetClientReady(cli WS_Client) {
	// 向 WSClient 客户端发送 BotBasicID 并使其得到初始化
	cli.conn.WriteJSON(Message{
		Type:    WSMSG_SET_BOT_BASIC_INFO,
		Content: getBotBasicInfo(),
	})
	cli.Ready()
}

func sendUpdateUQ(cli WS_Client) {
	// 向客户端发送全局玩家 UQ 更新信息
	cli.sendJson(Message{
		Type:    WSMSG_UPDATE_UQ,
		Content: simple_uq_map,
	})
}

func StartWSServer() {
	// 开启赞颂者 WebSocket 接口服务器
	http.HandleFunc("/", handleConnections)
	go handleBroadcasts()

	pterm.Info.Println("赞颂者在", ws_port, "开放 WebSocket 接口")
	pterm.Error.Println(http.ListenAndServe(fmt.Sprintf(":%d", ws_port), nil))
}
