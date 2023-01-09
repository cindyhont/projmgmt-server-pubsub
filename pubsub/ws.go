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

func runWS(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	myConn, _, _, err := ws.UpgradeHTTP(req, res)
	if err != nil {
		return
	}

	connections[&myConn] = true
	serverUserCount[&myConn] = make(map[string]int)

	go func() {
		defer deleteConnection(&myConn)

		var (
			r       = wsutil.NewReader(myConn, ws.StateServerSide)
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

			for conn := range connections {
				if conn == &myConn {
					continue
				}

				w := wsutil.NewWriter(*conn, ws.StateServerSide, ws.OpText)
				e := json.NewEncoder(w)
				e.Encode(msg)

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

	// myMap := make(map[string]string)
	for {
		for conn := range connections {
			(*conn).Write(ws.CompiledPing)
			// w := wsutil.NewWriter(*conn, ws.StateServerSide, ws.OpText)
			// e := json.NewEncoder(w)
			// e.Encode(myMap)

			// if err := w.Flush(); err != nil {
			// 	fmt.Println(err)
			// }
		}
		time.Sleep(30 * time.Second)
	}
}

func deleteConnection(myConn *net.Conn) {
	// announce to other servers that this connection is lost, and deduct the online user count of that server
	if len(serverUserCount[myConn]) != 0 {
		msg := Message{
			Type:                  "server-disconnect",
			OtherServersUserCount: serverUserCount[myConn],
		}

		for conn := range connections {
			if conn == myConn {
				continue
			}

			w := wsutil.NewWriter(*conn, ws.StateServerSide, ws.OpText)
			e := json.NewEncoder(w)
			e.Encode(msg)

			if err := w.Flush(); err != nil {
				fmt.Println(err)
				return
			}
		}
	}
	fmt.Println(myConn)
	delete(connections, myConn)
}
