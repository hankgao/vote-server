package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	skyutil "github.com/hankgao/superwallet-server/server/mobile"
	"github.com/hankgao/vote-server/src/api"
	"github.com/skycoin/skycoin/src/visor"
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

	values := r.URL.Query()
	status := values.Get("status")

	coins := api.RetrieveProjectCoins(status)

	coinsJson, err := json.MarshalIndent(coins, "", "")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(coinsJson)
}

func invoiceHandler(w http.ResponseWriter, r *http.Request) {
	// /api/invoice?coin=mzcoin&txid=skdfkdfkasfwerskdfadsfnas
	values := r.URL.Query()
	coinName := values.Get("coin")
	txID := values.Get("txid")

	coin := api.GetProjectCoin(coinName)

	if coin.Name == "" {
		// Log error here
		w.Write([]byte(fmt.Sprintf("not found - %s not found", coinName)))
		return
	}

	transaction, err := skyutil.GetTransaction(coin.PlatformCoinName, txID)
	if err != nil {
		// Log error
		w.Write([]byte("not found - failed to call get transaction"))
		return
	}

	readableTx := visor.ReadableTransaction{}
	err = json.Unmarshal([]byte(transaction), &readableTx)
	if err != nil {
		// Not a valid transaction response
		// We assume that the transaction is not found
		w.Write([]byte("not found - txid not found"))
		return

	}

	for _, o := range readableTx.Out {
		if o.Address == coin.VotingAddress {
			// We do find the transaction
			res := struct {
				Time  string `json:"time"`
				Coins string `json:"coins"`
			}{
				Time:  "2018-10-20 14:00:00",
				Coins: o.Coins,
			}

			resJSON, err := json.Marshal(res)
			if err != nil {
				// Log error
				w.Write([]byte("not found - failed to do JSON marshal"))
				return
			}

			w.Write(resJSON)
			return

		}
	}

	w.Write([]byte("not found - coins not deposited to voting address"))

}
