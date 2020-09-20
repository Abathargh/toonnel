package toonnel

type bijectiveMap struct {
	nameToChan map[string]chan Message
	chanToName map[chan Message]string
}

func newMap() *bijectiveMap {
	return &bijectiveMap{
		nameToChan: make(map[string]chan Message),
		chanToName: make(map[chan Message]string),
	}
}

func (m *bijectiveMap) add(name string, channel chan Message) {
	m.nameToChan[name] = channel
	m.chanToName[channel] = name
}

func (m *bijectiveMap) delete(name string, channel chan Message) {
	delete(m.nameToChan, name)
	delete(m.chanToName, channel)
}

func (m *bijectiveMap) closeAll() {
	for channel := range m.chanToName {
		close(channel)
	}
	m.nameToChan = nil
	m.chanToName = nil
}

func (m *bijectiveMap) getChannel(name string) (chan Message, bool) {
	channel, ok := m.nameToChan[name]
	return channel, ok
}

func (m *bijectiveMap) getName(channel chan Message) (string, bool) {
	name, ok := m.chanToName[channel]
	return name, ok
}

func (m *bijectiveMap) getChanNames() []string {
	index := 0
	chanList := make([]string, len(m.nameToChan))
	for key := range m.nameToChan {
		chanList[index] = key
		index++
	}
	return chanList
}
