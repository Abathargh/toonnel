package toonnel

import (
	"errors"
	"fmt"
	"net"
)

var (
	emptyChannelList = errors.New("Server has no channel")
)

type TServer struct {
	port         uint
	channels     []chan interface{}
	serverSocket *net.Listener
}

func (serv TServer) Start(port uint) error {
	if serv.channels == nil || len(serv.channels) == 0 {
		return
	}
	sPort := fmt.Sprintf(":%d", port)
	socketServer, err := net.Listen("tcp", sPort)
	if err != nil {
		return err
	}

}
