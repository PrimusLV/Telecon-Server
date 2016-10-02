package main

import (
	"fmt"
	"telecon/logger"
	"telecon/network"
	"telecon/utils"
)

type ChatServer struct {
	users      map[string]User
	Join       chan User
	Leave      chan User
	Connecting chan User
}

func (cs *ChatServer) Run() {
	for {
		select {
		case user := <-cs.Join:
			cs.users[user.RemoteAddr().String()] = user
			cs.BroadcastMessage(fmt.Sprintf("%s has joined the chat", user.GetName()))
		case user := <-cs.Leave:
			delete(cs.users, user.GetName())
			cs.BroadcastMessage(fmt.Sprintf("%s has left the chat", user.GetName()))
		case user := <-cs.Connecting:
			log.Debug("New user (" + user.RemoteAddr().String() + ") connecting...")
		}
	}
}

func (cs *ChatServer) handlePacket(user *User, packet network.Packet) {
	switch packet.GetType() {
	case network.PK_LOGIN:
		user.name = utils.BytesToStr(packet.Data[0])
		user.password = utils.BytesToStr(packet.Data[0])
		cs.Join <- *user
		go user.SendPacket(packet)
	case network.PK_DISCONNECT:
		user.Disconnect(utils.BytesToStr(packet.Data[0]))
	case network.PK_MESSAGE:
		cs.BroadcastMessage(fmt.Sprintf("<%s> %s", user.GetName(), utils.BytesToStr(packet.Data[0])))
	default:
		log.Error("Unhandled packet! Dumping...")
		packet.Dump()
		Stop()
	}
}

func (cs *ChatServer) BroadcastPacket(packet network.Packet) {
	for _, u := range cs.users {
		u.SendPacket(packet)
	}
}

func (cs *ChatServer) BroadcastMessage(message string) {
	pk := network.GetPacket(network.PK_MESSAGE)
	pk.Data[0] = utils.StrToBytes(message)
	cs.BroadcastPacket(*pk)
	logger.Log(logger.BLANK, "", message)
}
