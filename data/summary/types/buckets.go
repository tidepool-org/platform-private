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

func (B *GlucoseBucket) finalizeMinutes(wallMinutes float64) {
	B.Total.Percent = float64(B.Total.Minutes) / wallMinutes

	if (wallMinutes <= minutesPerDay && B.Total.Percent > 0.7) || (wallMinutes > minutesPerDay && B.Total.Minutes > minutesPerDay) {
		B.VeryLow.Percent = float64(B.VeryLow.Minutes) / wallMinutes
		B.Low.Percent = float64(B.Low.Minutes) / wallMinutes
		B.Target.Percent = float64(B.Target.Minutes) / wallMinutes
		B.High.Percent = float64(B.High.Minutes) / wallMinutes
		B.VeryHigh.Percent = float64(B.VeryHigh.Minutes) / wallMinutes
		B.ExtremeHigh.Percent = float64(B.ExtremeHigh.Minutes) / wallMinutes
		B.AnyLow.Percent = float64(B.VeryLow.Minutes+B.Low.Minutes) / wallMinutes
		B.AnyHigh.Percent = float64(B.VeryHigh.Minutes+B.High.Minutes) / wallMinutes
	}
}

func (B *GlucoseBucket) finalizeRecords() {
	B.Total.Percent = float64(B.Total.Minutes) / float64(B.Total.Records)
	B.VeryLow.Percent = float64(B.VeryLow.Minutes) / float64(B.Total.Records)
	B.Low.Percent = float64(B.Low.Minutes) / float64(B.Total.Records)
	B.Target.Percent = float64(B.Target.Minutes) / float64(B.Total.Records)
	B.High.Percent = float64(B.High.Minutes) / float64(B.Total.Records)
	B.VeryHigh.Percent = float64(B.VeryHigh.Minutes) / float64(B.Total.Records)
	B.ExtremeHigh.Percent = float64(B.ExtremeHigh.Minutes) / float64(B.Total.Records)
	B.AnyLow.Percent = float64(B.VeryLow.Minutes+B.Low.Minutes) / float64(B.Total.Records)
	B.AnyHigh.Percent = float64(B.VeryHigh.Minutes+B.High.Minutes) / float64(B.Total.Records)
}

func (B *GlucoseBucket) Finalize(shared *BucketShared) {
	if B.Total.Minutes != 0 {
		// if our bucket (period, at this point) has minutes
		wallMinutes := shared.LastData.Sub(shared.FirstData).Minutes()
		B.finalizeMinutes(wallMinutes)
	} else {
		// otherwise, we only have record counts
		B.finalizeRecords()
	}
}

type BucketShared struct {
	UserId    string    `json:"userId" bson:"userId"`
	Type      string    `json:"type" bson:"type"`
	Day       time.Time `json:"day" bson:"day"`
	FirstData time.Time `json:"firstTime" bson:"firstTime"`
	LastData  time.Time `json:"lastTime" bson:"lastTime"`
}

type Bucket[B BucketDataPt[A], A BucketData] struct {
	BucketShared
	Data B `json:"data" bson:"data"`
}

type BucketData interface {
	GlucoseBucket | ContinuousBucketData
}

type BucketDataPt[A BucketData] interface {
	*A
	Add(new *GlucoseBucket)
	Finalize(bucket *BucketShared)
	//CalculateStats(interface{}, *time.Time) (bool, error)
}

// ContinuousBucket TODO placeholder for generics testing
type ContinuousBucket GlucoseBucket

//type Bucket[A BucketDataPt[T], T BucketData] struct {
//	Date           time.Time `json:"date" bson:"date"`
//	LastRecordTime time.Time `json:"lastRecordTime" bson:"lastRecordTime"`
//	Type           string    `json:"type" bson:"type"`
//
//	Data A `json:"data" bson:"data"`
//}

//func CreateBucket[A BucketDataPt[T], T BucketData](t time.Time) *Bucket[A, T] {
//	bucket := new(Bucket[A, T])
//	bucket.Date = t
//	bucket.Data = new(T)
//	return bucket
//}
