package toonnel

import (
	"fmt"
	"net"
	"strings"
	"testing"
	"time"
)

func TestRemoteConn_listenToSocket(t *testing.T) {
	var port string
	rc := &remoteConn{inChan: make(chan Message), connOn: true}

	go mockServer(&port)
	for port == "" {
	}
	go clientSocketListenToSocket(port, rc)
	t.Run("msg recv", recvMsg(rc))
}

func mockServer(port *string) {
	mockSocket, _ := net.Listen("tcp", ":0")
	*port = fmt.Sprintf(":%s", strings.Split(mockSocket.Addr().String(), "]:")[1])
	conn, _ := mockSocket.Accept()
	_ = writeFromSocket(StringMessage("test"), conn)
	_ = mockSocket.Close()
}

func clientSocketListenToSocket(port string, rc *remoteConn) {
	conn, _ := net.Dial("tcp", port)
	rc.inConn = conn
	rc.listenToSocket()
}

func recvMsg(rc *remoteConn) func(t *testing.T) {
	return func(t *testing.T) {
		select {
		case <-rc.inChan:
		case <-time.After(1 * time.Second):
			t.Errorf("expected incoming mssg from rc.inChan, got none")
		}
	}
}

func TestRemoteConn_newOutSocket(t *testing.T) {
	var port string
	go mockServer(&port)
	for port == "" {
	}

	rc := &remoteConn{remoteHost: fmt.Sprintf("127.0.0.1%s", port)}
	t.Run("newOutSocket ok", func(t *testing.T) {
		if err := rc.newOutSocket(); err != nil {
			t.Errorf("unexpected error, %s", err.Error())
		}
		if rc.outConn == nil {
			t.Errorf("expected rc.outConn != nil, got nil")
		}
	})
	t.Run("newOutSocket err", func(t *testing.T) {
		if err := rc.newOutSocket(); err == nil {
			t.Error("expected error, got nothing")
		}
	})
}

func TestManager(t *testing.T) {
	t.Run("not started", func(t *testing.T) {
		if _, err := Manager("127.0.0.1:0"); err == nil {
			t.Error("mwNotStarted error expected, got nothing")
		}
	})

	// todo finish tests
	//_ = Start(0)
	//port := fmt.Sprintf(":%s", strings.Split(socketServer.Addr().String(), "]:")[1])
}
