package mc_server

import (
	"Eulogist/core/fb_auth/py_rpc"
	"Eulogist/core/minecraft/protocol/packet"
	RaknetConnection "Eulogist/core/raknet"
	"encoding/json"
	"fmt"
)

// ...
func (m *MinecraftServer) OnPyRpc(p *packet.PyRpc) (shouldSendCopy bool, err error) {
	if p.Value == nil {
		return true, nil
	}
	// prepare
	content, err := py_rpc.Unmarshal(p.Value)
	if err != nil {
		return true, fmt.Errorf("OnPyRpc: %v", err)
	}
	// unmarshal
	switch c := content.(type) {
	case *py_rpc.StartType:
		c.Content = m.fbClient.TransferData(c.Content)
		c.Type = py_rpc.StartTypeResponse
		err = m.WritePacket(
			RaknetConnection.MinecraftPacket{
				Packet: &packet.PyRpc{
					Value:         py_rpc.Marshal(c),
					OperationType: packet.PyRpcOperationTypeSend,
				},
			}, false,
		)
		if err != nil {
			return false, fmt.Errorf("OnPyRpc: %v", err)
		}
		// get data and send packet
	case *py_rpc.GetMCPCheckNum:
		if m.getCheckNumEverPassed {
			break
		}
		// if the challenges has been down,
		// we do NOTHING
		arg, _ := json.Marshal([]any{
			c.FirstArg,
			c.SecondArg.Arg,
			m.entityUniqueID,
		})
		ret := m.fbClient.TransferCheckNum(string(arg))
		// create request to the auth server and get response
		ret_p := []any{}
		json.Unmarshal([]byte(ret), &ret_p)
		if len(ret_p) > 7 {
			ret6, ok := ret_p[6].(float64)
			if ok {
				ret_p[6] = int64(ret6)
			}
		}
		// unmarshal response and adjust the data included
		err = m.WritePacket(
			RaknetConnection.MinecraftPacket{
				Packet: &packet.PyRpc{
					Value:         py_rpc.Marshal(&py_rpc.SetMCPCheckNum{ret_p}),
					OperationType: packet.PyRpcOperationTypeSend,
				},
			}, false,
		)
		if err != nil {
			return false, fmt.Errorf("OnPyRpc: %v", err)
		}
		m.getCheckNumEverPassed = true
		// send packet and mark this challenges was finished
	default:
		return true, nil
	}
	// do some actions for some specific PyRpc packets
	return false, nil
	// return
}
