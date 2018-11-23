package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	skyutil "github.com/hankgao/superwallet-server/server/mobile"
	"github.com/hankgao/vote-server/src/api"
	log "github.com/sirupsen/logrus"
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

type ErrorResponse struct {
	Status string `json:"status"` // failed
	Msg    string `json:"msg"`
}

func init() {
	log.SetLevel(log.InfoLevel)
}

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
		balanceCheckTicker := time.NewTicker(time.Second * 60)
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
			log.Error("failed to call GetBalance: ", err)
			return
		}

		bp := wallet.BalancePair{}
		json.Unmarshal([]byte(balance), &bp)

		bs, err := droplet.ToString(bp.Confirmed.Coins)
		if err != nil {
			log.Error("failed to convert droplets to float64 string", bp.Confirmed.Coins)
		}

		bal, err := strconv.ParseFloat(bs, 64)
		if err != nil {
			log.Error("failed to parse float from ", bs)
		} else {
			api.UpdateBalance(bal, c.Name)

		}

		if bp.Confirmed.Coins >= uint64(c.VoteCap)*1000000 {
			// we got enough coins
			// change the status from Open to Closed
			err := api.UpdateStatus("Closed", c.Name)
			if err != nil {
				log.Error("failed to update status from Open to Close for ", c.Name)
				return
			}

			// Notify someone of the event
			log.Info("status is changed from Open to Closed for ", c.Name)

		}

	}
}

// Note that name and status cannot coexist
// Will check name first
//
func getProjectCoins(w http.ResponseWriter, r *http.Request) {

	log.Info("GET /api/projectCoins?", r.URL.RawQuery)

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
		respondErrorMsg(w, err.Error())
		return
	}

	w.Write(coinsJson)

}

func respondErrorMsg(w http.ResponseWriter, msg string) error {
	er := ErrorResponse{
		Status: "failed",
		Msg:    msg,
	}

	bytesR, _ := json.MarshalIndent(er, "", "    ")

	w.Write(bytesR)

	return nil
}

func invoiceHandler(w http.ResponseWriter, r *http.Request) {

	log.Info("GET /api/invoice?", r.URL.RawQuery)

	// /api/invoice?coin=mzcoin&txid=skdfkdfkasfwerskdfadsfnas
	values := r.URL.Query()
	coinName := values.Get("coin")
	txID := values.Get("txid")

	coin := api.GetProjectCoin(coinName)

	if coin.Name == "" {
		log.Error("failed to call GetProjectCoin for ", coinName)
		respondErrorMsg(w, fmt.Sprintf("failed to call GtProjectCoin for %s", coinName))

		return
	}

	transaction, err := skyutil.GetTransaction(coin.PlatformCoinName, txID)
	if err != nil {
		log.Error("failed to call GetTransaction for ", coin.PlatformCoinName, "with txid=", txID)
		respondErrorMsg(w, fmt.Sprintf("failed to call GetTransaction for %s with txid=%s", coin.PlatformCoinName, txID))
		return
	}

	readableTx := visor.ReadableTransaction{}
	err = json.Unmarshal([]byte(transaction), &readableTx)
	if err != nil {
		// Not a valid transaction response
		// We assume that the transaction is not found
		log.Warn("failed to find txid ", txID, "coin: ", coin.PlatformCoinName, "=>", err)
		log.Warn(transaction)
		respondErrorMsg(w, fmt.Sprintf("txid %s not found", txID))
		return

	}

	for _, o := range readableTx.Out {
		if o.Address == coin.VotingAddress {

			loc, _ := time.LoadLocation("Asia/Shanghai")
			timeStamp := time.Unix(int64(readableTx.Timestamp), 0).In(loc)

			// We do find the transaction
			res := struct {
				Time  string `json:"time"`
				Coins string `json:"coins"`
			}{
				Time:  timeStamp.Format("2006-01-02 15:04:05"),
				Coins: o.Coins,
			}

			resJSON, err := json.MarshalIndent(res, "", "    ")
			if err != nil {
				log.Error("failed to marshal ", res)
				respondErrorMsg(w, fmt.Sprintf("failed to marshal: %s", err.Error()))
				return
			}

			w.Write(resJSON)
			return

		}
	}

	respondErrorMsg(w, fmt.Sprintf("coins were not deposited to the voting address"))

}
