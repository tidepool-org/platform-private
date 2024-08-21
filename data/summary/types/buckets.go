package types

import (
	"math"
	"time"
)

const minutesPerDay = 60 * 24

type GlucoseRange struct {
	Glucose  float64 `json:"glucose,omitempty" bson:"glucose,omitempty"`
	Percent  float64 `json:"percent,omitempty" bson:"percent,omitempty"`
	Variance float64 `json:"variance,omitempty" bson:"variance,omitempty"`

	Minutes int `json:"minutes,omitempty" bson:"minutes,omitempty"`
	Records int `json:"records,omitempty" bson:"records,omitempty"`
}

func (R *GlucoseRange) Add(new *GlucoseRange) {
	R.Variance = R.CombineVariance(new)
	R.Glucose += new.Glucose
	R.Minutes += new.Minutes
	R.Records += new.Records
	// We skip percent here as it has to be calculated relative to other ranges
}

// CombineVariance Implemented using https://en.wikipedia.org/wiki/Algorithms_for_calculating_variance#Parallel_algorithm
func (R *GlucoseRange) CombineVariance(new *GlucoseRange) float64 {
	// Exit early for No-Op case
	if R.Variance == 0 && new.Variance == 0 {
		return 0
	}

	// Return new if existing is 0
	if R.Variance == 0 {
		return new.Variance
	}

	// if we have no values in either bucket, this will result in NaN, and cant be added anyway, return what we started with
	if R.Minutes == 0 || new.Minutes == 0 {
		return R.Variance
	}

	n1 := float64(R.Minutes)
	n2 := float64(new.Minutes)
	n := n1 + n2
	delta := new.Glucose/n2 - R.Glucose/n1
	return R.Variance + new.Variance + math.Pow(delta, 2)*n1*n2/n
}

// CalculateVariance Implemented using https://en.wikipedia.org/wiki/Algorithms_for_calculating_variance#Weighted_incremental_algorithm
func (R *GlucoseRange) CalculateVariance(value float64, duration float64) float64 {
	var mean float64 = 0
	if R.Minutes > 0 {
		mean = R.Glucose / float64(R.Minutes)
	}

	weight := float64(R.Minutes) + duration
	newMean := mean + (duration/weight)*(value-mean)
	return R.Variance + duration*(value-mean)*(value-newMean)
}

type GlucoseRanges struct {
	Total       GlucoseRange `json:"cgmUse,omitempty" bson:"cgmUse,omitempty"`
	VeryLow     GlucoseRange `json:"inVeryLow,omitempty" bson:"inVeryLow,omitempty"`
	Low         GlucoseRange `json:"inLow,omitempty" bson:"inLow,omitempty"`
	Target      GlucoseRange `json:"inTarget,omitempty" bson:"inTarget,omitempty"`
	High        GlucoseRange `json:"inHigh,omitempty" bson:"inHigh,omitempty"`
	VeryHigh    GlucoseRange `json:"inVeryHigh,omitempty" bson:"inVeryHigh,omitempty"`
	ExtremeHigh GlucoseRange `json:"inExtremeHigh,omitempty" bson:"inExtremeHigh,omitempty"`
	AnyLow      GlucoseRange `json:"inAnyLow,omitempty" bson:"inAnyLow,omitempty"`
	AnyHigh     GlucoseRange `json:"inAnyHigh,omitempty" bson:"inAnyHigh,omitempty"`
}

func (R *GlucoseRanges) Add(new *GlucoseRanges) {
	R.Total.Add(&new.Total)
	R.VeryLow.Add(&new.VeryLow)
	R.Low.Add(&new.Low)
	R.Target.Add(&new.Target)
	R.High.Add(&new.High)
	R.VeryHigh.Add(&new.VeryHigh)
	R.ExtremeHigh.Add(&new.ExtremeHigh)
	R.AnyLow.Add(&new.AnyLow)
	R.AnyHigh.Add(&new.AnyHigh)
}

type GlucoseBucket struct {
	GlucoseRanges
	LastRecordDuration int `json:"lastRecordDuration,omitempty" bson:"lastRecordDuration,omitempty"`
}

func (R *GlucoseRanges) finalizeMinutes(wallMinutes float64) {
	R.Total.Percent = float64(R.Total.Minutes) / wallMinutes

	if (wallMinutes <= minutesPerDay && R.Total.Percent > 0.7) || (wallMinutes > minutesPerDay && R.Total.Minutes > minutesPerDay) {
		R.VeryLow.Percent = float64(R.VeryLow.Minutes) / wallMinutes
		R.Low.Percent = float64(R.Low.Minutes) / wallMinutes
		R.Target.Percent = float64(R.Target.Minutes) / wallMinutes
		R.High.Percent = float64(R.High.Minutes) / wallMinutes
		R.VeryHigh.Percent = float64(R.VeryHigh.Minutes) / wallMinutes
		R.ExtremeHigh.Percent = float64(R.ExtremeHigh.Minutes) / wallMinutes
		R.AnyLow.Percent = float64(R.VeryLow.Minutes+R.Low.Minutes) / wallMinutes
		R.AnyHigh.Percent = float64(R.VeryHigh.Minutes+R.High.Minutes) / wallMinutes
	}
}

