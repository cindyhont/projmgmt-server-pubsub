package main

import (
	"fmt"

	"github.com/cindyhont/projmgmt-server-pubsub/pubsub"
)

func init() {
	fmt.Println("server pubsub init complete")
}

func main() {
	pubsub.Listen()
}
