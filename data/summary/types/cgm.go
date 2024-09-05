package types

import (
	"context"
	"errors"
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/summary/fetcher"
	"github.com/tidepool-org/platform/data/types/blood/glucose/continuous"
	"strconv"
	"time"
)

type GlucosePeriods map[string]*GlucosePeriod

type CGMStats struct {
	Periods       GlucosePeriods `json:"periods,omitempty" bson:"periods,omitempty"`
	OffsetPeriods GlucosePeriods `json:"offsetPeriods,omitempty" bson:"offsetPeriods,omitempty"`
	TotalHours    int            `json:"totalHours,omitempty" bson:"totalHours,omitempty"`
}

func (*CGMStats) GetType() string {
	return SummaryTypeCGM
}

func (*CGMStats) GetDeviceDataTypes() []string {
	return []string{continuous.Type}
}

func (s *CGMStats) Init() {
	s.Periods = make(map[string]*GlucosePeriod)
	s.OffsetPeriods = make(map[string]*GlucosePeriod)
	s.TotalHours = 0
}

func (s *CGMStats) Update(ctx context.Context, shared SummaryShared, bucketsFetcher fetcher.BucketFetcher[*GlucoseBucket, GlucoseBucket], cursor fetcher.DeviceDataCursor) error {
	// move all of this to a generic method? fetcher interface?

	hasMoreData := true
	var buckets BucketsByTime[*GlucoseBucket, GlucoseBucket]
	var err error
	var userData []data.Datum
	var startTime time.Time
	var endTime time.Time

	for hasMoreData {
		userData, err = cursor.GetNextBatch(ctx)
		if errors.Is(err, fetcher.ErrCursorExhausted) {
			hasMoreData = false
		} else if err != nil {
			return err
		}

		if len(userData) > 0 {
			startTime = userData[0].GetTime().UTC().Truncate(time.Hour)
			endTime = userData[len(userData)].GetTime().UTC().Truncate(time.Hour)
			buckets, err = bucketsFetcher.GetBuckets(ctx, shared.UserID, startTime, endTime)
			if err != nil {
				return err
			}

			err = buckets.Update(shared.UserID, shared.Type, userData)
			if err != nil {
				return err
			}

			err = bucketsFetcher.WriteModifiedBuckets(ctx, shared.Dates.LastUpdatedDate, buckets)
			if err != nil {
				return err
			}

		}
	}

	s.CalculateSummary()
	s.CalculateDelta()

	return nil
}

func (s *CGMStats) CalculateSummary() {
	// count backwards (newest first) through hourly stats, stopping at 24, 24*7, 24*14, 24*30
	// currently only supports day precision
	nextStopPoint := 0
	nextOffsetStopPoint := 0
	totalStats := 0
	totalOffsetStats := GlucosePeriod{}
	dayCounted := false
	offsetDayCounted := false

	for i := 1; i <= len(s.Buckets); i++ {
		currentIndex := len(s.Buckets) - i

		// only add to offset stats when primary stop point is ahead of offset
		if nextStopPoint > nextOffsetStopPoint && len(stopPoints) > nextOffsetStopPoint {
			totalOffsetStats.Add(s.Buckets[currentIndex].Data)

			if s.Buckets[currentIndex].Data.Total.Records > 0 {
				totalOffsetStats.HoursWithData++

				if !offsetDayCounted {
					totalOffsetStats.DaysWithData++
					offsetDayCounted = true
				}
			}

			// new day, reset day counting flag
			if i%24 == 0 {
				offsetDayCounted = false
			}

			if i == stopPoints[nextOffsetStopPoint]*24*2 {
				s.CalculatePeriod(stopPoints[nextOffsetStopPoint], true, totalOffsetStats)
				nextOffsetStopPoint++
				totalOffsetStats = GlucosePeriod{}
			}
		}

		// only count primary stats when the next stop point is a real period
		if len(stopPoints) > nextStopPoint {
			totalStats.Add(s.Buckets[currentIndex].Data)

			if s.Buckets[currentIndex].Data.Total.Records > 0 {
				totalStats.HoursWithData++

				if !dayCounted {
					totalStats.DaysWithData++
					dayCounted = true
				}
			}

			// end of day, reset day counting flag
			if i > 0 && i%24 == 0 {
				dayCounted = false
			}

			if i == stopPoints[nextStopPoint]*24 {
				s.CalculatePeriod(stopPoints[nextStopPoint], false, totalStats)
				nextStopPoint++
			}
		}
	}

	// fill in periods we never reached
	for i := nextStopPoint; i < len(stopPoints); i++ {
		s.CalculatePeriod(stopPoints[i], false, totalStats)
	}
	for i := nextOffsetStopPoint; i < len(stopPoints); i++ {
		s.CalculatePeriod(stopPoints[i], true, totalOffsetStats)
		totalOffsetStats = GlucosePeriod{}
	}

	s.TotalHours = len(s.Buckets)
}

