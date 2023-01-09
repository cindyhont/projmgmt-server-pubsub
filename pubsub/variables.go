package pubsub

import "net"

var (
	serverUserCount = map[*net.Conn]map[string]int{}
	connections     = map[*net.Conn]bool{}
)
