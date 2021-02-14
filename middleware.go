// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package toonnel

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"net"
)

const bufferSize = 20

var (
	// Errors
	incomingError = errors.New("error while trying to read incoming data")
	formatError   = errors.New("data not correctly formatted")
)

func readFromSocket(msg *Message, inConn net.Conn) error {
	data, err := bufio.NewReader(inConn).ReadBytes('\n')
	if err != nil {
		return incomingError
	}
	if err := json.Unmarshal(data, &msg); err != nil {
		return formatError
	}
	return nil
}

func writeFromSocket(msg Message, conn net.Conn) error {
	strMsg, _ := json.Marshal(msg)

	if _, err := fmt.Fprintf(conn, "%s\n", strMsg); err != nil {
		return err
	}
	return nil
}
