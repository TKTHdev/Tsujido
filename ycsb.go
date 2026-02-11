package tsujido

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

const (
	ValueMax           = 1500
	ExperimentDuration = 10 * time.Second
)

type YCSBConfig struct {
	Workers  int
	Workload int // write ratio: 50=A, 5=B, 0=C
	Duration time.Duration
	Protocol string // for result output
}

type workerResult struct {
	count    int
	duration time.Duration
}

type YCSBRunner struct {
	submitter Submitter
	cfg       YCSBConfig
}

func NewYCSBRunner(submitter Submitter, cfg YCSBConfig) *YCSBRunner {
	if cfg.Duration == 0 {
		cfg.Duration = ExperimentDuration
	}
	return &YCSBRunner{
		submitter: submitter,
		cfg:       cfg,
	}
}

func (y *YCSBRunner) Run() {
	fmt.Printf("[YCSB] Starting benchmark: protocol=%s workload=%d workers=%d duration=%s\n",
		y.cfg.Protocol, y.cfg.Workload, y.cfg.Workers, y.cfg.Duration)

	ctx, cancel := context.WithTimeout(context.Background(), y.cfg.Duration)
	defer cancel()

	resultCh := make(chan workerResult, y.cfg.Workers)
	var wg sync.WaitGroup

	for i := 0; i < y.cfg.Workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			resultCh <- y.worker(ctx)
		}()
	}

	wg.Wait()
	close(resultCh)

	var totalCount int
	var totalDuration time.Duration
	for res := range resultCh {
		totalCount += res.count
		totalDuration += res.duration
	}

	throughput := float64(totalCount) / y.cfg.Duration.Seconds()
	avgLatency := float64(0)
	if totalCount > 0 {
		avgLatency = float64(totalDuration.Milliseconds()) / float64(totalCount)
	}

	workloadName := "unknown"
	switch y.cfg.Workload {
	case 50:
		workloadName = "ycsb-a"
	case 5:
		workloadName = "ycsb-b"
	case 0:
		workloadName = "ycsb-c"
	}

	fmt.Printf("[YCSB] Completed: ops=%d throughput=%.2f ops/sec latency=%.2f ms\n",
		totalCount, throughput, avgLatency)
	fmt.Printf("RESULT:%s,%s,%d,%.2f,%.2f\n",
		y.cfg.Protocol, workloadName, y.cfg.Workers, throughput, avgLatency)
}

func (y *YCSBRunner) worker(ctx context.Context) workerResult {
	keys := []string{"x", "y", "z", "a", "b", "c"}
	res := workerResult{}

	for {
		select {
		case <-ctx.Done():
			return res
		default:
		}

		op := y.createOp(keys)

		start := time.Now()
		result, err := y.submitter.Submit(ctx, op)
		if err != nil {
			select {
			case <-ctx.Done():
				return res
			default:
			}
			continue
		}

		if result.Success {
			res.count++
			res.duration += time.Since(start)
		}
	}
}

func (y *YCSBRunner) createOp(keys []string) Operation {
	key := keys[rand.Intn(len(keys))]
	if rand.Intn(100) < y.cfg.Workload {
		return Operation{
			Type:  OpSet,
			Key:   key,
			Value: fmt.Sprintf("value%d", rand.Intn(ValueMax)),
		}
	}
	return Operation{
		Type: OpGet,
		Key:  key,
	}
}
