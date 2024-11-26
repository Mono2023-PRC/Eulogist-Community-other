package ws_interface

import (
	"Eulogist/core/minecraft/protocol"
	"Eulogist/core/minecraft/protocol/packet"
)

// 维护一个简单的UqHolder, 储存全局玩家的基本信息

var simple_uq_map = make(PlayersUQMap)

// 处理 PlayerList 以更新 UQ 表
func handlePlayerList(pk packet.PlayerList) {
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

// 设置 (更新) UQ 表中的玩家实例
func SetUQMapPlayer(uqmap PlayersUQMap, playerobj PlayerBasicInfo) {
	uqmap[playerobj.Name] = playerobj
}

// 从玩家的 Unique ID 获取玩家名
func GetPlayerNameByUniqueID(uqid int64) (string, bool) {
	for k, v := range GetUQMap() {
		if v.UniqueID == uqid {
			return k, true
		}
	}
	return "", false
}

func (pb *PlayerBasicInfo) SetAbilities(abilities protocol.AbilityData) {
	pb.Abilities = abilities
}

// 向 UQ 表添加玩家对象
func joinUQ(entry protocol.PlayerListEntry) {
	simple_uq_map[entry.Username] = PlayerBasicInfo{
		Name:     entry.Username,
		UUID:     entry.UUID.String(),
		UniqueID: entry.EntityUniqueID,
		XUID:     entry.XUID,
	}
}

// 从 UQ 表移除玩家对象
func leaveUQ(entry protocol.PlayerListEntry) {
	delete(simple_uq_map, entry.Username)
}
