package changroberts

import (
	"context"
)

type candidateMessage struct {
	id int
}

type leaderMessage struct {
	id int
}

type Process struct {
	id        int
	leader    int
	initiator bool

	receive chan interface{}
	send    chan<- interface{}
}

func NewInitiator(id int) *Process {
	return &Process{
		id:        id,
		leader:    -1,
		initiator: true,
		receive:   make(chan interface{}, 1),
	}
}

func NewNonInitiator(id int) *Process {
	return &Process{
		id:        id,
		leader:    -1,
		initiator: false,
		receive:   make(chan interface{}, 1),
	}
}

func (p *Process) Id() int {
	return p.id
}

func (p *Process) Leader() int {
	return p.leader
}

func (p *Process) Run(ctx context.Context) error {
	if p.initiator {
		return p.runInitiator(ctx)
	} else {
		return p.runNonInitiator(ctx)
	}
}

func (p *Process) runInitiator(ctx context.Context) error {
	p.send <- candidateMessage{p.id}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg := <-p.receive:
			switch msg.(type) {
			case candidateMessage:
				id := msg.(candidateMessage).id
				if p.id == id {
					p.send <- leaderMessage{p.id}
				} else if p.id < id {
					// do nothing
				} else { // (n.id > id)
					p.send <- candidateMessage{id}
				}
			case leaderMessage:
				id := msg.(leaderMessage).id
				if p.id == id {
					p.leader = p.id
				} else if p.id != id {
					p.leader = id
					p.send <- leaderMessage{id}
				}
				return nil
			}
		}
	}
}

func (p *Process) runNonInitiator(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg := <-p.receive:
			switch msg.(type) {
			case candidateMessage:
				id := msg.(candidateMessage).id
				p.send <- candidateMessage{id}
			case leaderMessage:
				id := msg.(leaderMessage).id
				p.leader = id
				p.send <- leaderMessage{id}
				return nil
			}
		}
	}
}

func MakeRing(ps []*Process) {
	size := len(ps)
	for i := 0; i < size-1; i++ {
		ps[i].send = ps[i+1].receive
	}
	ps[size-1].send = ps[0].receive

}
