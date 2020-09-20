// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package toonnel

import "testing"

func TestStringMessage(t *testing.T) {
	str := "test"
	msg := StringMessage(str)

	if msg.Content != str {
		t.Errorf("msg.Content != %s", str)
	}
}

func TestMessage_IsValid(t *testing.T) {
	tests := []struct {
		target  Message
		outcome bool
	}{
		{StringMessage("test"), true},
		{Message{Type: 0}, false},
		{Message{Type: TypeChanList + 1}, false},
	}

	for _, test := range tests {
		if output := test.target.IsValid(); output != test.outcome {
			t.Errorf("%+v.IsValid() returned %t instead of %t", test.target, test.outcome, output)
		}
	}
}
