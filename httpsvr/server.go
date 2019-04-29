package httpsvr

import (
	"encoding/json"
	"fmt"
	"github.com/bryan-t/golang-ucp-sim/util"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

// Start starts the http server which serves as the UI
func Start(port int) {
	log.Println("Starting http server...")
	router := mux.NewRouter()
	router.HandleFunc("/", serveHome)
	router.HandleFunc("/api/failTPS", failTPS)
	router.HandleFunc("/api/successTPS", successTPS)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: router,
	}

	log.Fatal(server.ListenAndServe())
	log.Println("Started http server...")
}

func serveHome(w http.ResponseWriter, r *http.Request) {

}

type tps struct {
	TPS int64
}

func successTPS(w http.ResponseWriter, r *http.Request) {
	resp := tps{util.GetSuccessTPS()}
	jsonResp, _ := json.Marshal(resp)
	w.Write([]byte(jsonResp))
}
func failTPS(w http.ResponseWriter, r *http.Request) {
	resp := tps{util.GetFailTPS()}
	jsonResp, _ := json.Marshal(resp)
	w.Write([]byte(jsonResp))
}

type homeViewModel struct {
}
