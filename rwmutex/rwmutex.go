//go:build !solution

package rwmutex

type reqType int

const (
	rLock reqType = iota
	rUnlock
	wLock
	wUnlock
)

type request struct {
	typ  reqType
	resp chan struct{}
}

type RWMutex struct {
	req chan request
}

func New() *RWMutex {
	m := &RWMutex{req: make(chan request)}
	go m.actor()
	return m
}

func (m *RWMutex) actor() {
	var readers int
	var writer bool
	var rQ []chan struct{}
	var wQ []chan struct{}

	for rq := range m.req {
		switch rq.typ {
		case rLock:
			if !writer && len(wQ) == 0 {
				readers++
				rq.resp <- struct{}{}
			} else {
				rQ = append(rQ, rq.resp)
			}

		case rUnlock:
			readers--
			if readers == 0 && len(wQ) > 0 {
				writer = true
				ch := wQ[0]
				wQ = wQ[1:]
				ch <- struct{}{}
			}

		case wLock:
			if !writer && readers == 0 {
				writer = true
				rq.resp <- struct{}{}
			} else {
				wQ = append(wQ, rq.resp)
			}

		case wUnlock:
			writer = false
			if len(wQ) > 0 {
				writer = true
				ch := wQ[0]
				wQ = wQ[1:]
				ch <- struct{}{}
			} else {
				for _, ch := range rQ {
					readers++
					ch <- struct{}{}
				}
				rQ = nil
			}
		}
	}
}

func (m *RWMutex) RLock() {
	ch := make(chan struct{})
	m.req <- request{typ: rLock, resp: ch}
	<-ch
}

func (m *RWMutex) RUnlock() {
	m.req <- request{typ: rUnlock}
}

func (m *RWMutex) Lock() {
	ch := make(chan struct{})
	m.req <- request{typ: wLock, resp: ch}
	<-ch
}

func (m *RWMutex) Unlock() {
	m.req <- request{typ: wUnlock}
}
