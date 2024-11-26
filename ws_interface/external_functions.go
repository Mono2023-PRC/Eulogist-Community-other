package ws_interface

import (
	"Eulogist/core/minecraft/protocol"
	"Eulogist/core/minecraft/protocol/packet"
	raknet_wrapper "Eulogist/core/raknet/wrapper"
	"encoding/json"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/pterm/pterm"
)

type BreakBlockRequest struct {
	X          int32                 `json:"x"`
	Y          int32                 `json:"y"`
	Z          int32                 `json:"z"`
	HotbarSlot int32                 `json:"hotbarSlot"`
	ItemInfo   protocol.ItemInstance `json:"ItemInfo"`
}

func HandleBreakBlock(
	content map[string]any,
	writePacketsToServer func(packets []raknet_wrapper.MinecraftPacket),
) {
	var req BreakBlockRequest
	jsonTmp, err := json.Marshal(content)
	if err != nil {
		pterm.Error.Println("提供了无效 json 项")
		return
	}
	if err := json.Unmarshal(jsonTmp, &req); err != nil {
		pterm.Error.Println("无效物品实例 json:", err)
		return
	}
	pk_me := makeBreakBlockPacket(req)
	pk := raknet_wrapper.MinecraftPacket{
		Packet: pk_me,
	}
	writePacketsToServer([]raknet_wrapper.MinecraftPacket{pk})

}

func makeBreakBlockPacket(
	req BreakBlockRequest,
) *packet.InventoryTransaction {
	return &packet.InventoryTransaction{
		LegacyRequestID:    0,
		LegacySetItemSlots: []protocol.LegacySetItemSlot(nil),
		Actions:            []protocol.InventoryAction{},
		TransactionData: &protocol.UseItemTransactionData{
			LegacyRequestID:    0,
			LegacySetItemSlots: []protocol.LegacySetItemSlot(nil),
			Actions:            []protocol.InventoryAction(nil),
			ActionType:         protocol.UseItemActionBreakBlock,
			BlockPosition:      [3]int32{req.X, req.Y, req.Z},
			HotBarSlot:         req.HotbarSlot,
			HeldItem:           req.ItemInfo,
			Position:           mgl32.Vec3{},
			BlockRuntimeID:     11365,
		},
	}
}
