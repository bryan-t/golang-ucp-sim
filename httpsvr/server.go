package httpsvr

import (
	"encoding/json"
	"fmt"
	"github.com/bryan-t/golang-ucp-sim/models"
	"github.com/bryan-t/golang-ucp-sim/util"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
)

type deliverFn func(*models.DeliverSMReq)

var deliver deliverFn

// Start starts the http server which serves as the UI
func Start(port int, fn deliverFn) {
	log.Println("Starting http server...")
	deliver = fn
	router := mux.NewRouter()
	router.HandleFunc("/", serveHome)
	router.HandleFunc("/api/failTPS", failTPS)
	router.HandleFunc("/api/successTPS", successTPS)
	router.HandleFunc("/api/messages/deliverBulk", deliverBulk)

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

func deliverBulk(w http.ResponseWriter, r *http.Request) {
	log.Println("Got deliver bulk request")
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.Write([]byte("Encountered error on reading"))
		w.WriteHeader(500)
		return
	}
	log.Println("Got body: ", string(body))
	var bulkDeliverReq models.DeliverSMReqBulk
	err = json.Unmarshal(body, &bulkDeliverReq)
	if err != nil {
		w.Write([]byte("Failed parsing JSON"))
		w.WriteHeader(500)
	}
	go putBulkDeliverReq(&bulkDeliverReq)

	w.WriteHeader(200)

}

func putBulkDeliverReq(bulkDeliverReq *models.DeliverSMReqBulk) {
	for i := range bulkDeliverReq.Requests {
		deliver(&bulkDeliverReq.Requests[i])
	}
}

type homeViewModel struct {
}
