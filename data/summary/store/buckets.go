package store

import (
	"context"
	"errors"
	"fmt"
	"github.com/tidepool-org/platform/data/summary/types"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type Buckets[T types.BucketData] struct {
	*storeStructuredMongo.Repository
	Type string
}

func NewBuckets[T types.BucketData](delegate *storeStructuredMongo.Repository, typ string) *Buckets[T] {
	return &Buckets[T]{
		Repository: delegate,
		Type:       typ,
	}
}

func (r *Buckets[T]) GetBuckets(ctx context.Context, userId string) ([]T, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if userId == "" {
		return nil, errors.New("userId is missing")
	}

	buckets := make([]T, 0)
	selector := bson.M{
		"userId": userId,
		"type":   r.Type,
	}
	opts := options.Find()
	opts.SetSort(bson.D{{Key: "time", Value: 1}})

	cur, err := r.Find(ctx, selector, opts)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("unable to get buckets: %w", err)
	}

	if err = cur.All(ctx, &buckets); err != nil {
		return nil, fmt.Errorf("unable to decode buckets: %w", err)
	}

	return buckets, nil
}

func (r *Buckets[T]) AddBucket(ctx context.Context, bucket types.Bucket[T]) error {
	if ctx == nil {
		return errors.New("context is missing")
	}
	if bucket.UserId == "" {
		return errors.New("userId is missing")
	}
	if bucket.Type == "" {
		return errors.New("type is missing")
	}

	selector := bson.M{
		"userId": bucket.UserId,
		"type":   r.Type,
	}
	opts := options.Replace()
	opts.SetUpsert(true)

	// TODO Bulk insert

	_, err := r.ReplaceOne(ctx, selector, bucket, opts)
	return err
}

func (r *Buckets[T]) ClearInvalidatedBuckets(ctx context.Context, userId string, earliestModified time.Time) (firstData time.Time, err error) {
	selector := bson.M{
		"userId": userId,
		"type":   r.Type,
		"time":   bson.M{"$gt": earliestModified},
	}

	_, err = r.DeleteMany(ctx, selector)
	if err != nil {
		return time.Time{}, err
	}

	return r.GetNewestRecordTime(ctx, userId)
}

func (r *Buckets[T]) GetEnd(ctx context.Context, userId string, side int) (*types.Bucket[T], error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if userId == "" {
		return nil, errors.New("userId is missing")
	}

	buckets := make([]types.Bucket[T], 1)
	selector := bson.M{
		"userId": userId,
		"type":   r.Type,
	}
	opts := options.Find()
	opts.SetSort(bson.D{{Key: "time", Value: side}})
	opts.SetLimit(1)

	cur, err := r.Find(ctx, selector, opts)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("unable to get buckets: %w", err)
	}

	if err = cur.All(ctx, &buckets); err != nil {
		return nil, fmt.Errorf("unable to decode buckets: %w", err)
	}

	return &buckets[0], nil
}

func (r *Buckets[T]) GetNewestRecordTime(ctx context.Context, userId string) (time.Time, error) {
	if ctx == nil {
		return time.Time{}, errors.New("context is missing")
	}
	if userId == "" {
		return time.Time{}, errors.New("userId is missing")
	}

	bucket, err := r.GetEnd(ctx, userId, -1)
	if err != nil {
		return time.Time{}, err
	}

	return bucket.LastTime, nil
}

func (r *Buckets[T]) GetOldestRecordTime(ctx context.Context, userId string) (time.Time, error) {
	if ctx == nil {
		return time.Time{}, errors.New("context is missing")
	}
	if userId == "" {
		return time.Time{}, errors.New("userId is missing")
	}

	bucket, err := r.GetEnd(ctx, userId, 1)
	if err != nil {
		return time.Time{}, err
	}

	return bucket.FirstTime, nil
}

func (r *Buckets[T]) GetTotalHours(ctx context.Context, userId string) (int, error) {
	if ctx == nil {
		return 0, errors.New("context is missing")
	}
	if userId == "" {
		return 0, errors.New("userId is missing")
	}

	firstBucket, err := r.GetEnd(ctx, userId, 1)
	if err != nil {
		return 0, err
	}

	lastBucket, err := r.GetEnd(ctx, userId, 1)
	if err != nil {
		return 0, err
	}

	return int(lastBucket.LastTime.Sub(firstBucket.FirstTime).Hours()), nil
}
