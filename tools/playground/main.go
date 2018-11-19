package main

import (
	"fmt"
	"time"
)

func main() {
	ticker2S := time.NewTicker(time.Second * 1)

	go func() {
		for {
			select {
			case <-ticker2S.C:
				fmt.Println("2 seconds passed")
			}
		}

	}()

	t := time.NewTimer(time.Minute * 1)
	<-t.C

	ticker2S.Stop()

	fmt.Println("Done")

}
