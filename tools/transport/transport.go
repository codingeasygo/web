package main

import (
	"os"

	"github.com/codingeasygo/util/xprop"
	"github.com/codingeasygo/web/handler"
)

func main() {
	confFile := "transport.properties"
	if len(os.Args) > 1 {
		confFile = os.Args[1]
	}
	config := xprop.NewConfig()
	err := config.LoadFileWait(confFile, false)
	if err != nil {
		panic(err)
	}
	forward, err := handler.TransportForward(config)
	if err != nil {
		panic(err)
	}
	defer forward.Stop()
	waiter := make(chan int, 1)
	<-waiter
}
