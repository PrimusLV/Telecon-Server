package main

import (
	"fmt"
	"net"
	"telecon/network"
	"telecon/utils"
)

type User struct {
	net.Conn
	name     string
	password string
	Output   chan network.Packet
}

func (user *User) GetName() string {
	return user.name
}
func (user *User) SetName(name string) {
	user.name = name
}

func (user *User) SendPacket(packet network.Packet) {
	user.Output <- packet
}

func NewUser(name string, password string, conn net.Conn) User {
	return User{
		conn,
		name,
		"",
		make(chan network.Packet),
	}
}

func (user *User) Kick(reason string) {
	pk := network.GetPacket(network.PK_DISCONNECT)
	pk.Data[0] = utils.StrToBytes(reason)
	pk.Put(user)
	chatServer.Leave <- *user
}

func (user *User) Disconnect(reason string) {
	fmt.Println("Disconnecting " + user.GetName() + "...")
	chatServer.BroadcastMessage(user.name + " has left the chat: " + reason)
	chatServer.Leave <- *user
}
