package tsujido

import "sync"

type clientSession struct {
	seq    uint64
	result Result
}

type command struct {
	clientID string
	op       Operation
	done     chan Result
}

type StateMachine struct {
	mu       sync.RWMutex
	data     map[string]string
	sessions map[string]*clientSession
	ch       chan command
	pool     sync.Pool
}

func NewStateMachine() *StateMachine {
	sm := &StateMachine{
		data:     make(map[string]string),
		sessions: make(map[string]*clientSession),
		ch:       make(chan command),
		pool:     sync.Pool{New: func() any { return make(chan Result, 1) }},
	}
	go sm.loop()
	return sm
}

func (sm *StateMachine) Query(key string) Result {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return Result{Success: true, Value: sm.data[key]}
}

func (sm *StateMachine) Submit(clientID string, op Operation) Result {
	done := sm.pool.Get().(chan Result)
	sm.ch <- command{clientID: clientID, op: op, done: done}
	res := <-done
	sm.pool.Put(done)
	return res
}

func (sm *StateMachine) Stop() {
	close(sm.ch)
}

func (sm *StateMachine) loop() {
	for cmd := range sm.ch {
		cmd.done <- sm.apply(cmd.clientID, cmd.op)
	}
}

func (sm *StateMachine) apply(clientID string, op Operation) Result {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.data[op.Key] = op.Value
	res := Result{Success: true, Value: "OK"}

	s := sm.sessions[clientID]
	if s == nil {
		s = &clientSession{}
		sm.sessions[clientID] = s
	}
	s.seq++
	s.result = res
	return res
}
