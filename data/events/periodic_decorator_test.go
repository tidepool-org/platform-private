package events

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/log/devlog"
)

func TestPeriodicDecoratorSoloGoldenPath(t *testing.T) {
	test := newPeriodicDecoratorTest(t)
	taskCalls := []time.Time{}
	task := func(ctx context.Context) error {
		taskCalls = append(taskCalls, time.Now())
		return nil
	}
	tolerance := 5 * time.Millisecond
	period := 10 * time.Millisecond
	decorator := NewPeriodicDecoratorSolo(test.runner, period, task)

	test.RunFor(period, decorator)

	// It's impossible to control the timing exactly, but we should always have at least one
	// call. If there are multiple calls, they should e reasonably spaced (within
	// `tolerance`).
	if len(taskCalls) == 0 {
		t.Errorf("expected at least 1 task to run, got 0")
	} else if len(taskCalls) > 1 {
		last := taskCalls[0]
		for _, call := range taskCalls[1:] {
			delay := call.Sub(last)
			if delay < (period-tolerance) || delay > (period+tolerance) {
				t.Errorf("time between calls was %s, expected %s±%s", delay, period, tolerance)
			}
			last = call
		}
	}
}

type periodicDecoratorTest struct {
	*testing.T
	ctx    context.Context
	cancel context.CancelFunc
	runner *mockSaramaEventsRunner
	logger log.Logger
}

func newPeriodicDecoratorTest(t *testing.T) *periodicDecoratorTest {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(func() { cancel() })
	logger, err := devlog.NewWithDefaults(os.Stderr)
	if err != nil {
		t.Fatalf("creating devlog: %s", err)
	}
	ctx = log.NewContextWithLogger(ctx, logger)
	return &periodicDecoratorTest{
		T:      t,
		ctx:    ctx,
		runner: &mockSaramaEventsRunner{},
		logger: logger,
		cancel: cancel,
	}
}

func (t *periodicDecoratorTest) RunFor(d time.Duration, decorator SaramaEventsRunner) {
	t.Helper()
	var wg sync.WaitGroup
	wg.Add(1)
	go func() { defer wg.Done(); decorator.Run(t.ctx) }()
	<-time.After(d)
	t.cancel()
	wg.Wait()
}

type mockSaramaEventsRunner struct{}

func (r *mockSaramaEventsRunner) Run(ctx context.Context) error {
	<-ctx.Done()
	return nil
}
