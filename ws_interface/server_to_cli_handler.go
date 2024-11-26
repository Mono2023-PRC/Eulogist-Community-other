package ws_interface

// 处理由赞颂者发送到 WebSocket Client 的所有消息

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
var pool_backup = packet.NewServerPool()

var botDatasReady bool
var botName string
var botUniqueID int64
var botUUID uuid.UUID
var botRuntimeID uint64

// 将 json array 序列转换为 []uint32 序列。
// 由于 json 数组中的 int 似乎被认为是 float64,
// 因此需要进一步的转换为 uint32
func convertToIntArr(a any) ([]uint32, bool) {
	newArr := []uint32{}
	arr, ok := a.([]any)
	if !ok {
		return nil, false
	}
	for _, e := range arr {
		e1, ok := e.(float64)
		if !ok {
			return nil, false
		}
		newArr = append(newArr, uint32(e1))
	}
	return newArr, true
}

// 获取数据包的 ID
func getPacketID(pk_bytes []byte) (pkID uint32) {
	protocol.Varuint32(bytes.NewReader(pk_bytes), &pkID)
	return
}

// 将数据包的字节流转换为数据包结构
func bytesToPacket(conn *raknet_wrapper.Raknet, bs []byte) (uint32, packet.Packet, bool) {
	var pk1 packet.Packet
	buffer := bytes.NewBuffer(bs)
	reader := protocol.NewReader(buffer, conn.GetShieldID().Load(), false)
	packetHeader := packet.Header{}
	packetHeader.Read(buffer)
	packetFunc := packet.ListAllPackets()[packetHeader.PacketID]
	if packetFunc == nil {
		return packetHeader.PacketID, nil, false
	}
	pk1 = packetFunc()
	pkID := pk1.ID()
	pk1.Marshal(reader)
	return pkID, pk1, true
}

// 处理第一次获取到的 StartGame 数据包
// 以获取到当前赞颂者所使用的用户的 EntityRuntimeID 以及 EntityUniqueID
func handleStartGame(pk packet.StartGame) {
	botRuntimeID = pk.EntityRuntimeID
	botUniqueID = pk.EntityUniqueID
}

// 处理第一次获取到的 PlayerList 数据包
// 以获取到当前的玩家列表信息
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

// 处理 UpdateAbilities 数据包
// 以更新玩家能力
func handleAbilitySet(pk packet.UpdateAbilities) {
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
	BroadcastMessageToWSClients(Message{
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

// 由客户端发往服务端的特定数据包是否需要监听
func packetIDClientNeedListen(id uint32) bool {
	_, ok := client_to_server_listen_packets[id]
	return ok
}

// 由服务端发往客户端的特定数据包是否需要监听
func packetIDServerNeedListen(id uint32) bool {
	_, ok := server_to_client_listen_packets[id]
	return ok
}

// 过滤掉被拦截的数据包
func FilterBlockingPackets(
	pks []raknet_wrapper.MinecraftPacket,
	filterFunc func(uint32) bool,
) []raknet_wrapper.MinecraftPacket {
	new_pks := []raknet_wrapper.MinecraftPacket{}
	for _, pk := range pks {
		if !filterFunc(getPacketID(pk.Bytes)) {
			new_pks = append(new_pks, pk)
		}
	}
	return new_pks
}

// 判断 租赁服 -> Minecraft客户端 的数据包是否应该被 ws_interface 处理
// 并转发至 WebSocket 客户端
func HandleServerPacketsToWS(server *Server.MinecraftServer, pks []raknet_wrapper.MinecraftPacket) {
	for _, m_pk := range pks {
		pkID := getPacketID(m_pk.Bytes)
		if packetIDServerNeedListen(pkID) {
			_, pk, ok := bytesToPacket(server.Conn, m_pk.Bytes)
			if !ok {
				pterm.Error.Println("无法解析数据包:", pkID)
				return
			}
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
				handleAbilitySet(*pk1)
			}
			BroadcastMessageToWSClients(Message{
				Type: WSMSG_SERVER_PACKET,
				Content: map[string]any{
					"ID":      pkID,
					"Content": pk,
				},
			})
		}
	}
}

// 判断 Minecraft客户端 -> 租赁服 的数据包是否应该被 ws_interface 处理
// 并转发至 WebSocket 客户端
func HandleClientPacketsToWS(client *Client.MinecraftClient, pks []raknet_wrapper.MinecraftPacket) {
	for _, m_pk := range pks {
		pkID := getPacketID(m_pk.Bytes)
		if packetIDClientNeedListen(pkID) {
			_, pk, ok := bytesToPacket(client.Conn, m_pk.Bytes)
			if !ok {
				pterm.Error.Println("无法解析数据包:", pkID)
				return
			}
			BroadcastMessageToWSClients(Message{
				Type: WSMSG_CLIENT_PACKET,
				Content: map[string]any{
					"ID":      pkID,
					"Content": pk,
				},
			})
		}
	}
}

// 来自 Minecraft 客户端的数据包的 ID 是否在需要阻拦的数据包 ID 表里
func InClientPacketsNeedBlocking(pkID uint32) bool {
	_, ok := client_to_server_block_packets[pkID]
	return ok
}

// 来自租赁服的数据包的 ID 是否在需要阻拦的数据包 ID 表里
func InServerPacketsNeedBlocking(pkID uint32) bool {
	_, ok := server_to_client_block_packets[pkID]
	return ok
}
