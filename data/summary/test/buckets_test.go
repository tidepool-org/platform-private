package test_test

import (
	"github.com/tidepool-org/platform/data/types/blood/glucose/continuous"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/data"
	. "github.com/tidepool-org/platform/data/summary/test/generators"
	"github.com/tidepool-org/platform/data/summary/types"
	userTest "github.com/tidepool-org/platform/user/test"
)

var _ = Describe("Buckets", func() {
	var bucketTime time.Time
	var err error
	var userId string

	BeforeEach(func() {
		now := time.Now()
		userId = userTest.RandomID()
		bucketTime = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	})

	Context("Adding Data records to a bucket", func() {
		var userBucket *types.Bucket[*types.GlucoseBucket, types.GlucoseBucket]
		var cgmDatum data.Datum

		It("With a fresh bucket", func() {
			datumTime := bucketTime.Add(5 * time.Minute)
			userBucket = types.NewBucket[*types.GlucoseBucket](userId, datumTime, types.SummaryTypeCGM)
			cgmDatum = NewGlucoseWithValue(continuous.Type, datumTime, InTargetBloodGlucose)

			err = userBucket.Update(cgmDatum)
			Expect(err).ToNot(HaveOccurred())

			Expect(userBucket.LastData).To(Equal(datumTime))
			Expect(userBucket.FirstData).To(Equal(datumTime))
			Expect(userBucket.Time).To(Equal(bucketTime))
			Expect(userBucket.Data.Target.Records).To(Equal(1))
			Expect(userBucket.Data.Target.Glucose).To(Equal(InTargetBloodGlucose))
			Expect(userBucket.Data.Target.Minutes).To(Equal(5))
			Expect(userBucket.IsModified()).To(BeTrue())
		})

		//It("With no existing buckets", func() {
		//	userBuckets = &types.BucketsByTime[*types.GlucoseBucket, types.GlucoseBucket]{}
		//	cgmDatum = NewGlucoseWithValue(types.SummaryTypeCGM, datumTime)
		//
		//	userBuckets.Update(userId,)
		//
		//	Expect(err).ToNot(HaveOccurred())
		//})

	})
})
