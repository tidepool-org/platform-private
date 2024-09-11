package types

import (
	"github.com/tidepool-org/platform/data"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

const minutesPerDay = 60 * 24

type BucketShared struct {
	ID        primitive.ObjectID `json:"-" bson:"_id,omitempty"`
	UserId    string             `json:"userId" bson:"userId"`
	Type      string             `json:"type" bson:"type"`
	Time      time.Time          `json:"time" bson:"time"`
	FirstData time.Time          `json:"firstTime" bson:"firstTime"`
	LastData  time.Time          `json:"lastTime" bson:"lastTime"`

	modified bool
}

func (BS *BucketShared) SetModified() {
	BS.modified = true
}

func (BS *BucketShared) IsModified() bool {
	return BS.modified
}

type BucketData interface {
	GlucoseBucket | ContinuousBucket
}

type BucketDataPt[A BucketData] interface {
	*A
	Add(bucket *A)
	Update(record data.Datum, shared *BucketShared) error
}

type Bucket[B BucketDataPt[A], A BucketData] struct {
	BucketShared
	Data B `json:"data" bson:"data"`
}

func NewBucket[B BucketDataPt[A], A BucketData](userId string, date time.Time, typ string) *Bucket[B, A] {
	return &Bucket[B, A]{
		BucketShared: BucketShared{
			UserId: userId,
			Type:   typ,
			Time:   date,
		},
		Data: new(A),
	}
}

func (B *Bucket[B, A]) Update(record data.Datum) error {
	B.SetModified()
	return B.Data.Update(record, &B.BucketShared)
}

type BucketsByTime[B BucketDataPt[A], A BucketData] map[time.Time]*Bucket[B, A]

func (BT BucketsByTime[B, A]) Update(userId string, typ string, userData []data.Datum) error {
	for _, r := range userData {
		// truncate time is not timezone/DST safe here, even if we do expect UTC, never truncate non-utc
		recordHour := r.GetTime().UTC().Truncate(time.Hour)

		// OPTIMIZATION this could check if recordHour equal to previous hour, to save a map lookup, probably saves 0ms
		if _, ok := BT[recordHour]; !ok {
			// we don't already have a bucket for this data
			BT[recordHour] = NewBucket[B](userId, recordHour, typ)

			// fresh bucket, pull LastData from previous hour if possible for dedup
			if _, ok = BT[recordHour.Add(-time.Hour)]; ok {
				BT[recordHour].BucketShared.LastData = BT[recordHour.Add(-time.Hour)].LastData
			}
		}

		err := BT[recordHour].Update(r)
		if err != nil {
			return err
		}

	}

	return nil
}
