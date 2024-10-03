package throwaway

import (
	"context"
	"os"
	"time"

	"github.com/tidepool-org/platform/log/devlog"
	"github.com/tidepool-org/platform/task"
	"github.com/tidepool-org/platform/task/queue"
)

type Runner struct {
	timeNow timeNow
}

func NewRunner(timeNow timeNow) *Runner {
	r := &Runner{}
	if timeNow == nil {
		r.timeNow = time.Now
	} else {
		r.timeNow = timeNow
	}
	return r
}

type timeNow func() time.Time

var _ queue.Runner = (*Runner)(nil)

func (r *Runner) GetRunnerType() string {
	return RunnerType
}

const RunnerType = "throwaway"

func (r *Runner) GetRunnerDeadline() time.Time {
	return r.timeNow().Add(RunDurationMax)
}

const RunDurationMax = 5 * time.Second

func (r *Runner) GetRunnerMaximumDuration() time.Duration {
	return RunDurationMax
}

func (r *Runner) Run(ctx context.Context, tsk *task.Task) bool {
	logger, _ := devlog.NewWithDefaults(os.Stderr)
	logger.Info("RUN! throwaway task")
	// This controls when the task runs again, however, the lower bound is the configuration
	// of the task service, which is 5 seconds.
	tsk.RepeatAvailableAfter(time.Until(r.timeNow().Add(time.Second)))
	return false
}
