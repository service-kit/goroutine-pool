package pool

import (
	"testing"
	"time"
)

func Test_LoadBalabce(t *testing.T) {
	lb := NewLoadBalancer()
	var count uint32 = 5
	var i uint32 = 0
	for i = 1; i <= count; i++ {
		lb.RegisterLoadCounter(i)
	}
	for i = 1; i <= count; i++ {
		go func() {
			for {
				mid := lb.GetMinLoadID()
				lc := lb.GetLoacCounter(mid)
				lc.AddLoad()
				time.Sleep(50 * time.Millisecond)
				lc.ReduceLoad()
				time.Sleep(50 * time.Millisecond)
			}
		}()
	}
	for {
		time.Sleep(1000 * time.Millisecond)
		lb.ShowLoad()
	}
}
