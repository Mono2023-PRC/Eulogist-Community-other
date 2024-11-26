package ws_interface

import (
	"github.com/gorilla/websocket"
)

// WebSocket 通信部分

// 基本消息类型
const (
	WSMSG_CLIENT_PACKET               = "ClientMCPacket"           // [赞颂者->WSCli] 来自 Minecraft 客户端的数据包
	WSMSG_SERVER_PACKET               = "ServerMCPacket"           // [赞颂者->WSCli] 来自租赁服服务端的数据包
	WSMSG_SET_BOT_BASIC_INFO          = "SetBotBasicInfo"          // [赞颂者->WSCli] 设置玩家基本信息 (RuntimeID 等)
	WSMSG_SET_SERVER_LISTEN_PACKETS   = "SetServerListenPackets"   // [WSCli->赞颂者] 设置需要监听的来自服务器的数据包
	WSMSG_SET_CLIENT_LISTEN_PACKETS   = "SetClientListenPackets"   // [WSCli->赞颂者] 设置需要监听的来自客户端的数据包
	WSMSG_SET_BLOCKING_SERVER_PACKETS = "SetBlockingServerPackets" // [WSCli->赞颂者] 设置需要拦截的来自服务器的数据包
	WSMSG_SET_BLOCKING_CLIENT_PACKETS = "SetBlockingClientPackets" // [WSCli->赞颂者] 设置需要拦截的来自客户端的数据包
	WSMSG_UPDATE_UQ                   = "UpdateUQ"                 // [赞颂者->WSCli] 更新客户端玩家数据
	WSMSG_UPDATE_ABILITIES            = "UpdateAbilities"          // [赞颂者->WSCli] 更新客户端玩家能力数据
)

// 额外的特殊消息类型
const (
	WSMSG_BreakBlock = "BreakBlock" // [WSCli->赞颂者] 请求挖掘方块
)

type WS_Client struct {
	// 来自客户端的连接
	conn *websocket.Conn
	// 客户端通信是否已就绪。
	// 如果未就绪, 赞颂者会尝试向客户端更新机器人基本信息以及 UQHolder 信息。
	ready bool
}

type Message struct {
	// 消息类型
	Type string `json:"type"`
	// 消息正文
	Content any `json:"content"`
}

// UQ 部分

// 赞颂者所操控的玩家的基本信息。
type BotBasicData struct {
	Name           string `json:"bot_name"`
	UUID           string `json:"uuid"`
	EntityUniqueID int64  `json:"bot_entity_unique_id"`
	RuntimeID      uint64 `json:"bot_runtime_id"`
}

// 赞颂者所在租赁服的玩家的基本信息。
type PlayerBasicInfo struct {
	Name      string `json:"name"`
	UUID      string `json:"uuid"`
	XUID      string `json:"xuid"`
	UniqueID  int64  `json:"uniqueID"`
	Abilities any    `json:"abilities"`
}

type PlayersUQMap map[string]PlayerBasicInfo
