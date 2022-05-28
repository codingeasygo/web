package main

import (
	"os"

	"github.com/codingeasygo/web/handler"
)

func main() {
	var err error
	forward := handler.NewTransportForward()
	if len(os.Args) > 1 {
		err = forward.Config.LoadFileWait(os.Args[1], false)
		if err != nil {
			panic(err)
		}
	}
	err = forward.Start()
	if err != nil {
		panic(err)
	}
	defer forward.Stop()
	waiter := make(chan int, 1)
	<-waiter
}