func (s *CGMStats) CalculateDelta() {
	for k := range s.Periods {
		BinDelta(&s.Periods[k].Total, &s.OffsetPeriods[k].Total, &s.Periods[k].Delta.Total, &s.OffsetPeriods[k].Delta.Total)
		BinDelta(&s.Periods[k].VeryLow, &s.OffsetPeriods[k].VeryLow, &s.Periods[k].Delta.VeryLow, &s.OffsetPeriods[k].Delta.VeryLow)
		BinDelta(&s.Periods[k].Low, &s.OffsetPeriods[k].Low, &s.Periods[k].Delta.Low, &s.OffsetPeriods[k].Delta.Low)
		BinDelta(&s.Periods[k].Target, &s.OffsetPeriods[k].Target, &s.Periods[k].Delta.Target, &s.OffsetPeriods[k].Delta.Target)
		BinDelta(&s.Periods[k].High, &s.OffsetPeriods[k].High, &s.Periods[k].Delta.High, &s.OffsetPeriods[k].Delta.High)
		BinDelta(&s.Periods[k].VeryHigh, &s.OffsetPeriods[k].VeryHigh, &s.Periods[k].Delta.VeryHigh, &s.OffsetPeriods[k].Delta.VeryHigh)
		BinDelta(&s.Periods[k].ExtremeHigh, &s.OffsetPeriods[k].ExtremeHigh, &s.Periods[k].Delta.ExtremeHigh, &s.OffsetPeriods[k].Delta.ExtremeHigh)
		BinDelta(&s.Periods[k].AnyLow, &s.OffsetPeriods[k].AnyLow, &s.Periods[k].Delta.AnyLow, &s.OffsetPeriods[k].Delta.AnyLow)
		BinDelta(&s.Periods[k].AnyHigh, &s.OffsetPeriods[k].AnyHigh, &s.Periods[k].Delta.AnyHigh, &s.OffsetPeriods[k].Delta.AnyHigh)

		Delta(&s.Periods[k].AverageGlucose, &s.OffsetPeriods[k].AverageGlucose, &s.Periods[k].Delta.AverageGlucose, &s.OffsetPeriods[k].Delta.AverageGlucose)
		Delta(&s.Periods[k].GlucoseManagementIndicator, &s.OffsetPeriods[k].GlucoseManagementIndicator, &s.OffsetPeriods[k].Delta.GlucoseManagementIndicator, &s.Periods[k].Delta.GlucoseManagementIndicator)
		Delta(&s.Periods[k].AverageDailyRecords, &s.OffsetPeriods[k].AverageDailyRecords, &s.OffsetPeriods[k].Delta.AverageDailyRecords, &s.Periods[k].Delta.AverageDailyRecords)
		Delta(&s.Periods[k].StandardDeviation, &s.OffsetPeriods[k].StandardDeviation, &s.Periods[k].Delta.StandardDeviation, &s.OffsetPeriods[k].Delta.StandardDeviation)
		Delta(&s.Periods[k].CoefficientOfVariation, &s.OffsetPeriods[k].CoefficientOfVariation, &s.Periods[k].Delta.CoefficientOfVariation, &s.OffsetPeriods[k].Delta.CoefficientOfVariation)
		Delta(&s.Periods[k].DaysWithData, &s.OffsetPeriods[k].DaysWithData, &s.Periods[k].Delta.DaysWithData, &s.OffsetPeriods[k].Delta.DaysWithData)
		Delta(&s.Periods[k].HoursWithData, &s.OffsetPeriods[k].HoursWithData, &s.Periods[k].Delta.HoursWithData, &s.OffsetPeriods[k].Delta.HoursWithData)
	}
}

func (s *CGMStats) CalculatePeriod(i int, offset bool, period GlucosePeriod) {
	// We don't make a copy of period, as the struct has no pointers... right? you didn't add any pointers right?
	period.Finalize(i)

	if offset {
		s.OffsetPeriods[strconv.Itoa(i)+"d"] = &period
	} else {
		s.Periods[strconv.Itoa(i)+"d"] = &period
	}

}
