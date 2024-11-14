package ws_interface

import (
	"Eulogist/core/minecraft/protocol"
	"Eulogist/core/minecraft/protocol/packet"
	raknet_wrapper "Eulogist/core/raknet/wrapper"
	Server "Eulogist/proxy/mc_server"
	"bytes"
	"encoding/json"

	"github.com/pterm/pterm"
)

var pool = packet.NewClientPool()
var client_to_server_listen_packets = []uint32{}
var server_to_client_listen_packets = []uint32{}

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

func bytesToPacket(server *Server.MinecraftServer, bs []byte) (uint32, string) {
	var pk1 packet.Packet
	buffer := bytes.NewBuffer(bs)
	reader := protocol.NewReader(buffer, server.Conn.GetShieldID().Load(), false)
	// 获取数据包头和数据包处理函数
	packetHeader := packet.Header{}
	packetHeader.Read(buffer)
	packetFunc := packet.ListAllPackets()[packetHeader.PacketID]
	pk1 = packetFunc()
	pkID := pk1.ID()
	pk1.Marshal(reader)
	strbytes, _ := json.Marshal(pk1)
	strbytes_blk := string(strbytes)
	return pkID, strbytes_blk
}

func BroadcastPacketsToWS(server *Server.MinecraftServer, pks []raknet_wrapper.MinecraftPacket) {
	for _, m_pk := range pks {
		pkID, pk_str := bytesToPacket(server, m_pk.Bytes)
		if packetIDServerNeedListen(pkID) {
			BroadcastMessageToWS(Message{
				Type: "MCPacket",
				Content: PacketMessage{
					ID:      pkID,
					Content: pk_str,
				},
			})
		}
	}
}

func HandleWSClientMessages(
	writePacketsToServer func(packets []raknet_wrapper.MinecraftPacket),
) {
	for {
		msg := <-receiver
		if msg.Type == "MCPacket" {
			pkmsg, ok := msg.Content.(PacketMessage)
			pk := pool[pkmsg.ID]()
			if ok {
				err := json.Unmarshal([]byte(pkmsg.Content), &pk)
				if err != nil {
					pterm.Error.Println("数据包解析错误:", err)
					continue
				}
				pk1 := raknet_wrapper.MinecraftPacket{
					Packet: pk,
				}
				writePacketsToServer([]raknet_wrapper.MinecraftPacket{pk1})
			}
		} else if msg.Type == "SetListenPackets" {
			pkmsg, ok := msg.Content.(SetListenPacketsMessage)
			if !ok {
				pterm.Error.Println("无法识别监听数据包请求")
				continue
			}
			setServerToClientListenPackets(pkmsg.PacketIDs)
		}
	}
}
