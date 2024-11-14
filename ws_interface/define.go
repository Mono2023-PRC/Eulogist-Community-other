package ws_interface

import (
	"Eulogist/core/minecraft/protocol"

	"github.com/gorilla/websocket"
)

// WebSocket 通信部分

const (
	WSMSG_CLIENT_PACKET             = "ClientMCPacket"
	WSMSG_SERVER_PACKET             = "ServerMCPacket"
	WSMSG_SET_BOT_BASIC_INFO        = "SetBotBasicInfo"
	WSMSG_SET_SERVER_LISTEN_PACKETS = "SetServerListenPackets"
	WSMSG_SET_CLIENT_LISTEN_PACKETS = "SetClientListenPackets"
	WSMSG_UPDATE_UQ                 = "UpdateUQ"
	WSMSG_UPDATE_ABILITIES          = "UpdateAbilities"
)

type WS_Client struct {
	conn  *websocket.Conn
	ready bool
}

type Message struct {
	Type    string `json:"type"`
	Content any    `json:"content"`
}

// UQ 部分

type BotBasicData struct {
	Name           string `json:"bot_name"`
	UUID           string `json:"uuid"`
	EntityUniqueID int64  `json:"bot_entity_unique_id"`
	RuntimeID      uint64 `json:"bot_runtime_id"`
}

type PlayerBasicInfo struct {
	Name      string `json:"name"`
	UUID      string `json:"uuid"`
	XUID      string `json:"xuid"`
	UniqueID  int64  `json:"uniqueID"`
	Abilities any    `json:"abilities"`
}

type PlayersUQMap map[string]PlayerBasicInfo

func (pb *PlayerBasicInfo) SetAbilities(abilities protocol.AbilityData) {
	pb.Abilities = abilities
}

func (cli *WS_Client) Ready() {
	cli.ready = true
}

func (cli *WS_Client) sendJson(content interface{}) error {
	if cli.conn != nil {
		return cli.conn.WriteJSON(content)
	} else {
		return nil
	}
}
