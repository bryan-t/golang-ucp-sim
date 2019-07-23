package httpsvr

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"golang-ucp-sim/models"
	"golang-ucp-sim/util"
	"io/ioutil"
	"log"
	"net/http"
)

type deliverFn func(*models.DeliverSMReq)

var deliver deliverFn

// Start starts the http server which serves as the UI
func NewServer(port int, fn deliverFn) *http.Server {
	log.Println("Starting http server...")
	deliver = fn
	router := mux.NewRouter()
	router.HandleFunc("/", serveHome)
	router.HandleFunc("/api/failTPS", failTPS)
	router.HandleFunc("/api/successTPS", successTPS)
	router.HandleFunc("/api/incomingTPS", incomingTPS)
	router.HandleFunc("/api/setMaxTPS", setMaxTPS)
	router.HandleFunc("/api/messages/deliverBulk", deliverBulk)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: router,
	}

	return server
}
func serveHome(w http.ResponseWriter, r *http.Request) {

}

type tps struct {
	TPS int64
}

func setMaxTPS(w http.ResponseWriter, r *http.Request) {
	log.Println("Got setMaxTPS request")
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println("Error reading body")
		w.WriteHeader(500)
		w.Write([]byte("Encountered error on reading"))
		return
	}
	log.Println("Got body: ", string(body))
	tps := tps{}
	err = json.Unmarshal(body, &tps)
	if err != nil {
		log.Printf("Failed to parse json err:%s\n", err.Error())
		w.WriteHeader(500)
		w.Write([]byte("Failed parsing JSON"))
		return
	}
	util.UpdateMaxTPS(tps.TPS)
}
func incomingTPS(w http.ResponseWriter, r *http.Request) {
	resp := tps{util.GetIncomingTPS()}
	jsonResp, _ := json.Marshal(resp)
	w.Write([]byte(jsonResp))
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
		log.Println("Error reading body")
		w.WriteHeader(500)
		w.Write([]byte("Encountered error on reading"))
		return
	}
	log.Println("Got body: ", string(body))
	var bulkDeliverReq models.DeliverSMReqBulk
	err = json.Unmarshal(body, &bulkDeliverReq)
	if err != nil {
		log.Printf("Failed to parse json err:%s\n", err.Error())
		w.WriteHeader(500)
		w.Write([]byte("Failed parsing JSON"))
		return
	}
	log.Printf("%v\n", bulkDeliverReq)
	putBulkDeliverReq(&bulkDeliverReq)

	log.Println("Done queueing")
	w.WriteHeader(200)

}

func putBulkDeliverReq(bulkDeliverReq *models.DeliverSMReqBulk) {
	for i := range bulkDeliverReq.Requests {
		deliver(&bulkDeliverReq.Requests[i])
	}
}

type homeViewModel struct {
}
