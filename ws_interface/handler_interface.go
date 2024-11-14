package ws_interface

import (
	"Eulogist/core/minecraft/protocol"
	"Eulogist/core/minecraft/protocol/packet"
	raknet_wrapper "Eulogist/core/raknet/wrapper"
	Client "Eulogist/proxy/mc_client"
	Server "Eulogist/proxy/mc_server"
	"bytes"

	"github.com/google/uuid"
	"github.com/pterm/pterm"
)

var pool = packet.NewClientPool()
var client_to_server_listen_packets = []uint32{}
var server_to_client_listen_packets = []uint32{}

var botDatasReady bool
var botName string
var botUniqueID int64
var botUUID uuid.UUID
var botRuntimeID uint64

func convertToIntArr(a any) ([]uint32, bool) {
	newArr := []uint32{}
	arr, ok := a.([]any)
	if !ok {
		//print("trial 1 failed\n")
		return nil, false
	}
	for _, e := range arr {
		e1, ok := e.(float64)
		if !ok {
			//print("trial 2 failed on ", i, "=", e, "\n")
			return nil, false
		}
		newArr = append(newArr, uint32(e1))
	}
	return newArr, true
}

func handleStartGame(pk packet.StartGame) {
	botRuntimeID = pk.EntityRuntimeID
	botUniqueID = pk.EntityUniqueID
}

func handleFirstPlayerList(pk packet.PlayerList) {
	for _, entry := range pk.Entries {
		if entry.EntityUniqueID == botUniqueID {
			botName = entry.Username
			botUUID = entry.UUID
			botDatasReady = true
			handoutBotBasicInfo()
		}
	}
}

func HandleAbilitySet(pk packet.UpdateAbilities) {
	playername, found := GetPlayerNameByUniqueID(pk.AbilityData.EntityUniqueID)
	if found {
		uqmap := GetUQMap()
		playerobj := uqmap[playername]
		playerobj.SetAbilities(pk.AbilityData)
		SetUQMapPlayer(uqmap, playerobj)
	} else {
		pterm.Error.Println("未找到", pk.AbilityData.EntityUniqueID, "所对应的玩家")
		return
	}
	BroadcastMessageToWS(Message{
		Type:    WSMSG_UPDATE_UQ,
		Content: simple_uq_map,
	})
}

func getBotBasicInfo() BotBasicData {
	return BotBasicData{
		Name:           botName,
		UUID:           botUUID.String(),
		EntityUniqueID: botUniqueID,
		RuntimeID:      botRuntimeID,
	}
}

// 设置要监听的由 Minecraft 客户端发往服务端的数据包
func setClientToServerListenPackets(pk_ids []uint32) {
	client_to_server_listen_packets = pk_ids
}

// 设置要监听的由服务端发往 Minecraft 客户端的数据包
func setServerToClientListenPackets(pk_ids []uint32) {
	server_to_client_listen_packets = pk_ids
}

// 由客户端发往服务端的特定数据包是否需要监听
func packetIDClientNeedListen(id uint32) bool {
	for _, pkid := range client_to_server_listen_packets {
		if pkid == id {
			return true
		}
	}
	return false
}

// 由服务端发往客户端的特定数据包是否需要监听
func packetIDServerNeedListen(id uint32) bool {
	for _, pkid := range server_to_client_listen_packets {
		if pkid == id {
			return true
		}
	}
	return false
}

func bytesToPacket(conn *raknet_wrapper.Raknet, bs []byte) (uint32, packet.Packet) {
	var pk1 packet.Packet
	buffer := bytes.NewBuffer(bs)
	reader := protocol.NewReader(buffer, conn.GetShieldID().Load(), false)
	packetHeader := packet.Header{}
	packetHeader.Read(buffer)
	packetFunc := packet.ListAllPackets()[packetHeader.PacketID]
	pk1 = packetFunc()
	pkID := pk1.ID()
	pk1.Marshal(reader)
	return pkID, pk1
}

func HandleServerPacketsToWS(server *Server.MinecraftServer, pks []raknet_wrapper.MinecraftPacket) {
	for _, m_pk := range pks {
		pkID, pk := bytesToPacket(server.Conn, m_pk.Bytes)
		if !botDatasReady {
			if pk1, ok := pk.(*packet.StartGame); ok {
				handleStartGame(*pk1)
			} else if pk1, ok := pk.(*packet.PlayerList); ok {
				handleFirstPlayerList(*pk1)
			}
		}
		if pk1, ok := pk.(*packet.PlayerList); ok {
			handlePlayerList(*pk1)
		}
		if pk1, ok := pk.(*packet.UpdateAbilities); ok {
			HandleAbilitySet(*pk1)
		}
		if packetIDServerNeedListen(pkID) {
			BroadcastMessageToWS(Message{
				Type: "ServerMCPacket",
				Content: map[string]any{
					"ID":      pkID,
					"Content": pk,
				},
			})
		}
	}
}

func HandleClientPacketsToWS(client *Client.MinecraftClient, pks []raknet_wrapper.MinecraftPacket) {
	for _, m_pk := range pks {
		pkID, pk := bytesToPacket(client.Conn, m_pk.Bytes)
		if packetIDClientNeedListen(pkID) {
			BroadcastMessageToWS(Message{
				Type: WSMSG_CLIENT_PACKET,
				Content: map[string]any{
					"ID":      pkID,
					"Content": pk,
				},
			})
		}
	}
}
