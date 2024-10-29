package service

import (
	"context"
	"time"

	"github.com/tidepool-org/platform/alerts"
	"github.com/tidepool-org/platform/auth"
	"github.com/tidepool-org/platform/data/events"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/push"
	"github.com/tidepool-org/platform/task"
)

type CarePartnerRunner struct {
	logger log.Logger

	alerts       AlertsClient
	deviceTokens auth.DeviceTokensClient
	pusher       events.Pusher
}

type AlertsClient interface {
	List(context.Context, string) ([]*alerts.Config, error)

	// UsersWithoutCommunication returns a slice of user ids for those users that haven't
	// uploaded data recently.
	UsersWithoutCommunication(context.Context) ([]string, error)
}

func NewCarePartnerRunner(logger log.Logger, alerts AlertsClient,
	deviceTokens auth.DeviceTokensClient, pusher events.Pusher) (*CarePartnerRunner, error) {

	logger.Info("NewCarePartnerRunner") // TODO: remove
	return &CarePartnerRunner{
		logger: logger,

		alerts:       alerts,
		deviceTokens: deviceTokens,
		pusher:       pusher,
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
	// RepeatAvailableAfter controls when the task runs again. The lower bound is controlled by
	// the configuration of the task service, which has a minimum time to wait between
	// cycles. If the value specified is evaluated to be in the past, the task is marked as
	// failed, and not run again.
	nextDesiredRun := time.Now().Add(time.Second)
	defer func() {
		now := time.Now()
		if nextDesiredRun.Before(now) {
			r.logger.Info("care partner is bumping nextDesired")
			// nextDesiredRun must be in the future when it's read by the task queue or the task
			// will be marked failed (and not run again). One workaround is to take a guess at
			// how long it will take Run() to return and the task queue to evaluate the task's
			// AvailableAfter time. This is more or less a guess. Maybe the task queue could be
			// re-worked to accept a value that indicates "as soon as possible"? Or if it
			// accepted a time.Duration, then time.Nanosecond could be set, and the Zero value
			// might mean don't repeat. Or the Zero value could mean repeat now. Or a negative
			// value could mean repeat now. Whatever. It would prevent the task from being
			// marked a failure for not being able to guess when the value would be read. Which
			// wasn't its intent I'm sure, it just wasn't designed for tasks with the level of
			// resolution and repetition expected for this purpose.
			nextDesiredRun = now.Add(25 * time.Millisecond)
		}
		tsk.RepeatAvailableAfter(time.Until(nextDesiredRun))
	}()
	r.logger.Info("care partner no communication check")

	// time.Sleep(250 * time.Millisecond)
	// r.logger.Info("care partner has stayed a while... to listen")

	userIDs, err := r.alerts.UsersWithoutCommunication(ctx)
	if err != nil {
		r.logger.WithError(err).Info("unable to list users without communications")
		return
	}

	notifications := []*alerts.Notification{}
	for _, userID := range userIDs {
		lgr := r.logger.WithField("followedUserID", userID)
		lgr.Info("checking for follower alerts for non-communicating user")
		configs, err := r.alerts.List(ctx, userID)
		if err != nil {
			r.logger.WithError(err).Info("unable to list follower alerts configs")
			continue
		}
		for _, config := range configs {
			if config.Alerts.NoCommunication != nil && config.Alerts.NoCommunication.Enabled {
				noComms := config.Alerts.NoCommunication
				lastReceived := time.Time{}
				notification := noComms.EvaluateLastReceived(ctx, lastReceived)
				if notification != nil {
					notifications = append(notifications, notification)
				}
			}
		}
	}

	for _, notification := range notifications {
		lgr := r.logger.WithField("recipientUserID", notification.RecipientUserID)
		tokens, err := r.deviceTokens.GetDeviceTokens(ctx, notification.RecipientUserID)
		if err != nil {
			lgr.WithError(err).Info("unable to retrieve device tokens")
		}
		if len(tokens) == 0 {
			lgr.Debug("no device tokens found, won't push any notifications")
		}
		pushNote := push.FromAlertsNotification(notification)
		for _, token := range tokens {
			if err := r.pusher.Push(ctx, token, pushNote); err != nil {
				lgr.WithError(err).Info("unable to push notification")
			}
		}
	}
}
