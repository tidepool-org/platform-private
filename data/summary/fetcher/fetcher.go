package fetcher

import (
	"context"
	"github.com/tidepool-org/platform/data/summary/types"
	"time"

	"go.mongodb.org/mongo-driver/mongo"

	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types/upload"
)

type DeviceDataCursor interface {
	Decode(val interface{}) error
	RemainingBatchLength() int
	Next(ctx context.Context) bool
	Close(ctx context.Context) error
	GetNextBatch(ctx context.Context) ([]data.Datum, error)
	Err() error
}

type AnyCursor interface {
	Decode(val interface{}) error
	Next(ctx context.Context) bool
	Close(ctx context.Context) error
	Err() error
}

type DeviceDataFetcher interface {
	GetDataSetByID(ctx context.Context, dataSetID string) (*upload.Upload, error)
	GetLastUpdatedForUser(ctx context.Context, userId string, typ []string, lastUpdated time.Time) (*data.UserDataStatus, error)
	GetDataRange(ctx context.Context, userId string, typ []string, status *data.UserDataStatus) (*mongo.Cursor, error)
	DistinctUserIDs(ctx context.Context, typ []string) ([]string, error)
}

type BucketFetcher[B types.BucketDataPt[A], A types.BucketData] interface {
	GetBuckets(ctx context.Context, userId string, startTime, endTime time.Time) (types.BucketsByTime[B, A], error)
	GetAllBuckets(ctx context.Context, userId string) (*mongo.Cursor, error)
	WriteModifiedBuckets(ctx context.Context, startTime time.Time, buckets types.BucketsByTime[B, A]) error
}
