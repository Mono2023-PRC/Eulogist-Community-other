package mc_server

import (
	"Eulogist/core/minecraft/protocol/packet"
	RaknetConnection "Eulogist/core/raknet"
	"Eulogist/core/tools/py_rpc"
	"fmt"
)

/*
数据包过滤器过滤来自租赁服的多个数据包，
然后并将过滤后的多个数据包抄送至客户端。

如果需要，
将根据实际情况由本处的桥接直接发送回应。

writeSinglePacketToClient 指代
用于向客户端抄送数据包的函数。

syncFunc 用于将数据同步到 Minecraft 客户端，
它会在每个数据包被过滤处理后执行一次。

返回的 []error 是一个列表，
分别对应 packets 中每一个数据包的处理成功情况
*/
func (m *MinecraftServer) FiltePacketsAndSendCopy(
	packets []RaknetConnection.MinecraftPacket,
	writePacketsToClient func(packets []RaknetConnection.MinecraftPacket),
	syncFunc func() error,
) []error {
	// 初始化
	sendCopy := make([]RaknetConnection.MinecraftPacket, 0)
	doNotSendCopy := make([]bool, len(packets))
	errResults := make([]error, len(packets))
	// 处理每个数据包
	for index, minecraftPacket := range packets {
		// 如果传入的数据包为空
		if minecraftPacket.Packet == nil {
			continue
		}
		// 根据数据包的类型进行不同的处理
		switch pk := minecraftPacket.Packet.(type) {
		case *packet.PyRpc:
			doNotSendCopy[index], errResults[index] = m.OnPyRpc(pk)
			if err := errResults[index]; err != nil {
				errResults[index] = fmt.Errorf("FiltePacketsAndSendCopy: %v", err)
			}
		case *packet.StartGame:
			// 预处理
			m.entityUniqueID = m.HandleStartGame(pk)
			playerSkin := m.GetPlayerSkin()
			// 发送简要身份证明
			m.WriteSinglePacket(RaknetConnection.MinecraftPacket{
				Packet: &packet.NeteaseJson{
					Data: []byte(
						fmt.Sprintf(`{"eventName":"LOGIN_UID","resid":"","uid":"%s"}`,
							m.fbClient.ClientInfo.Uid,
						),
					),
				},
			})
			// 皮肤特效处理
			if playerSkin == nil {
				m.WriteSinglePacket(RaknetConnection.MinecraftPacket{
					Packet: &packet.PyRpc{
						Value:         py_rpc.Marshal(&py_rpc.SyncUsingMod{}),
						OperationType: packet.PyRpcOperationTypeSend,
					},
				})
			} else {
				m.WriteSinglePacket(RaknetConnection.MinecraftPacket{
					Packet: &packet.PyRpc{
						Value: py_rpc.Marshal(&py_rpc.SyncUsingMod{
							[]any{},
							playerSkin.SkinUUID,
							playerSkin.SkinItemID,
							true,
							map[string]any{},
						}),
						OperationType: packet.PyRpcOperationTypeSend,
					},
				})
			}
		case *packet.UpdatePlayerGameType:
			if pk.PlayerUniqueID == m.entityUniqueID {
				// 如果玩家的唯一 ID 与数据包中记录的值匹配，
				// 则向客户端发送 SetPlayerGameType 数据包，
				// 并放弃当前数据包的发送，
				// 以确保 Minecraft 客户端可以正常同步游戏模式更改。
				// 否则，按原样抄送当前数据包
				writePacketsToClient([]RaknetConnection.MinecraftPacket{
					{
						Packet: &packet.SetPlayerGameType{GameType: pk.GameType},
					},
				})
				doNotSendCopy[index] = true
			}
		default:
			// 默认情况下，
			// 我们需要将数据包同步到客户端
		}
		// 同步数据到 Minecraft 客户端
		if err := syncFunc(); err != nil {
			errResults[index] = fmt.Errorf("FiltePacketsAndSendCopy: %v", err)
		}
	}
	// 抄送数据包
	for index, pk := range packets {
		if doNotSendCopy[index] {
			continue
		}
		sendCopy = append(sendCopy, pk)
	}
	writePacketsToClient(sendCopy)
	// 返回值
	return errResults
}
