package peterson

type state uint

const (
	passive = iota
	active
	lost
	elected
)

type Process struct {
	id       int
	leader   int
	current  int
	neighbor int
	state    state

	receive chan interface{}
	next    chan<- interface{}
}

type oneMessage struct {
	id int
}

type twoMessage struct {
	id int
}

type leaderMessage struct {
	id int
}

func NewInitiator(id int) *Process {
	return &Process{
		id:       id,
		leader:   -1,
		current:  id,
		neighbor: -1,
		state:    active,
		receive:  make(chan interface{}, 1),
	}
}

func NewNonInitiator(id int) *Process {
	return &Process{
		id:       id,
		leader:   -1,
		current:  id,
		neighbor: -1,
		state:    passive,
		receive:  make(chan interface{}, 1),
	}
}

func (p *Process) Id() int {
	return p.id
}

func (p *Process) Leader() int {
	return p.leader
}

func (p *Process) Run() {
	for p.leader == -1 {
		if p.state == active {
			p.next <- oneMessage{p.current}
			msg := <-p.receive
			p.neighbor = msg.(oneMessage).id
			if p.current == p.neighbor {
				p.next <- leaderMessage{p.current}
				p.leader = p.current
				<-p.receive
			} else {
				p.next <- twoMessage{p.neighbor}
				msg := <-p.receive
				id := msg.(twoMessage).id
				if id > p.neighbor && p.current > p.neighbor {
					p.current = p.neighbor
				} else {
					p.state = passive
				}

			}
		} else if p.state == passive {
			msg := <-p.receive
			id := msg.(oneMessage).id
			p.next <- oneMessage{id}

			msg = <-p.receive
			switch msg.(type) {
			case leaderMessage:
				id = msg.(leaderMessage).id
				p.next <- leaderMessage{id}
				p.leader = id
			case twoMessage:
				id = msg.(twoMessage).id
				p.next <- twoMessage{id}
			}
		}
	}
	if p.id == p.leader {
		p.state = elected
	} else {
		p.state = lost
	}
}

func MakeRing(ps []*Process) {
	size := len(ps)
	for i := 0; i < size-1; i++ {
		ps[i].next = ps[i+1].receive
	}
	ps[size-1].next = ps[0].receive

}
