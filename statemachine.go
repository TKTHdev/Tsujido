package tsujido

type OpType uint8

const (
	OpGet    OpType = 0
	OpSet    OpType = 1
	OpDelete OpType = 2
)

type Operation struct {
	Type  OpType
	Key   string
	Value string
}

type Result struct {
	Success bool
	Value   string
}

type StateMachine interface {
	Apply(op Operation) Result
}

type KVStore struct {
	Data map[string]string
}

func NewKVStore() *KVStore {
	return &KVStore{Data: make(map[string]string)}
}

func (kv *KVStore) Apply(op Operation) Result {
	switch op.Type {
	case OpGet:
		val, ok := kv.Data[op.Key]
		if ok {
			return Result{Success: true, Value: val}
		}
		return Result{Success: true, Value: ""}
	case OpSet:
		kv.Data[op.Key] = op.Value
		return Result{Success: true, Value: "OK"}
	case OpDelete:
		delete(kv.Data, op.Key)
		return Result{Success: true, Value: "OK"}
	default:
		return Result{Success: false, Value: "unknown operation"}
	}
}
