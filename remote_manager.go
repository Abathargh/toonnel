// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package toonnel

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"reflect"
	"strings"
	"sync"
)

const (
	maxRetries = 5
)

var (
	errorPrint           = stdErrPrint
	noSuchChannel        = errors.New("runtime error: no channel with the specified name")
	channelAlreadyExists = errors.New("runtime error: channel with the same name already exists")
)

func stdErrPrint(err error) {
	_, _ = fmt.Fprint(os.Stderr, err.Error())
}

type remoteConn struct {
	remoteHost string
	connOn     bool
	inConn     net.Conn
	outConn    net.Conn
	inChan     chan Message
	outChan    chan Message
}

func (rc *remoteConn) listenToSocket() {
	// todo fix with blocking maybe next func too
	// todo add while manager not ready
	for rc.inConn == nil {
	}

	for rc.connOn {
		var msg Message
		if err := readFromSocket(&msg, rc.inConn); err == nil {
			msg.Direction = DirectionDOWN
			rc.inChan <- msg
		}
	}
}

func (rc *remoteConn) consumeMessages() {
	try := 0
	for rc.outConn == nil && try < maxRetries {
		if err := rc.newOutSocket(); err != nil {
			errorPrint(err)
			try++
		}
	}
	if try == maxRetries {
		return
	}

	for msg := range rc.outChan {
		if err := writeFromSocket(msg, rc.outConn); err != nil {
			err := rc.newOutSocket()
			if err != nil {
				errorPrint(err)
			}
			rc.outChan <- msg
		}
	}
}

func (rc *remoteConn) newOutSocket() error {
	conn, err := net.Dial("tcp", rc.remoteHost)
	if err != nil {
		return err
	}
	rc.outConn = conn
	return nil
}

func (rc *remoteConn) start() {
	rc.connOn = true
	go rc.listenToSocket()
	go rc.consumeMessages()
}

func (rc *remoteConn) close() {
	rc.connOn = false
	_ = rc.inConn.Close()
	_ = rc.outConn.Close()
	close(rc.inChan)
	close(rc.outChan)
}

// A RemoteManager is used to manage channels with a specific remote host on which a toonnel
// application is running.
// It provides a simple interface to obtain information on the open channels, on how to create new ones
// and how to obtain a reference to an already open one.
type RemoteManager struct {
	remoteConn          *remoteConn // A struct that is used as gateway for the connection layer
	serviceChannel      chan Message
	channels            *bijectiveMap        // List of channel managers for each channel open with this remote host
	channelsOutgoingIfc []reflect.SelectCase // Copy of the open channels used to select a message ready to be forwarded
	mutex               sync.Mutex
}

// Returns a new Manager connected to the passed remote host. If one already exists for the passed host
// then a reference to that is returned.
func Manager(remoteHost string) (*RemoteManager, error) {
	if !started {
		return nil, mwNotStarted
	}

	mutex.Lock()
	defer mutex.Unlock()
	// TODO THIS IS THE FIX MAYBE WRITE IT BETTER
	hostOnly := strings.Split(remoteHost, ":")[0]
	manager, ok := managerMappings[hostOnly]
	if ok {
		return manager, nil
	}

	remConn, connExists := extConnections[hostOnly]
	if !connExists {
		extConnections[hostOnly] = &remoteConn{
			remoteHost: remoteHost,
			inChan:     make(chan Message, bufferSize),
			outChan:    make(chan Message, bufferSize),
		}
	} else {
		remConn.remoteHost = remoteHost
	}

	mng := &RemoteManager{
		remoteConn:     extConnections[hostOnly],
		serviceChannel: make(chan Message, bufferSize),
		channels:       newMap(),
	}
	managerMappings[hostOnly] = mng

	go mng.consumeExtIncoming()
	go mng.produceOutgoing()
	go mng.remoteConn.start()

	return mng, nil
}

// Closes the manager, each single channel/channel managers tied to it and every communication coming
// from the specified remote host.
func (c *RemoteManager) Close() {
	c.remoteConn.inChan <- Message{Type: TypeClose, Direction: DirectionUP}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.channels.closeAll()
	c.remoteConn.close()

	c.channels = nil
	c.channelsOutgoingIfc = nil

	removeConn(c.remoteConn.remoteHost)
}

