package pubsub

import (
	"net/http"
	"os"

	"github.com/julienschmidt/httprouter"
)

var Router *httprouter.Router

func init() {
	Router = httprouter.New()
}

func Listen() {
	go pingWebsocket()
	Router.GET("/", runWS)
	http.ListenAndServe(":"+os.Getenv("PROJMGMT_SERVER_PUBSUB_PORT"), Router)
}
