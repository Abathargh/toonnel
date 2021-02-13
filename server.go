package toonnel

import (
	"errors"
	"fmt"
	"net"
)

const (
	maxRetryErr = 5
)

var (
	emptyChannelList = errors.New("server has no channels")
	invalidTCPPort   = errors.New("invalid tcp port")
)

// Internal support struct used to map a channel with its
// respective callback function
type tChannel struct {
	channel  chan Message
	callback func(msg Message) Message
}

// The TServer interface describes a generic Toonnel Server
// By default a TCP Server is implemented but this could work with other
// protocols
type TServer interface {
	Start() error
	Listen() error
	Close() error
}

// TOptions is used to build a specific option mapping for a new client
type TOptions struct {
	port     uint
	channels map[string]*tChannel
}

// TCP based Toonnel Server
type TCPServer struct {
	opts         *TOptions
	statusOpen   bool
	serverSocket net.Listener
}

// Adds a port to the option mapping
func (opts *TOptions) AddPort(port uint) {
	opts.port = port
}

// Adds a channel identified by a string id and its respective callback function
// onto an option mapping
func (opts *TOptions) AddChannel(channelId string, callback func(msg Message) Message) {
	if opts.channels == nil {
		opts.channels = make(map[string]*tChannel)
	}
	opts.channels[channelId] = &tChannel{
		channel:  make(chan Message),
		callback: callback,
	}
}

func NewTServer(opts *TOptions) TServer {
	ts := &TCPServer{}
	ts.opts = opts
	return ts
}

func (serv *TCPServer) Start() error {
	if serv.opts.channels == nil || len(serv.opts.channels) == 0 {
		return emptyChannelList
	}
	if serv.opts.port > 65535 {
		return invalidTCPPort
	}
	sPort := fmt.Sprintf(":%d", serv.opts.port)
	serverSocket, err := net.Listen("tcp", sPort)
	if err != nil {
		return err
	}
	serv.serverSocket = serverSocket
	serv.statusOpen = true
	return nil
}

func (serv *TCPServer) handleIncomingConnection(conn net.Conn) {
	var msg Message
	if err := readFromSocket(&msg, conn); err != nil {
		// TODO handle error via log?
		return
	}
	tChan, tcExists := serv.opts.channels[msg.ChannelName]
	if !tcExists {
		currTry := 0
		errMsg := Message{Type: TypeError, Direction: DirectionDOWN, Content: "chan does not exist"}
		writeErr := writeFromSocket(errMsg, conn)
		for writeErr != nil && currTry < maxRetries {
			writeErr = writeFromSocket(errMsg, conn)
			currTry++
		}
	}
	go func() {
		for {

		}
	}()
}

func (serv *TCPServer) Listen() {
	for {
		conn, _ := serv.serverSocket.Accept()
		go serv.handleIncomingConnection(conn)
	}
}

func (serv *TCPServer) Close() error {
	if err := serv.serverSocket.Close(); err != nil {
		return err
	}
	serv.statusOpen = true
	return nil
}