// Creates and starts a new channel manager which refers to the channel with the passed name.
// If a channel/manager with the passed name already exists this will return an error.
func (c *RemoteManager) NewChan(name string, bufferSize int) (chan Message, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if _, ok := c.channels.getChannel(name); ok {
		return nil, channelAlreadyExists
	}

	channel := make(chan Message, bufferSize)
	c.channels.add(name, channel)
	newChanCase := reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(channel)}
	c.channelsOutgoingIfc = append(c.channelsOutgoingIfc, newChanCase)
	return channel, nil
}

// Closes the channel identified by the passed name, returning an error if it does not exist
func (c *RemoteManager) CloseChan(name string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	channel, ok := c.channels.getChannel(name)
	if !ok {
		return noSuchChannel
	}

	for index, chanSelect := range c.channelsOutgoingIfc {
		if chanSelect.Chan.Interface() == channel {
			// this is pretty heavy but it's done rarely/never on a usually small array
			c.channelsOutgoingIfc = append(c.channelsOutgoingIfc[:index], c.channelsOutgoingIfc[:index+1]...)
		}
	}

	close(channel)
	c.channels.delete(name, channel)
	return nil
}

// Returns the channel identified by the passed name, returning an error if it does not exist
func (c *RemoteManager) GetChan(name string) (chan Message, error) {
	channel, ok := c.channels.getChannel(name)
	if !ok {
		return nil, noSuchChannel
	}
	return channel, nil
}

func (c *RemoteManager) GetRemoteChanList() ([]string, error) {
	chanReq := Message{
		Type:      TypeChanListReq,
		Direction: DirectionUP,
	}
	c.remoteConn.outChan <- chanReq
	resp := <-c.serviceChannel

	var chanList []string
	if err := json.Unmarshal([]byte(resp.Content), &chanList); err != nil {
		return nil, err
	}
	return chanList, nil
}

// Private stuff

func (c *RemoteManager) consumeExtIncoming() {
	for msg := range c.remoteConn.inChan {
		msg.Direction = DirectionDOWN
		go func(msg Message, processFunction func(msg Message) error, inChan chan Message) {
			if err := processFunction(msg); err != nil {
				inChan <- msg
			}
		}(msg, c.processIncoming, c.remoteConn.inChan)
	}
}

func (c *RemoteManager) processIncoming(msg Message) error {
	switch msg.Type {
	case TypeData:
		if err := c.processDataMessage(msg); err != nil {
			return err
		}
	case TypeChanList:
		c.serviceChannel <- msg
	case TypeChanListReq:
		if err := c.processChanListReqMessage(); err != nil {
			return err
		}
	case TypeClose:
		if err := c.processCloseMessage(); err != nil {
			return err
		}
	default:
	}
	return nil
}

func (c *RemoteManager) processDataMessage(msg Message) error {
	channel, ok := c.channels.getChannel(msg.ChannelName)
	if !ok {
		// this message will be put again into the incoming buffer until
		// a channel with the specified name is created
		return noSuchChannel
	}
	channel <- msg
	return nil
}

func (c *RemoteManager) processChanListReqMessage() error {
	chanList := c.channels.getChanNames()
	jsonChanList, err := json.Marshal(chanList)
	if err != nil {
		return err
	}

	chanListMsg := Message{
		Type:      TypeChanList,
		Direction: DirectionUP,
		Content:   string(jsonChanList),
	}

	c.remoteConn.outChan <- chanListMsg
	return nil
}
func (c *RemoteManager) processCloseMessage() error {
	return nil
}

func (c *RemoteManager) produceOutgoing() {
	for {
		index, data, ok := reflect.Select(c.channelsOutgoingIfc)
		// the channel was closed
		if !ok {
			continue
		}

		msg := data.Interface().(Message)
		channel := c.channelsOutgoingIfc[index].Chan.Interface().(chan Message)

		if msg.Direction == DirectionDOWN {
			channel <- msg
			continue
		}
		chanName, _ := c.channels.getName(channel)
		msg.ChannelName = chanName
		c.remoteConn.outChan <- msg
	}
}
