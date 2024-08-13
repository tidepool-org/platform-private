package types

import (
	"time"
)

type GlucoseBin struct {
	Glucose  float64 `json:"glucose,omitempty" bson:"glucose,omitempty"`
	Percent  float64 `json:"percent,omitempty" bson:"percent,omitempty"`
	Variance float64 `json:"variance,omitempty" bson:"variance,omitempty"`

	Minutes int `json:"minutes,omitempty" bson:"minutes,omitempty"`
	Records int `json:"records,omitempty" bson:"records,omitempty"`
}

type GlucoseRangeBins struct {
	Total       GlucoseBin `json:"cgmUse,omitempty" bson:"cgmUse,omitempty"`
	VeryLow     GlucoseBin `json:"inVeryLow,omitempty" bson:"inVeryLow,omitempty"`
	Low         GlucoseBin `json:"inLow,omitempty" bson:"inLow,omitempty"`
	Target      GlucoseBin `json:"inTarget,omitempty" bson:"inTarget,omitempty"`
	High        GlucoseBin `json:"inHigh,omitempty" bson:"inHigh,omitempty"`
	VeryHigh    GlucoseBin `json:"inVeryHigh,omitempty" bson:"inVeryHigh,omitempty"`
	ExtremeHigh GlucoseBin `json:"inExtremeHigh,omitempty" bson:"inExtremeHigh,omitempty"`
	AnyLow      GlucoseBin `json:"inAnyLow,omitempty" bson:"inAnyLow,omitempty"`
	AnyHigh     GlucoseBin `json:"inAnyHigh,omitempty" bson:"inAnyHigh,omitempty"`
}

type GlucoseBucketData struct {
	GlucoseRangeBins
	LastRecordDuration int `json:"lastRecordDuration,omitempty" bson:"lastRecordDuration,omitempty"`
}

type GlucoseBucket struct {
	GlucoseBucketData
	Type           string    `json:"type" bson:"type"`
	Date           time.Time `json:"date" bson:"date"`
	LastRecordTime time.Time `json:"lastRecordTime" bson:"lastRecordTime"`
}

type GlucoseBuckets []*GlucoseBucket

// ContinuousBuckets TODO placeholder for generics testing
type ContinuousBuckets []*GlucoseBucket

type Buckets interface {
	GlucoseBuckets | ContinuousBuckets
}

type BucketsPt[T Buckets] interface {
	*T
	GetType() string
}

//type BucketDataPt[T BucketData] interface {
//	*T
//	CalculateStats(interface{}, *time.Time) (bool, error)
//}

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