func (R *GlucoseRanges) finalizeRecords() {
	R.Total.Percent = float64(R.Total.Minutes) / float64(R.Total.Records)
	R.VeryLow.Percent = float64(R.VeryLow.Minutes) / float64(R.Total.Records)
	R.Low.Percent = float64(R.Low.Minutes) / float64(R.Total.Records)
	R.Target.Percent = float64(R.Target.Minutes) / float64(R.Total.Records)
	R.High.Percent = float64(R.High.Minutes) / float64(R.Total.Records)
	R.VeryHigh.Percent = float64(R.VeryHigh.Minutes) / float64(R.Total.Records)
	R.ExtremeHigh.Percent = float64(R.ExtremeHigh.Minutes) / float64(R.Total.Records)
	R.AnyLow.Percent = float64(R.VeryLow.Minutes+R.Low.Minutes) / float64(R.Total.Records)
	R.AnyHigh.Percent = float64(R.VeryHigh.Minutes+R.High.Minutes) / float64(R.Total.Records)
}

func (R *GlucoseRanges) Finalize(shared *BucketShared) {
	if R.Total.Minutes != 0 {
		// if our bucket (period, at this point) has minutes
		wallMinutes := shared.LastData.Sub(shared.FirstData).Minutes()
		R.finalizeMinutes(wallMinutes)
	} else {
		// otherwise, we only have record counts
		R.finalizeRecords()
	}
}

func (B *GlucoseBucket) Add(bucket *GlucoseBucket) {
	B.Add(bucket)
}

type BucketShared struct {
	UserId    string    `json:"userId" bson:"userId"`
	Type      string    `json:"type" bson:"type"`
	Time      time.Time `json:"time" bson:"time"`
	FirstData time.Time `json:"firstTime" bson:"firstTime"`
	LastData  time.Time `json:"lastTime" bson:"lastTime"`
}

func (BS *BucketShared) Add(shared *BucketShared) {
	if shared.FirstData.Before(BS.FirstData) {
		BS.FirstData = shared.FirstData
	}

	if shared.LastData.After(BS.LastData) {
		BS.LastData = shared.LastData
	}
}

type Bucket[B BucketDataPt[A], A BucketData] struct {
	BucketShared
	Data B `json:"data" bson:"data"`
}

func (B *Bucket[B, A]) Add(bucket *Bucket[B, A]) {
	B.Data.Add(bucket.Data)
	B.BucketShared.Add(&bucket.BucketShared)
}

type BucketData interface {
	GlucoseBucket | ContinuousBucketData
}

type BucketDataPt[A BucketData] interface {
	*A
	Add(bucket *A)
}

// ContinuousBucket TODO placeholder for generics testing
type ContinuousBucket GlucoseBucket

type GlucosePeriod struct {
	GlucoseRanges
	state BucketShared

	HoursWithData int `json:"hoursWithData,omitempty" bson:"hoursWithData,omitempty"`
	DaysWithData  int `json:"daysWithData,omitempty" bson:"daysWithData,omitempty"`

	AverageGlucose             float64 `json:"averageGlucoseMmol,omitempty" bson:"avgGlucose,omitempty"`
	GlucoseManagementIndicator float64 `json:"glucoseManagementIndicator,omitempty" bson:"GMI,omitempty"`

	CoefficientOfVariation float64 `json:"coefficientOfVariation,omitempty" bson:"CV,omitempty"`
	StandardDeviation      float64 `json:"standardDeviation,omitempty" bson:"SD,omitempty"`

	AverageDailyRecords float64 `json:"averageDailyRecords,omitempty" b;son:"avgDailyRecords,omitempty,omitempty"`

	Delta *GlucosePeriod `json:"delta,omitempty" bson:"delta,omitempty"`
}

func (P GlucosePeriod) Finalize(days int) {
	P.GlucoseRanges.Finalize(&P.state)
	P.AverageGlucose = P.Total.Glucose / float64(P.Total.Minutes)

	// we only add GMI if cgm use >70%, otherwise clear it
	if P.Total.Percent > 0.7 {
		P.GlucoseManagementIndicator = CalculateGMI(P.AverageGlucose)
	}

	P.StandardDeviation = math.Sqrt(P.Total.Variance / float64(P.Total.Minutes))
	P.CoefficientOfVariation = P.StandardDeviation / P.AverageGlucose

	P.AverageDailyRecords = float64(P.Total.Records) / float64(days)

	// TODO HoursWithData
	// TODO DaysWithData
	// can it be centralized here? does it have to be in the iteration?
}

// TODO standardize everything into Add and Finalize at every level
