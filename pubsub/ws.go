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

			if hdr.OpCode == ws.OpPong {
				continue
			}

			var msg Message
			if err := decoder.Decode(&msg); err != nil {
				return
			}

			if msg.Type == "" {
				continue
			} else if msg.Type == "user-status" {
				newUserStatus(&myConn, &msg)
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
	for {
		for conn := range connections {
			(*conn).Write(ws.CompiledPing)
		}
		time.Sleep(30 * time.Second)
	}
}

func newUserStatus(myConn *net.Conn, msg *Message) {
	uid := msg.Payload["id"].(string)
	online := msg.Payload["online"].(bool)

	if online {
		if count, exists := serverUserCount[myConn][uid]; exists {
			serverUserCount[myConn][uid] = count + 1
		} else {
			serverUserCount[myConn] = map[string]int{uid: 1}
		}
	} else {
		if count, exists := serverUserCount[myConn][uid]; exists {
			if count > 1 {
				serverUserCount[myConn][uid] = count - 1
			} else {
				delete(serverUserCount[myConn], uid)
			}
		}
	}
}

func deleteConnection(myConn *net.Conn) {
	// announce to other servers that this connection is lost, and deduct the online user count of that server
	if len(serverUserCount[myConn]) != 0 {
		msg := Message{
			Type:                  "server-disconnect",
			OtherServersUserCount: serverUserCount[myConn],
			ToAllRecipients:       true,
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
	delete(connections, myConn)
	(*myConn).Close()
}
