package tsujido

type Operation struct {
	Key   string
	Value string
}

type Result struct {
	Success bool
	Value   string
}

type Request struct {
	ClientID string
	Op       Operation
	done     chan Result
}
