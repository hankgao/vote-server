package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/hankgao/vote-server/src/api"
)

// endpoints

// - projectcoins?status=ongoing,queuing, finished,aborted

const version = "0.0.1"

var (
	nodeServer = "http://localhost"
	serverPort = "6789"
)

func main() {
	r := mux.NewRouter()

	// /api/projectcoins?status=Open
	r.HandleFunc("/api/projectcoins", getProjectCoins)

	// /api/invoice?coin=mzcoin&txid=skdfkdfkasfwerskdfadsfnas
	r.HandleFunc("/api/invoice", invoiceHandler)
	r.PathPrefix("/logo").Handler(http.StripPrefix("/logo", http.FileServer(http.Dir("./logos"))))
	r.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.Dir("./dist"))))
	http.Handle("/", r)

	// start server
	srv := &http.Server{
		Addr: "0.0.0.0:" + serverPort,
		// Good practice to set timeouts to avoid Slowloris attacks.
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      r, // Pass our instance of gorilla/mux in.
	}

	err := srv.ListenAndServe()
	if err != nil {
		// log.Errorf("Failed to start server: %s", err)
	}

	// We need to check balance of every project coin every 1 minute
	// Once the balance reaches the cap, we change the status from open to closed
	// then we notify someone of this event,
	// someone maybe Mr. Fu
	// he will create an internal asset for that project coin
	// allow users to deposit and withdraw
	// balanceCheckTicker := time.NewTicker(time.Minute * 1)
	// for {
	// 	select {
	// 	case <-balanceCheckTicker.C:
	// 		// updateBalance()
	// 	}
	// }

}

func getProjectCoins(w http.ResponseWriter, r *http.Request) {
	status := "Open"

	coins := api.RetrieveProjectCoins(status)

	coinsJson, err := json.MarshalIndent(coins, "", "")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(coinsJson)
}

func invoiceHandler(w http.ResponseWriter, r *http.Request) {
	// vars := mux.Vars(r)

	// coinName := vars["coin"]
	// txID := vars["txid"]

	// coin := api.GetProjectCoin(coinName)

	w.Write([]byte("Yeah, this is invoice handler!"))
}
