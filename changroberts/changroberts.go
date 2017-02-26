package changroberts

import (
	"context"
)

type State uint

const (
	Sleep = iota
	Lost
	Candidate
	Elected
)

func (s State) String() string {
	switch s {
	case Sleep:
		return "Sleep"
	case Lost:
		return "Lost"
	case Candidate:
		return "Candidate"
	case Elected:
		return "Elected"
	default:
		panic("invalid type")
	}
}

type CandidateMessage struct {
	Id int
}

type LeaderMessage struct {
	Id int
}

type Process struct {
	id        int
	leader    int
	state     State
	initiator bool

	receive chan interface{}
	send    chan<- interface{}
}

func NewInitiator(id int) *Process {
	return &Process{
		id:        id,
		leader:    -1,
		state:     Candidate,
		initiator: true,
		receive:   make(chan interface{}, 1),
	}
}

func NewNonInitiator(id int) *Process {
	return &Process{
		id:        id,
		leader:    -1,
		state:     Lost,
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

func (p *Process) State() State {
	return p.state
}

func (p *Process) Run(ctx context.Context) error {
	if p.initiator {
		return p.runInitiator(ctx)
	} else {
		return p.runNonInitiator(ctx)
	}
}
func (p *Process) runInitiator(ctx context.Context) error {
	p.send <- CandidateMessage{p.id}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg := <-p.receive:
			switch msg.(type) {
			case CandidateMessage:
				id := msg.(CandidateMessage).Id
				if p.id == id {
					p.state = Elected
					p.send <- LeaderMessage{p.id}
				} else if p.id < id {
					// do nothing
				} else { // (n.id > id)
					p.state = Lost
					p.send <- CandidateMessage{id}
				}
			case LeaderMessage:
				id := msg.(LeaderMessage).Id
				if p.id == id {
					p.leader = p.id
				} else if p.id != id {
					p.leader = id
					p.send <- LeaderMessage{id}
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
			case CandidateMessage:
				id := msg.(CandidateMessage).Id
				p.send <- CandidateMessage{id}
			case LeaderMessage:
				id := msg.(LeaderMessage).Id
				p.leader = id
				p.send <- LeaderMessage{id}
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
