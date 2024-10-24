package service

import (
	"context"
	"time"

	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/task"
)

type CarePartnerRunner struct {
	logger log.Logger
}

func NewCarePartnerRunner(logger log.Logger) (*CarePartnerRunner, error) {
	return &CarePartnerRunner{
		logger: logger,
	}, nil
}

func (r *CarePartnerRunner) GetRunnerType() string {
	return task.CarePartnerType
}

func (r *CarePartnerRunner) GetRunnerTimeout() time.Duration {
	return 30 * time.Second
}

func (r *CarePartnerRunner) GetRunnerDeadline() time.Time {
	return time.Now().Add(30 * time.Second)
}

func (r *CarePartnerRunner) GetRunnerDurationMaximum() time.Duration {
	return 30 * time.Second
}

func (r *CarePartnerRunner) Run(ctx context.Context, tsk *task.Task) {
	r.logger.Debug("=== this is the care partner task ===")

	// TODO: do the thing; query for patients with followers from whom we haven't seen data
	// in more than > minutes

	// RepeatAvailableAfter controls when the task runs again, however, the lower bound is
	// controlled by the configuration of the task service, which is 5 seconds.
	tsk.RepeatAvailableAfter(time.Until(time.Now().Add(5 * time.Second)))
}
