package main

import (
	"fmt"
	"sync"
	"time"

	rt "github.com/shreyasHpandya/ratelimit/pkg/ratelimit"
)

func main() {
	apiConf := map[string]rt.Rate{
		"/home": {Limit: 100, Duration: time.Second * 5},
	}
	myrt := rt.NewRateLimitter(apiConf)
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			for {
				ok, err := myrt.CheckThrottle("/home")
				if err != nil {
					fmt.Println("error in RateLimiter", err)
					break
				} else if ok {
					fmt.Println("served")
				} else if !ok {
					fmt.Println("throttle reached")
				}
				time.Sleep(time.Millisecond * 1000)
			}
		}()
	}
	wg.Wait()
}
