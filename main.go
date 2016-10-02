package main

import (
	"bytes"
	"flag"
	"net"
	"os"
	"telecon/logger"
	"telecon/network"
)

var log logger.Logger = logger.Logger{}

var chatServer ChatServer

func handleConn(chatServer *ChatServer, conn net.Conn) {
	defer conn.Close()

	user := NewUser("Username", "Password", conn)
	chatServer.Connecting <- user
	defer func() {
		chatServer.Leave <- user
	}()

	var buffer bytes.Buffer
	var current []byte = make([]byte, 1024*5)
	go func() {
		for {
			n, err := conn.Read(current)
			if err != nil {
				if err.Error() == "EOF" {
					user.Disconnect("Disconnected")
					break
				}
			}
			buffer.Write(current[:n])
			packets, rest := network.ReadPackets(buffer.Bytes())
			buffer.Reset()
			buffer.Write(rest)
			for _, p := range packets {
				go chatServer.handlePacket(&user, p)
			}
		}
	}()

	for pk := range user.Output {
		go pk.Put(conn)
	}
}

func main() {
	flag.Parse()
	server, err := net.Listen("tcp", ":9000")
	if err != nil {
		log.Critical(err)
	}
	defer server.Close()

	log.Info("Running on localhost:9000")

	chatServer := &ChatServer{
		users:      make(map[string]User),
		Join:       make(chan User),
		Leave:      make(chan User),
		Connecting: make(chan User),
	}
	go chatServer.Run()

	for {
		conn, err := server.Accept()
		if err != nil {
			log.Critical(err)
		}
		go handleConn(chatServer, conn)
	}
}

func Stop() {
	log.Info("Stopping...")
	log.Info("Stopped.")
	os.Exit(0)
}
