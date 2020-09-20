package toonnel

type messageType uint

const (
	TypeUndefined messageType = iota
	TypeData
	TypeClose
	TypeChanListReq
	TypeChanList
)

type Direction uint

const (
	DirectionUP Direction = iota
	DirectionDOWN
)

type Message struct {
	Direction Direction `json:"-"` // This is set by the mw when the msg is spawned or recv

	ChannelName string      `json:"channelName"`
	Type        messageType `json:"type"`
	Content     string      `json:"content"`
}

func (msg Message) IsValid() bool {
	return msg.Type >= TypeData && msg.Type <= TypeChanList
}

func StringMessage(content string) Message {
	return Message{Type: TypeData, Direction: DirectionUP, Content: content}
}
