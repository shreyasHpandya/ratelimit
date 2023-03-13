package ratelimit

import (
	"context"
	"math"
	"sync"
	"testing"
	"time"
)

func TestRateLimitter_CheckThrottle(t *testing.T) {
	rate := Rate{Limit: 10, Duration: time.Second}
	numClients := 10 //parallel clients that will use rate limiter
	desiredRps := float64(rate.Limit) / rate.Duration.Seconds()
	testDuration := 30.0 //seconds
	reqLatency := time.Millisecond * 100

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(testDuration))
	defer cancel()

	var wg sync.WaitGroup
	for i := 0; i < numClients; i++ {
		wg.Add(1)
		go func() {
			deniedcount := 0
			servedcount := 0
			myrt := NewRateLimitter(
				map[string]Rate{
					"/1": rate,
				},
			)
			for {
				select {
				case <-ctx.Done():
					// t.Logf("requests denied:%v", deniedcount)
					// t.Logf("requests served:%v", servedcount)
					avgReqServed := float64(servedcount) / testDuration
					t.Logf("avgReqServed: %v", avgReqServed)
					if math.Floor(avgReqServed) > desiredRps {
						t.Errorf("served more requests: %v req served", servedcount)
					}
					wg.Done()
					return
				default:
					{
						ok, err := myrt.CheckThrottle("/1")
						time.Sleep(reqLatency)
						if err != nil {
							t.Error(err)
						}
						if ok {
							servedcount++
						} else {
							deniedcount++
						}
					}
				}
			}
		}()
	}

	wg.Wait()

	// t.Run("/1", func(t *testing.T) {

	// 	// simulate requests using sleep
	// 	for i := 0; i < numReq; i++ {
	// 		ok, err := myrt.CheckThrottle("/1")
	// 		time.Sleep(reqLatency)
	// 		if err != nil {
	// 			t.Error(err)
	// 		}
	// 		if ok {
	// 			servedcount++
	// 		} else {
	// 			deniedcount++
	// 		}

	// 	}

	// t.Logf("requests denied:%v", deniedcount)
	// t.Logf("requests served:%v", servedcount)
	// avgReqServed := float64(servedcount) / testDuration
	// t.Logf("avgReqServed: %v", avgReqServed)
	// if math.Floor(avgReqServed) > desiredRps {
	// 	t.Errorf("served more requests: %v req served", servedcount)
	// }
	// ok, err := myrt.CheckThrottle("/1")
	// if !ok {
	// 	t.Error("Did not receive ok, err:", err)
	// }
	// if err != nil {
	// 	t.Error(err)
	// }
	// })
}

func BenchmarkCheckThrottle(b *testing.B) {
	myrt := NewRateLimitter(
		map[string]Rate{
			"/1": {Limit: 100, Duration: time.Second},
		},
	)
	okcount := 0
	notokcount := 0
	for i := 0; i < b.N; i++ {
		ok, err := myrt.CheckThrottle("/1")
		if err != nil {
			b.Errorf("received Error: %s", err)
		}
		if ok {
			okcount++
		} else {
			notokcount++
		}
	}
	b.Logf("okcount: %v, notokcount:%v", okcount, notokcount)
	// b.Run("BenchCheckThrottle", func(b *testing.B) {
	// 	okcount := 0
	// 	notokcount := 0
	// 	for i := 0; i < b.N; i++ {
	// 		ok, err := myrt.CheckThrottle("/1")
	// 		if err != nil {
	// 			b.Errorf("received Error: %s", err)
	// 		}
	// 		if ok {
	// 			okcount++
	// 		} else {
	// 			notokcount++
	// 		}
	// 	}
	// 	b.Logf("okcount: %v, notokcount:%v", okcount, notokcount)
	// })

}
