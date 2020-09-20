// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package toonnel

// TODO future add support for inter process channels without tcp sockets

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"runtime/debug"
	"strings"
	"sync"
)

const bufferSize = 20

var (
	mutex        sync.Mutex
	started      bool
	socketServer net.Listener

	extConnections  map[string]*remoteConn
	managerMappings map[string]*RemoteManager

	// Errors
	alreadyStarted = errors.New("MW is already started")
	mwNotStarted   = errors.New("MW was not started")
	invalidMessage = errors.New("error while trying to process an incoming message: wrong format")
	incomingError  = errors.New("error while trying to read incoming data")
	formatError    = errors.New("data not correctly formatted")
)

type ConnectionError struct {
	remoteHost string
}

// Initializes the MW
func Start(port uint) error {
	if started {
		return alreadyStarted
	}

	fPort := fmt.Sprintf(":%d", port)

	var err error
	socketServer, err = net.Listen("tcp", fPort)
	if err != nil {
		return err
	}

	started = true
	extConnections = make(map[string]*remoteConn)
	managerMappings = make(map[string]*RemoteManager)

	go listen()
	return nil
}

func listen() {
	for {
		inConn, _ := socketServer.Accept()
		go handleNewIncomingConnection(inConn)
	}
}

func readFromSocket(msg *Message, inConn net.Conn) error {
	data, err := bufio.NewReader(inConn).ReadBytes('\n')
	if err != nil {
		return incomingError
	}

	if err := json.Unmarshal(data, &msg); err != nil {
		return formatError
	}

	if !msg.IsValid() {
		return invalidMessage
	}
	// log.Printf("Received: %+v\n", msg)
	return nil
}

func writeFromSocket(msg Message, conn net.Conn) error {
	strMsg, _ := json.Marshal(msg)

	if _, err := fmt.Fprintf(conn, "%s\n", strMsg); err != nil {
		return err
	}
	return nil
}

func handleNewIncomingConnection(inConn net.Conn) {
	var msg Message
	if err := readFromSocket(&msg, inConn); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, string(debug.Stack()))
		return
	}

	mutex.Lock()
	defer mutex.Unlock()

	hostAddr := strings.Split(inConn.RemoteAddr().String(), ":")[0]
	manager, mngExists := managerMappings[hostAddr]

	if !mngExists {
		conn, connExists := extConnections[hostAddr]
		if !connExists {
			extConnections[hostAddr] = &remoteConn{
				inConn:  inConn,
				inChan:  make(chan Message, bufferSize),
				outChan: make(chan Message, bufferSize),
			}
		} else {
			conn.inConn = inConn
		}
	} else {
		manager.remoteConn.inConn = inConn
	}

	msg.Direction = DirectionDOWN
	extConnections[hostAddr].inChan <- msg
}

func removeConn(hostName string) {
	mutex.Lock()
	defer mutex.Unlock()

	delete(extConnections, hostName)
	delete(managerMappings, hostName)
}
