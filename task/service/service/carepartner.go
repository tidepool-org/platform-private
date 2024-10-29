package service

import (
	"context"
	"time"

	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/task"
)

type CarePartnerRunner struct {
	logger  log.Logger
	querier NoCommunicationQuerier
}

type NoCommunicationQuerier interface {
	// UsersWithoutCommunication returns a slice of user ids for those users that haven't
	// uploaded data recently.
	UsersWithoutCommunication(context.Context) ([]string, error)
}

func NewCarePartnerRunner(logger log.Logger, querier NoCommunicationQuerier) (*CarePartnerRunner, error) {
	return &CarePartnerRunner{
		querier: querier,
		logger:  logger,
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
	// RepeatAvailableAfter controls when the task runs again. The lower bound is controlled
	// by the configuration of the task service, which is 5 seconds.
	tsk.RepeatAvailableAfter(time.Until(time.Now().Add(time.Second)))

	// TODO: do the thing; query for patients with followers from whom we haven't seen data
	// in more than > minutes

	// Step 1: get the user ids for all users for whom we haven't seen data in at least 5
	// minutes.
	userIDs, err := r.querier.UsersWithoutCommunication(ctx)
	if err != nil {
		r.logger.WithError(err).Info("unable to list users without communications")
		return
	}

	for _, userID := range userIDs {
		r.logger.WithField("userID", userID).Info("checking for follower alerts for non-communicating user")
	}
}
