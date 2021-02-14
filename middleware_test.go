// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package toonnel

import (
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"testing"
)

func TestReadFromSocket(t *testing.T) {
	var port string
	go mockReadServer(&port)
	for port == "" {
	}

	t.Run("read ok", okRead(port))
	t.Run("read conn error", connErrRead(port))
	t.Run("read fmt error", fmtErrRead(port))
	t.Run("read invalid msg error", invalidErrRead(port))
}

func mockReadServer(port *string) {
	mockSocket, _ := net.Listen("tcp", ":0")
	*port = fmt.Sprintf(":%s", strings.Split(mockSocket.Addr().String(), "]:")[1])

	conn, _ := mockSocket.Accept()
	msg, _ := json.Marshal(StringMessage("test"))
	_, _ = fmt.Fprintf(conn, "%s\n", msg)

	conn, _ = mockSocket.Accept()
	_ = conn.Close()

	conn, _ = mockSocket.Accept()
	msgFmt := "invalid"
	_, _ = fmt.Fprintf(conn, "%s\n", msgFmt)

	conn, _ = mockSocket.Accept()
	iMsg := StringMessage("invalid")
	iMsg.Type = TypeUndefined
	msg, _ = json.Marshal(iMsg)
	_, _ = fmt.Fprintf(conn, "%s\n", msg)
	_ = mockSocket.Close()
}

func okRead(port string) func(t *testing.T) {
	return func(t *testing.T) {
		conn, _ := net.Dial("tcp", port)
		var msg Message
		err := readFromSocket(&msg, conn)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}

		if msg.Content != "test" {
			t.Errorf("expected message.Content = test, got %s", msg.Content)
		}
	}
}

func connErrRead(port string) func(t *testing.T) {
	return func(t *testing.T) {
		conn, _ := net.Dial("tcp", port)
		var msg Message
		err := readFromSocket(&msg, conn)
		if err == nil || err != incomingError {
			t.Error("expected incomingError")
		}
	}
}

func fmtErrRead(port string) func(t *testing.T) {
	return func(t *testing.T) {
		conn, _ := net.Dial("tcp", port)
		var msg Message
		err := readFromSocket(&msg, conn)
		if err == nil || err != formatError {
			t.Error("expected formatError")
		}
	}
}

func invalidErrRead(port string) func(t *testing.T) {
	return func(t *testing.T) {
		conn, _ := net.Dial("tcp", port)
		var msg Message
		err := readFromSocket(&msg, conn)
		if err == nil || err != invalidMessage {
			t.Error("expected formatError")
		}
	}
}

func TestWriteFromSocket(t *testing.T) {
	var port string
	go mockWriteServer(&port)

	for port == "" {
	}

	t.Run("write ok", okWrite(port))
}

func mockWriteServer(port *string) {
	mockSocket, _ := net.Listen("tcp", ":0")
	*port = fmt.Sprintf(":%s", strings.Split(mockSocket.Addr().String(), "]:")[1])

	conn, _ := mockSocket.Accept()
	var msg Message
	_ = readFromSocket(&msg, conn)

	_ = mockSocket.Close()
}

func okWrite(port string) func(t *testing.T) {
	return func(t *testing.T) {
		conn, _ := net.Dial("tcp", port)
		msg := StringMessage("test")

		if err := writeFromSocket(msg, conn); err != nil {
			t.Errorf("unexpected error, %s", err.Error())
		}
	}
}
