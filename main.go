package main

import (
	"fmt"
	"time"

	"awesomeProject1/subpub"
)

func main() {
	bus := subpub.NewSubPub()

	bus.Subscribe("news", func(msg interface{}) {
		fmt.Println("news received:", msg)
	})

	bus.Publish("news", "hello")

	time.Sleep(200 * time.Millisecond)
}
