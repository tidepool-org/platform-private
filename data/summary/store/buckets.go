package store

import (
	"context"
	"errors"
	"github.com/tidepool-org/platform/data/summary/types"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
	"go.mongodb.org/mongo-driver/bson"
)

type Buckets[A types.BucketsPt[T], T types.Buckets] struct {
	*storeStructuredMongo.Repository
}

func NewBuckets[A types.BucketsPt[T], T types.Buckets](delegate *storeStructuredMongo.Repository) *Buckets[A, T] {
	return &Buckets[A, T]{
		delegate,
	}
}

func (r *Buckets[A, T]) GetBuckets(ctx context.Context, userId string) (A, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if userId == "" {
		return nil, errors.New("userId is missing")
	}

	buckets := make(T, 0)
	selector := bson.M{
		"userId": userId,
		"type":   buckets.GetType(),
	}

	//err := r.FindOne(ctx, selector).Decode(&summary)
	//if errors.Is(err, mongo.ErrNoDocuments) {
	//	return nil, nil
	//} else if err != nil {
	//	return nil, fmt.Errorf("unable to get summary: %w", err)
	//}

	return buckets, nil
}
