# *toonnel*

This is a library that uses the channel primitives of the Go programming languages as abstractions to represent remote TCP connections.

## Requirements

You need to have a rather recent version of the Go programming language that supports go modules (I'm currently using Go 1.15 while developing the library).
## Usage

A typical application will initialize the library as soon as it starts and create remote managers as it's needed.
A Remote Manager offers means to manage a remote connection and to spawn and use new channels with the remote machine identified by the address passed to the manager.

## Example
_First machine exposing the toonnel server on port 7777_
```go
package main

import (
	"fmt"
	"github.com/abathargh/toonnel"
	"log"
	"time"
)

func main() {
	if err := toonnel.Start(7777); err != nil {
		log.Fatal(err)
	}

	time.Sleep(2 * time.Second)

	manager, err := toonnel.Manager("127.0.0.1:6666")
	if err != nil {
		log.Fatal(err)
	}

	test, err := manager.NewChan("test", 5)
	if err != nil {
		log.Fatal(err)
	}

	var a byte

	for {
		msg := <-test
		fmt.Println(msg.Content)

		test <- toonnel.StringMessage(fmt.Sprintf("%d", a))
		a++
	}
}
```

_Second machine exposing the tunnel server on port 6666_

```go
package main

import (
	"fmt"
	"github.com/abathargh/toonnel"
	"log"
	"time"
)

func main() {
	if err := toonnel.Start(6666); err != nil {
		log.Fatal(err)
	}

	time.Sleep(2 * time.Second)

	manager, err := toonnel.Manager("127.0.0.1:7777")
	if err != nil {
		log.Fatal(err)
	}

	test, err := manager.NewChan("test", 5)
	if err != nil {
		log.Fatal(err)
	}

	var a byte = 10

	for {
		test <- toonnel.StringMessage(fmt.Sprintf("%d", a))
		time.Sleep(2 * time.Second)
		a++

		msg := <-test
		fmt.Println(msg.Content)
	}
}

```
## TODOs

- Finish adding features (e.g. for the remote host to exist with fault tolerance)
- Support for local inter-process communication with the same abstraction
- Some API may change in the future (e.g. choosing buffer size for channels)