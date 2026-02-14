package tsujido

type Endpoint struct {
	ProposeCh chan Request
	CommitCh  chan Request
	sm        *stateMachine
}

func NewEndpoint() *Endpoint {
	e := &Endpoint{
		ProposeCh: make(chan Request),
		CommitCh:  make(chan Request),
		sm:        newStateMachine(),
	}
	go e.commitLoop()
	return e
}

func (e *Endpoint) Query(key string) Result {
	return e.sm.query(key)
}

func (e *Endpoint) Handle(clientID string, op Operation) Result {
	req := Request{
		ClientID: clientID,
		Op:       op,
		done:     make(chan Result, 1),
	}
	e.ProposeCh <- req
	return <-req.done
}

func (e *Endpoint) Stop() {
	e.sm.stop()
}

func (e *Endpoint) commitLoop() {
	for req := range e.CommitCh {
		res := e.sm.submit(req.ClientID, req.Op)
		req.done <- res
	}
}
