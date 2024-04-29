package evalalerts

import (
	"context"
	"os"
	"slices"
	"sort"
	"time"

	"github.com/tidepool-org/platform/alerts"
	nontypesglucose "github.com/tidepool-org/platform/data/blood/glucose"
	"github.com/tidepool-org/platform/data/store"
	"github.com/tidepool-org/platform/data/summary/types"
	"github.com/tidepool-org/platform/data/types/blood/glucose"
	"github.com/tidepool-org/platform/data/types/blood/glucose/continuous"
	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/log/devlog"
)

func Evaluate(ctx context.Context, repo store.DataRepository, alertsClient *alerts.Client, followedUserID, userID string) (err error) {

	lgr := log.LoggerFromContext(ctx)
	if lgr == nil {
		lgr, _ = devlog.NewWithDefaults(os.Stderr)
	}

	config, err := alertsClient.Get(ctx, followedUserID, userID)
	if err != nil {
		return errors.Wrap(err, "getting client config")
	}

	if config.Alerts.NoCommunication != nil && config.Alerts.NoCommunication.Enabled {
		// If the user is already being notified about NoCommunication alerts,
		// then there's no point in expecting recent data from which to send
		// any other alerts.
		if config.Alerts.NoCommunication.IsActive() {
			return nil
		}
	}

	datum, err := FollowedUsersDatum(ctx, alertsClient, repo, config.FollowedUserID)
	if err != nil {
		return errors.Wrap(err, "retrieving user data")
	}

	if config.Alerts.UrgentLow != nil && config.Alerts.UrgentLow.Enabled {
		changed, err := evaluateUrgentLow(ctx, datum, config.UrgentLow)
		if err != nil {
			return errors.Wrap(err, "evaluating urgent low")
		}
		if changed {
			if err := alertsClient.Upsert(ctx, config); err != nil {
				return errors.Wrap(err, "upserting alerts info after urgent low")
			}
		}
		return nil
	}
	if config.Alerts.Low != nil && config.Alerts.Low.Enabled {
		changed, err := evaluateLow(ctx, datum, config.Low)
		if err != nil {
			return err
		}
		if changed {
			if err := alertsClient.Upsert(ctx, config); err != nil {
				return err
			}
		}
		return nil
	}
	if config.Alerts.High != nil && config.Alerts.High.Enabled {
		changed, err := evaluateHigh(ctx, datum, config.High)
		if err != nil {
			return err
		}
		if changed {
			if err := alertsClient.Upsert(ctx, config); err != nil {
				return err
			}
		}
	}

	lgr.Debugf("Evaluate is returning nil")
	return nil
}

// evaluateUrgentLow determines if an alert should be sent. It returns an
// updated Activity if changes need to be saved.
func evaluateUrgentLow(ctx context.Context, datum *glucose.Glucose, alert *alerts.UrgentLowAlert) (_ bool, err error) {

	if datum.Blood.Units == nil || datum.Blood.Value == nil || datum.Blood.Time == nil {
		return false, errors.New("Unable to evaluate datum: Units, Value, or Time is nil")
	}
	threshold := nontypesglucose.NormalizeValueForUnits(&alert.Threshold.Value, datum.Blood.Units)
	if threshold == nil {
		return false, errors.New("Unable to calculate the alert's normalized threshold")
	}

	alertIsFiring := *datum.Blood.Value < *threshold
	lgr := log.LoggerFromContext(ctx)
	if lgr == nil {
		lgr, _ = devlog.NewWithDefaults(os.Stderr)
	}
	lgr.WithFields(log.Fields{
		"blood":     datum.Blood,
		"threshold": *threshold,
		"isFiring?": alertIsFiring,
	}).Debug("urgent low")
	if !alertIsFiring {
		alert.Activity.Resolved = time.Now()
		return true, nil
	}
	if !alert.IsActive() {
		alert.Activity.Triggered = time.Now()
	}
	if err := sendNotificationUrgentLow(ctx, datum, alert); err != nil {
		return false, err
	}
	alert.Activity.Notified = time.Now()
	return true, nil
}

func sendNotificationUrgentLow(ctx context.Context, datum *glucose.Glucose, alert *alerts.UrgentLowAlert) error {
	lgr := log.LoggerFromContext(ctx)
	if lgr == nil {
		lgr, _ = devlog.NewWithDefaults(os.Stderr)
	}
	lgr.Debug("FIRE AN URGENT LOW NOTIFICATION")
	return nil
}

func evaluateLow(ctx context.Context, datum *glucose.Glucose, alert *alerts.LowAlert) (bool, error) {
	return false, nil
}

func evaluateHigh(ctx context.Context, datum *glucose.Glucose, alert *alerts.HighAlert) (bool, error) {
	return false, nil
}

func evaluateNoCommunication(ctx context.Context, datum *glucose.Glucose, alert *alerts.NoCommunicationAlert) error {
	return nil
}

func evaluateNotLooping(ctx context.Context, datum *glucose.Glucose, alert *alerts.NotLoopingAlert) error {
	return nil
}

func FollowedUsersDatum(ctx context.Context, client *alerts.Client, repo store.DataRepository, userID string) (*glucose.Glucose, error) {
	lgr := log.LoggerFromContext(ctx)
	if lgr == nil {
		lgr, _ = devlog.NewWithDefaults(os.Stderr)
	}
	var status *types.UserLastUpdated
	// Retrieve the most recent upload only. The intention is for this
	// function to be triggered for each new update, so only the latest
	// datapoint needs to be retrieved.
	since := time.Now().Add(-24 * time.Hour) // TODO: Tune this value
	status, err := repo.GetLastUpdatedForUser(ctx, userID, continuous.Type, since)
	if err != nil {
		return nil, err
	}

	lgr.WithField("userID", userID).WithField("status", status).Debug("GetLastUpdatedForUser")

	cursor, err := repo.GetDataRange(ctx, userID, continuous.Type, status)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	allData := []*glucose.Glucose{}
	for cursor.Next(ctx) {
		allData = slices.Grow(allData, cursor.RemainingBatchLength())
		g := &glucose.Glucose{}
		if err := cursor.Decode(g); err != nil {
			return nil, err
		}
		allData = append(allData, g)
	}

	// The only interesting datapoint for this function, is the newest one.
	sort.Sort(SortByTime(allData))

	return allData[len(allData)-1], nil
}

type SortByTime []*glucose.Glucose

func (s SortByTime) Less(i, j int) bool {
	return s[i].Blood.Time != nil && s[j].Blood.Time != nil && s[i].Blood.Time.Before(*s[j].Blood.Time)
}

func (s SortByTime) Len() int {
	return len(s)
}

func (s SortByTime) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
