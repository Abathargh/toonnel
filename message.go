// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package toonnel

type MessageType uint
type Direction uint

const (
	TypeError = iota
	TypeData
)

const (
	DirectionUP Direction = iota
	DirectionDOWN
)

type Message struct {
	Direction   Direction `json:"-"` // This is set by the mw when the msg is spawned or recv
	Type        MessageType
	ChannelName string `json:"channelName"`
	Content     string `json:"content"`
}

func StringMessage(content string) Message {
	return Message{Direction: DirectionUP, Content: content}
}

func ErrorMessage(content string) Message {
	return Message{}
}
