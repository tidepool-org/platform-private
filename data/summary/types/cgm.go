package types

import (
	"context"
	"errors"
	"math"
	"strconv"
	"time"

	"github.com/tidepool-org/platform/data/blood/glucose"
	"github.com/tidepool-org/platform/data/summary/fetcher"
	glucoseDatum "github.com/tidepool-org/platform/data/types/blood/glucose"
	"github.com/tidepool-org/platform/data/types/blood/glucose/continuous"
	"github.com/tidepool-org/platform/pointer"
)

type GlucosePeriod struct {
	GlucoseRanges
	HoursWithData int `json:"hoursWithData,omitempty" bson:"hoursWithData,omitempty"`
	DaysWithData  int `json:"daysWithData,omitempty" bson:"daysWithData,omitempty"`

	AverageGlucose             float64 `json:"averageGlucoseMmol,omitempty" bson:"avgGlucose,omitempty"`
	GlucoseManagementIndicator float64 `json:"glucoseManagementIndicator,omitempty" bson:"GMI,omitempty"`

	CoefficientOfVariation float64 `json:"coefficientOfVariation,omitempty" bson:"CV,omitempty"`
	StandardDeviation      float64 `json:"standardDeviation,omitempty" bson:"SD,omitempty"`

	AverageDailyRecords float64 `json:"averageDailyRecords,omitempty" bson:"avgDailyRecords,omitempty,omitempty"`

	Delta *GlucosePeriod `json:"delta,omitempty" bson:"delta,omitempty"`
}

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

func (s *CGMStats) Update(ctx context.Context, cursor fetcher.DeviceDataCursor) error {
	hasMoreData := true
	for hasMoreData {
		userData, err := cursor.GetNextBatch(ctx)
		if errors.Is(err, fetcher.ErrCursorExhausted) {
			hasMoreData = false
		} else if err != nil {
			return err
		}

		if len(userData) > 0 {
			err = AddData(&s.Buckets, userData)
			if err != nil {
				return err
			}
		}
	}

	s.CalculateSummary()

	return nil
}

func (B *GlucoseBucket) CalculateStats(r any, lastRecordTime *time.Time) (bool, error) {
	dataRecord, ok := r.(*glucoseDatum.Glucose)
	if !ok {
		return false, errors.New("CGM record for calculation is not compatible with Glucose type")
	}

	// this is a new bucket, use current record as duration reference
	if B.LastRecordDuration == 0 {
		B.LastRecordDuration = GetDuration(dataRecord)
	}

	// calculate blackoutWindow based on duration of previous value
	blackoutWindow := time.Duration(B.LastRecordDuration)*time.Minute - 10*time.Second

	// Skip record unless we are beyond the blackout window
	if dataRecord.Time.Sub(*lastRecordTime) > blackoutWindow {
		normalizedValue := *glucose.NormalizeValueForUnits(dataRecord.Value, pointer.FromAny(glucose.MmolL))
		duration := GetDuration(dataRecord)

		if normalizedValue < veryLowBloodGlucose {
			B.VeryLow.Minutes += duration
			B.VeryLow.Records++
		} else if normalizedValue > veryHighBloodGlucose {
			B.VeryHigh.Minutes += duration
			B.VeryHigh.Records++

			// VeryHigh is inclusive of extreme high, this is intentional
			if normalizedValue >= extremeHighBloodGlucose {
				B.ExtremeHigh.Minutes += duration
				B.ExtremeHigh.Records++
			}
		} else if normalizedValue < lowBloodGlucose {
			B.Low.Minutes += duration
			B.Low.Records++
		} else if normalizedValue > highBloodGlucose {
			B.High.Minutes += duration
			B.High.Records++
		} else {
			B.Target.Minutes += duration
			B.Target.Records++
		}

		// this must occur before the counters below as the pre-increment counters are used during calc
		B.Total.Variance = B.Total.CalculateVariance(normalizedValue, float64(duration))

		B.Total.Minutes += duration
		B.Total.Records++
		B.Total.Glucose += normalizedValue * float64(duration)
		B.LastRecordDuration = duration

		return false, nil
	}

	return true, nil
}

func (s *CGMStats) CalculateSummary() {
	// count backwards (newest first) through hourly stats, stopping at 24, 24*7, 24*14, 24*30
	// currently only supports day precision
	nextStopPoint := 0
	nextOffsetStopPoint := 0
	totalStats := GlucosePeriod{}
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

	s.CalculateDelta()
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
	if period.Total.Records != 0 {
		realMinutes := CalculateWallMinutes(i, s.Buckets[len(s.Buckets)-1].LastRecordTime, s.Buckets[len(s.Buckets)-1].Data.LastRecordDuration)
		period.Total.Percent = float64(period.Total.Minutes) / realMinutes

		// if we are storing under 1d, apply 70% rule
		// if we are storing over 1d, check for 24h cgm use
		if (i <= 1 && period.Total.Percent > 0.7) || (i > 1 && period.Total.Minutes > 1440) {
			period.Target.Percent = float64(period.Target.Minutes) / float64(period.Total.Minutes)
			period.Low.Percent = float64(period.Low.Minutes) / float64(period.Total.Minutes)
			period.VeryLow.Percent = float64(period.VeryLow.Minutes) / float64(period.Total.Minutes)
			period.AnyLow.Percent = float64(period.VeryLow.Records+period.Low.Records) / float64(period.Total.Records)
			period.High.Percent = float64(period.High.Minutes) / float64(period.Total.Minutes)
			period.VeryHigh.Percent = float64(period.VeryHigh.Minutes) / float64(period.Total.Minutes)
			period.ExtremeHigh.Percent = float64(period.ExtremeHigh.Minutes) / float64(period.Total.Minutes)
			period.AnyHigh.Percent = float64(period.VeryHigh.Records+period.High.Records) / float64(period.Total.Records)
		}

		period.AverageGlucose = period.Total.Glucose / float64(period.Total.Minutes)

		// we only add GMI if cgm use >70%, otherwise clear it
		if period.Total.Percent > 0.7 {
			period.GlucoseManagementIndicator = CalculateGMI(period.AverageGlucose)
		}

		period.StandardDeviation = math.Sqrt(period.Total.Variance / float64(period.Total.Minutes))
		period.CoefficientOfVariation = period.StandardDeviation / period.AverageGlucose
	}

	if offset {
		s.OffsetPeriods[strconv.Itoa(i)+"d"] = &period
	} else {
		s.Periods[strconv.Itoa(i)+"d"] = &period
	}

}
