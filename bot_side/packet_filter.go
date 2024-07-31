package BotSide

import (
	"Eulogist/minecraft/protocol/packet"
	"fmt"
)

// 数据包过滤器过滤来自租赁服的数据包，
// 并根据实际情况由本处的桥接选择是否直接发送回应。
//
// shouldSendCopy 指代该数据包是否需要同步到 ModPC
func (b *BotSide) PacketFilter(pk packet.Packet) (shouldSendCopy bool, err error) {
	if pk == nil {
		return true, nil
	}

	switch p := pk.(type) {
	case *packet.PyRpc:
		shouldSendCopy, err = b.OnPyRpc(p)
		if err != nil {
			err = fmt.Errorf("PacketFilter: %v", err)
		}
	case *packet.StartGame:
		b.gameData, err = b.HandleStartGame(p)
		b.SetShouldDecode(false)
		if err != nil {
			err = fmt.Errorf("PacketFilter: %v", err)
		}
		shouldSendCopy = true
	default:
		shouldSendCopy = true
	}

	return shouldSendCopy, err
}
