package ws_interface

import (
	"Eulogist/core/minecraft/protocol"
	"Eulogist/core/minecraft/protocol/packet"
)

// 维护一个简单的UqHolder, 储存全局玩家的基本信息

var simple_uq_map = make(PlayersUQMap)

func joinUQ(entry protocol.PlayerListEntry) {
	simple_uq_map[entry.Username] = PlayerBasicInfo{
		Name:     entry.Username,
		UUID:     entry.UUID.String(),
		UniqueID: entry.EntityUniqueID,
		XUID:     entry.XUID,
	}
}

func leaveUQ(entry protocol.PlayerListEntry) {
	delete(simple_uq_map, entry.Username)
}

func handlePlayerList(pk packet.PlayerList) {
	// 处理 PlayerList 以更新 UQ 表
	if pk.ActionType == 0 {
		for _, entry := range pk.Entries {
			joinUQ(entry)
		}
	} else {
		for _, entry := range pk.Entries {
			leaveUQ(entry)
		}
	}
	BroadcastMessageToWSClients(Message{
		Type:    WSMSG_UPDATE_UQ,
		Content: simple_uq_map,
	})
}

func GetUQMap() PlayersUQMap {
	return simple_uq_map
}

func SetUQMapPlayer(uqmap PlayersUQMap, playerobj PlayerBasicInfo) {
	uqmap[playerobj.Name] = playerobj
}

func GetPlayerNameByUniqueID(uqid int64) (string, bool) {
	for k, v := range GetUQMap() {
		if v.UniqueID == uqid {
			return k, true
		}
	}
	return "", false
}
