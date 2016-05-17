package bolus_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	fixtures "github.com/tidepool-org/platform/data/_fixtures"
	"github.com/tidepool-org/platform/data/types"
	"github.com/tidepool-org/platform/data/types/bolus"
)

var _ = Describe("Square", func() {

	var bolusObj = fixtures.TestingDatumBase()
	bolusObj["type"] = "bolus"
	bolusObj["subType"] = "square"
	bolusObj["extended"] = 1.0
	bolusObj["duration"] = 3600000

	var helper *types.TestingHelper

	BeforeEach(func() {
		helper = types.NewTestingHelper()
	})

	Context("from obj", func() {

		It("if the obj is valid", func() {
			Expect(helper.ValidDataType(bolus.Build(bolusObj, helper.ErrorProcessing))).To(BeNil())
		})

		Context("validation", func() {

			Context("duration", func() {

				It("is required", func() {
					delete(bolusObj, "duration")

					Expect(
						helper.ErrorIsExpected(
							bolus.Build(bolusObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/duration",
								Detail: "Must be greater than 0 and less than 86400000 given '<nil>'",
							}),
					).To(BeNil())
				})

				It("invalid when less than zero", func() {
					bolusObj["duration"] = -1

					Expect(
						helper.ErrorIsExpected(
							bolus.Build(bolusObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/duration",
								Detail: "Must be greater than 0 and less than 86400000 given '-1'",
							}),
					).To(BeNil())

				})

				It("invalid when over 86400000", func() {
					bolusObj["duration"] = 86400001

					Expect(
						helper.ErrorIsExpected(
							bolus.Build(bolusObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/duration",
								Detail: "Must be greater than 0 and less than 86400000 given '86400001'",
							}),
					).To(BeNil())

				})

				It("valid greater than zero", func() {
					bolusObj["duration"] = 4000
					Expect(helper.ValidDataType(bolus.Build(bolusObj, helper.ErrorProcessing))).To(BeNil())
				})

			})
			Context("extended", func() {

				It("is required", func() {
					delete(bolusObj, "extended")

					Expect(
						helper.ErrorIsExpected(
							bolus.Build(bolusObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/extended",
								Detail: "Must be greater than 0 and less than or equal to 100.0 given '<nil>'",
							}),
					).To(BeNil())

				})

				It("invalid when zero", func() {
					bolusObj["extended"] = 0.0

					Expect(
						helper.ErrorIsExpected(
							bolus.Build(bolusObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/extended",
								Detail: "Must be greater than 0 and less than or equal to 100.0 given '0'",
							}),
					).To(BeNil())

				})

				It("invalid when greater than 100", func() {
					bolusObj["extended"] = 105.0

					Expect(
						helper.ErrorIsExpected(
							bolus.Build(bolusObj, helper.ErrorProcessing),
							types.ExpectedErrorDetails{
								Path:   "0/extended",
								Detail: "Must be greater than 0 and less than or equal to 100.0 given '105'",
							}),
					).To(BeNil())

				})

				It("valid greater than zero", func() {
					bolusObj["extended"] = 0.7
					Expect(helper.ValidDataType(bolus.Build(bolusObj, helper.ErrorProcessing))).To(BeNil())
				})

			})
		})
	})
})