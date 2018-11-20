package main

import (
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/hankgao/vote-server/src/api"
)

func main() {

	defer api.CloseDatabaseConn()

	coins := api.RetrieveProjectCoins("Open")
	for _, c := range coins {
		fmt.Println(c.Name)
	}

	// ticker2S := time.NewTicker(time.Second * 1)

	// go func() {
	// 	for {
	// 		select {
	// 		case <-ticker2S.C:
	// 			fmt.Println("2 seconds passed")
	// 		}
	// 	}

	// }()

	// t := time.NewTimer(time.Minute * 1)
	// <-t.C

	// ticker2S.Stop()

	// fmt.Println("Done")

}
