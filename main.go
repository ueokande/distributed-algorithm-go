package main

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/ueokande/distributed-algorithm-go/changroberts"
)

func main() {
	const size = 50
	ps := make([]*changroberts.Process, size)
	for i := 0; i < size; i++ {
		if rand.Intn(2) == 0 {
			ps[i] = changroberts.NewInitiator(i + 1)
		} else {
			ps[i] = changroberts.NewNonInitiator(i + 1)
		}
	}
	for i := len(ps) - 1; i >= 0; i-- {
		j := rand.Intn(i + 1)
		ps[i], ps[j] = ps[j], ps[i]
	}
	changroberts.MakeRing(ps)
	ctx := context.Background()

	type result struct {
		p   *changroberts.Process
		err error
	}
	ch := make(chan result)
	for _, p := range ps {
		go func(p *changroberts.Process) {
			err := p.Run(ctx)
			ch <- result{p, err}
		}(p)
	}

	for i := 0; i < size; i++ {
		r := <-ch
		if r.err != nil {
			fmt.Printf("%v returns an error: %v\n", r.p.Id(), r.err)
		} else {
			fmt.Printf("%v's leader is %v\n", r.p.Id(), r.p.Leader())
		}
	}
	close(ch)

}
