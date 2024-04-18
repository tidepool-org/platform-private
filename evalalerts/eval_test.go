package evalalerts

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/alerts"
	nontypesglucose "github.com/tidepool-org/platform/data/blood/glucose"
	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/data/types/blood"
	"github.com/tidepool-org/platform/data/types/blood/glucose"
	"github.com/tidepool-org/platform/pointer"
)

var _ = Describe("evaluateUrgentLow", func() {
	It("can't function without datum units", func() {
		ctx := context.Background()
		alert := newTestUrgentLowAlert()
		datum := newTestStaticDatumMmolL(11)
		datum.Blood.Units = nil

		_, err := evaluateUrgentLow(ctx, datum, alert)

		Expect(err).To(MatchError("Unable to evaluate datum: Units, Value, or Time is nil"))
	})

	It("can't function without datum value", func() {
		ctx := context.Background()
		alert := newTestUrgentLowAlert()
		datum := newTestStaticDatumMmolL(11)
		datum.Blood.Value = nil

		_, err := evaluateUrgentLow(ctx, datum, alert)

		Expect(err).To(MatchError("Unable to evaluate datum: Units, Value, or Time is nil"))
	})

	It("can't function without datum time", func() {
		ctx := context.Background()
		alert := newTestUrgentLowAlert()
		datum := newTestStaticDatumMmolL(11)
		datum.Blood.Time = nil
		_, err := evaluateUrgentLow(ctx, datum, alert)
		Expect(err).To(MatchError("Unable to evaluate datum: Units, Value, or Time is nil"))
	})

	It("is marked resolved", func() {
		ctx := context.Background()
		alert := newTestUrgentLowAlert()
		datum := newTestStaticDatumMmolL(11)
		alert.Threshold.Value = *datum.Blood.Value - 1
		updated, err := evaluateUrgentLow(ctx, datum, alert)
		Expect(err).To(Succeed())
		Expect(updated).To(BeTrue())
		Expect(alert.Resolved).To(BeTemporally("~", time.Now(), time.Second))
	})

	It("is marked both notified and triggered", func() {
		ctx := context.Background()
		alert := newTestUrgentLowAlert()
		datum := newTestStaticDatumMmolL(11)
		alert.Threshold.Value = *datum.Blood.Value + 1
		updated, err := evaluateUrgentLow(ctx, datum, alert)
		Expect(err).To(Succeed())
		Expect(updated).To(BeTrue())
		Expect(alert.Notified).To(BeTemporally("~", time.Now(), time.Second))
		Expect(alert.Triggered).To(BeTemporally("~", time.Now(), time.Second))
	})

	It("sends notifications regardless of previous notification time", func() {
		ctx := context.Background()
		datum := newTestStaticDatumMmolL(11)
		alert := newTestUrgentLowAlert()
		lastTime := time.Now().Add(-10 * time.Second)
		alert.Activity.Notified = lastTime
		alert.Threshold.Value = *datum.Blood.Value + 1
		updated, err := evaluateUrgentLow(ctx, datum, alert)
		Expect(err).To(Succeed())
		Expect(updated).To(BeTrue())
		Expect(alert.Notified).To(BeTemporally("~", time.Now(), time.Second))
	})
})

func newTestStaticDatumMmolL(value float64) *glucose.Glucose {
	return &glucose.Glucose{
		Blood: blood.Blood{
			Base: types.Base{
				Time: pointer.FromTime(time.Now()),
			},
			Units: pointer.FromString(nontypesglucose.MmolL),
			Value: pointer.FromFloat64(value),
		},
	}
}

func newTestUrgentLowAlert() *alerts.UrgentLowAlert {
	return &alerts.UrgentLowAlert{
		Base: alerts.Base{
			Enabled:  true,
			Activity: alerts.Activity{},
		},
		Threshold: alerts.Threshold{
			Units: nontypesglucose.MmolL,
		},
	}
}
