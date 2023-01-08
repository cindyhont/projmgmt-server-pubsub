package pubsub

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/julienschmidt/httprouter"
)

var connections = map[*net.Conn]bool{}

func runWS(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	conn, _, _, err := ws.UpgradeHTTP(req, res)
	if err != nil {
		return
	}

	connections[&conn] = true

	go func() {
		defer delete(connections, &conn)

		var (
			r       = wsutil.NewReader(conn, ws.StateServerSide)
			decoder = json.NewDecoder(r)
		)

		for {
			hdr, err := r.NextFrame()
			if err != nil {
				return
			}

			if hdr.OpCode == ws.OpClose {
				return
			}

			var msg Message
			if err := decoder.Decode(&msg); err != nil {
				return
			}

			if msg.Type == "" {
				continue
			}

			for connection := range connections {
				if connection == &conn {
					continue
				}
				w := wsutil.NewWriter(conn, ws.StateServerSide, ws.OpText)
				e := json.NewEncoder(w)
				e.Encode(res)

				if err := w.Flush(); err != nil {
					fmt.Println(err)
					return
				}
			}
		}
	}()
}

func pingWebsocket() {
	// message has to be map, or the client side will disconnect
	// interval has to be under 60 seconds

	myMap := make(map[string]string)
	for {
		for conn := range connections {
			w := wsutil.NewWriter(*conn, ws.StateServerSide, ws.OpText)
			e := json.NewEncoder(w)
			e.Encode(myMap)

			if err := w.Flush(); err != nil {
				fmt.Println(err)
			}
		}
		time.Sleep(30 * time.Second)
	}
}
