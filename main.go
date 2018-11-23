package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	skyutil "github.com/hankgao/superwallet-server/server/mobile"
	"github.com/hankgao/vote-server/src/api"
	"github.com/skycoin/skycoin/src/util/droplet"
	"github.com/skycoin/skycoin/src/visor"
	"github.com/skycoin/skycoin/src/wallet"
)

// endpoints

// - projectcoins?status=ongoing,queuing, finished,aborted

const version = "0.0.1"

var (
	nodeServer = "http://localhost"
	serverPort = "6789"
)

func main() {

	defer api.CloseDatabaseConn()

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

	// We need to check balance of every project coin every 1 minute
	// Once the balance reaches the cap, we change the status from open to closed
	// then we notify someone of this event,
	// someone maybe Mr. Fu
	// he will create an internal asset for that project coin
	// allow users to deposit and withdraw
	go func() {
		balanceCheckTicker := time.NewTicker(time.Second * 30)
		for {
			select {
			case <-balanceCheckTicker.C:
				checkBalance()
			}
		}
	}()

	err := srv.ListenAndServe()
	if err != nil {
		// log.Errorf("Failed to start server: %s", err)
	}

}

func checkBalance() {
	coins := api.RetrieveProjectCoins("Open")
	for _, c := range coins {
		balance, err := skyutil.GetBalance(c.PlatformCoinName, c.VotingAddress)
		if err != nil {
			// Log warning
			return
		}

		bp := wallet.BalancePair{}
		json.Unmarshal([]byte(balance), &bp)

		bs, err := droplet.ToString(bp.Confirmed.Coins)
		if err != nil {
			// Log error
		}

		fmt.Printf("%s(%s) => %s\n", c.Name, c.PlatformCoinName, bs)
		if bp.Confirmed.Coins >= uint64(c.VoteCap)*1000000 {
			// we got enough coins
			// change the status from Open to Closed
			err := api.UpdateStatus("Closed", c.Name)
			if err != nil {
				// Log warning!
				return
			}

			// Notify someone of the event

		}

	}
}

// Note that name and status cannot coexist
// Will check name first
//
func getProjectCoins(w http.ResponseWriter, r *http.Request) {

	coins := api.ProjectCoins{}

	values := r.URL.Query()

	status := ""
	coinName := values.Get("name")

	if coinName == "" {
		status = values.Get("status")
		coins = api.RetrieveProjectCoins(status)
	} else {
		coin := api.GetProjectCoin(coinName)
		coins = append(coins, coin)
	}

	coinsJson, err := json.MarshalIndent(coins, "", "    ")
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

			timeStamp := time.Unix(int64(readableTx.Timestamp), 0)

			// We do find the transaction
			res := struct {
				Time  string `json:"time"`
				Coins string `json:"coins"`
			}{
				Time:  timeStamp.Format("2006-01-02 15:04:05"),
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
