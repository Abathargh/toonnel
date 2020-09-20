package toonnel

import (
	"reflect"
	"sort"
	"sync"
	"testing"
)

func Test_newMap(t *testing.T) {
	bMap := newMap()
	if bMap.nameToChan == nil || bMap.chanToName == nil {
		t.Errorf("Failed to initialize correctly the map: %+v", bMap)
	}
}

func TestBijectiveMap_add(t *testing.T) {
	bMap := newMap()

	name, channel := "test", make(chan Message)
	bMap.add(name, channel)

	actualChannel, okName := bMap.nameToChan[name]
	actualName, okChan := bMap.chanToName[channel]

	if !okName || !okChan || name != actualName || channel != actualChannel {
		t.Errorf("add failed: expected nameMappings[%s] = %v, chanMappings[%v] = %s"+
			" got nameMappings[%s] = %v, chanMappings[%v] = %s",
			name, channel, channel, name, name, actualChannel, channel, actualName)
	}
}

func TestBijectiveMap_delete(t *testing.T) {
	bMap := newMap()

	name, channel := "test", make(chan Message)
	bMap.add(name, channel)

	bMap.delete(name, channel)

	actualChannel, okName := bMap.nameToChan[name]
	actualName, okChan := bMap.chanToName[channel]

	if okName || okChan {
		t.Errorf("add failed: expected nameMappings[%s]: not present, chanMappings[%v]: not present"+
			" got nameMappings[%s] = %v, chanMappings[%v] = %s",
			name, channel, name, actualChannel, channel, actualName)
	}
}

func TestBijectiveMap_closeAll(t *testing.T) {
	bMap := newMap()
	var wg sync.WaitGroup

	targets := []*struct {
		name    string
		closed  bool
		channel chan Message
	}{
		{"test", false, make(chan Message)},
		{"test1", false, make(chan Message)},
		{"test2", false, make(chan Message)},
	}

	wg.Add(len(targets))

	for _, target := range targets {
		bMap.add(target.name, target.channel)
		go func(target *struct {
			name    string
			closed  bool
			channel chan Message
		}) {
			for range target.channel {
			}
			target.closed = true
			wg.Done()
		}(target)
	}

	bMap.closeAll()
	wg.Wait()

	for _, target := range targets {
		if !target.closed {
			t.Errorf("expected closed channel for entry with name '%s': found open", target.name)
		}
	}
}

func TestBijectiveMap_getChannel(t *testing.T) {
	bMap := newMap()
	name, channel := "test", make(chan Message)

	bMap.add(name, channel)
	if actualChan, ok := bMap.getChannel(name); !ok || actualChan != channel {
		t.Errorf("expected bMap.getChan(%s) = %v, got %v", name, channel, actualChan)
	}
}

func TestBijectiveMap_getName(t *testing.T) {
	bMap := newMap()
	name, channel := "test", make(chan Message)

	bMap.add(name, channel)
	if actualName, ok := bMap.getName(channel); !ok || actualName != name {
		t.Errorf("expected bMap.getName(%v) = %s, got %s", channel, name, actualName)
	}
}

func TestBijectiveMap_getChanName(t *testing.T) {
	bMap := newMap()

	names := []string{
		"test",
		"test1",
		"test2",
	}

	for _, name := range names {
		bMap.add(name, nil)
	}

	actualNames := bMap.getChanNames()
	sort.Slice(actualNames, func(i, j int) bool { return actualNames[i] < actualNames[j] })

	if !reflect.DeepEqual(names, actualNames) {
		t.Errorf("expected bmap.getChanNames() = %v, got %v", names, actualNames)
	}
}
