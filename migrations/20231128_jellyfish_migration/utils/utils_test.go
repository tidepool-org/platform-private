package utils_test

import (
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/tidepool-org/platform/data/blood/glucose"
	"github.com/tidepool-org/platform/data/normalizer"
	"github.com/tidepool-org/platform/data/types/blood/glucose/continuous"
	glucoseTest "github.com/tidepool-org/platform/data/types/blood/glucose/test"
	bolusTest "github.com/tidepool-org/platform/data/types/bolus/test"
	"github.com/tidepool-org/platform/data/types/common"
	"github.com/tidepool-org/platform/data/types/settings/pump"
	pumpTest "github.com/tidepool-org/platform/data/types/settings/pump/test"
	"github.com/tidepool-org/platform/migrations/20231128_jellyfish_migration/utils"
	"github.com/tidepool-org/platform/migrations/20231128_jellyfish_migration/utils/test"
	"github.com/tidepool-org/platform/pointer"
)

var _ = Describe("back-37", func() {
	var _ = Describe("utils", func() {
		var _ = Describe("GetDatumUpdates", func() {
			var existingBolusDatum bson.M
			const expectedID = "some-id"

			var getBSONData = func(datum interface{}) bson.M {
				var bsonData bson.M
				bsonAsByte, _ := bson.Marshal(&datum)
				bson.Unmarshal(bsonAsByte, &bsonData)
				return bsonData
			}

			BeforeEach(func() {
				datum := bolusTest.RandomBolus()
				*datum.ID = expectedID
				*datum.UserID = "some-user-id"
				*datum.DeviceID = "some-device-id"
				datum.SubType = "some-subtype"
				theTime, _ := time.Parse(time.RFC3339, "2016-09-01T11:00:00Z")
				*datum.Time = theTime
				existingBolusDatum = getBSONData(datum)
				existingBolusDatum["_id"] = expectedID
				Expect(existingBolusDatum).ToNot(BeNil())
			})

			Context("_deduplicator hash", func() {
				DescribeTable("should",
					func(getInput func() bson.M, expectedUpdates []bson.M, expectError bool) {
						input := getInput()
						actualID, actualUpdates, err := utils.GetDatumUpdates(input)
						if expectError {
							Expect(err).ToNot(BeNil())
							Expect(actualUpdates).To(BeNil())
							return
						}
						Expect(err).To(BeNil())
						if expectedUpdates != nil {
							Expect(actualUpdates).To(Equal(expectedUpdates))
							Expect(actualID).To(Equal(expectedID))
						} else {
							Expect(actualUpdates).To(BeNil())
						}
					},
					Entry("error when missing _userId", func() bson.M {
						existingBolusDatum["_userId"] = nil
						return existingBolusDatum
					}, nil, true),
					Entry("error when missing deviceId", func() bson.M {
						existingBolusDatum["deviceId"] = nil
						return existingBolusDatum
					}, nil, true),
					Entry("error when missing time", func() bson.M {
						existingBolusDatum["time"] = nil
						return existingBolusDatum
					}, nil, true),
					Entry("error when missing type", func() bson.M {
						existingBolusDatum["type"] = nil
						return existingBolusDatum
					}, nil, true),
					Entry("adds hash when vaild", func() bson.M {
						return existingBolusDatum
					},
						[]bson.M{
							{"$set": bson.M{"_deduplicator": bson.M{"hash": "FVjexdlY6mWkmoh5gdmtdhzVH4R03+iGE81ro08/KcE="}}},
						},
						false,
					),
				)
			})

			Context("pumpSettings", func() {

				var pumpSettingsDatum *pump.Pump

				BeforeEach(func() {
					mmolL := pump.DisplayBloodGlucoseUnitsMmolPerL
					pumpSettingsDatum = pumpTest.NewPump(&mmolL)
					*pumpSettingsDatum.ID = expectedID
					*pumpSettingsDatum.UserID = "some-user-id"
					*pumpSettingsDatum.DeviceID = "some-device-id"
					theTime, _ := time.Parse(time.RFC3339, "2016-09-01T11:00:00Z")
					*pumpSettingsDatum.Time = theTime
				})

				Context("with mis-named jellyfish bolus", func() {
					var bolusData = &pump.BolusMap{
						"bolus-1": pumpTest.NewRandomBolus(),
						"bolus-2": pumpTest.NewRandomBolus(),
					}
					var settingsBolusDatum bson.M

					BeforeEach(func() {
						settingsBolusDatum = getBSONData(pumpSettingsDatum)
						settingsBolusDatum["bolus"] = bolusData
						settingsBolusDatum["_id"] = expectedID
					})

					DescribeTable("should",
						func(getInput func() bson.M, expected []bson.M, expectError bool) {
							input := getInput()
							_, actual, err := utils.GetDatumUpdates(input)
							if expectError {
								Expect(err).ToNot(BeNil())
								Expect(actual).To(BeNil())
								return
							}
							Expect(err).To(BeNil())
							if expected != nil {
								Expect(actual).To(BeEquivalentTo(expected))
							} else {
								Expect(actual).To(BeNil())
							}
						},

						Entry("do nothing when wrong type",
							func() bson.M {
								return existingBolusDatum
							},
							[]bson.M{{"$set": bson.M{"_deduplicator": bson.M{"hash": "FVjexdlY6mWkmoh5gdmtdhzVH4R03+iGE81ro08/KcE="}}}},
							false,
						),
						Entry("do nothing when has no bolus",
							func() bson.M {
								settingsBolusDatum["bolus"] = nil
								return settingsBolusDatum
							},
							[]bson.M{{"$set": bson.M{"_deduplicator": bson.M{"hash": "RlrPcuPDfRim29UwnM7Yf0Ib0Ht4F35qvHu62CCYXnM="}}}},
							false,
						),
						Entry("add boluses when bolus found",
							func() bson.M {
								settingsBolusDatum["bolus"] = bolusData
								return settingsBolusDatum
							},
							[]bson.M{
								{"$set": bson.M{"_deduplicator": bson.M{"hash": "RlrPcuPDfRim29UwnM7Yf0Ib0Ht4F35qvHu62CCYXnM="}}},
								{"$rename": bson.M{"bolus": "boluses"}},
							},
							false,
						),
					)
				})
				Context("unordered sleepSchedules", func() {
					expectedSleepSchedulesMap := &pump.SleepScheduleMap{}
					var invalidDays *pump.SleepSchedule
					var s1Days *pump.SleepSchedule
					var s2Days *pump.SleepSchedule
					var sleepSchedulesDatum bson.M
					BeforeEach(func() {
						s1 := pumpTest.RandomSleepSchedule()
						s2 := pumpTest.RandomSleepSchedule()
						(*expectedSleepSchedulesMap)["1"] = s1
						(*expectedSleepSchedulesMap)["2"] = s2

						s1Days = pumpTest.CloneSleepSchedule(s1)
						for key, day := range *s1Days.Days {
							(*s1Days.Days)[key] = strings.ToUpper(day)
						}
						s2Days = pumpTest.CloneSleepSchedule(s2)
						for key, day := range *s2Days.Days {
							(*s2Days.Days)[key] = strings.ToUpper(day)
						}
						invalidDays = pumpTest.CloneSleepSchedule(s2)
						invalidDays.Days = &[]string{"not-a-day", common.DayFriday}

						//to ensure correct sorting
						expectedSleepSchedulesMap.Normalize(normalizer.New())

						Expect(expectedSleepSchedulesMap).ToNot(BeNil())
						pumpSettingsDatum.SleepSchedules = nil
						sleepSchedulesDatum = getBSONData(pumpSettingsDatum)
						sleepSchedulesDatum["_id"] = expectedID
						sleepSchedulesDatum["bolus"] = nil //remove as not testing here
					})

					It("does nothing when wrong type", func() {
						_, actual, err := utils.GetDatumUpdates(existingBolusDatum)
						Expect(err).To(BeNil())
						Expect(len(actual)).To(Equal(1))
						Expect(actual).To(Equal([]bson.M{{"$set": bson.M{"_deduplicator": bson.M{"hash": "FVjexdlY6mWkmoh5gdmtdhzVH4R03+iGE81ro08/KcE="}}}}))
					})
					It("does nothing when no sleepSchedules", func() {
						sleepSchedulesDatum["sleepSchedules"] = nil
						_, actual, err := utils.GetDatumUpdates(sleepSchedulesDatum)
						Expect(err).To(BeNil())
						Expect(len(actual)).To(Equal(1))
						Expect(actual).To(Equal([]bson.M{{"$set": bson.M{"_deduplicator": bson.M{"hash": "RlrPcuPDfRim29UwnM7Yf0Ib0Ht4F35qvHu62CCYXnM="}}}}))
					})
					It("returns updated sleepSchedules when valid", func() {
						sleepSchedulesDatum["sleepSchedules"] = []*pump.SleepSchedule{s1Days, s2Days}
						_, actual, err := utils.GetDatumUpdates(sleepSchedulesDatum)
						Expect(err).To(BeNil())
						Expect(len(actual)).To(Equal(1))
						setData := actual[0]["$set"].(bson.M)
						Expect(setData["sleepSchedules"]).ToNot(BeNil())
						Expect(setData["sleepSchedules"]).To(Equal(expectedSleepSchedulesMap))
					})
					It("returns error when sleepSchedules have invalid days", func() {
						sleepSchedulesDatum["sleepSchedules"] = []*pump.SleepSchedule{invalidDays}
						_, actual, err := utils.GetDatumUpdates(sleepSchedulesDatum)
						Expect(err).ToNot(BeNil())
						Expect(err.Error()).To(Equal("pumpSettings.sleepSchedules has an invalid day of week not-a-day"))
						Expect(actual).To(BeNil())
					})
				})
			})
			Context("datum with glucose", func() {

				var newContinuous = func(units *string) *continuous.Continuous {
					datum := continuous.New()
					datum.Glucose = *glucoseTest.NewGlucose(units)
					datum.Type = "cbg"
					*datum.ID = expectedID
					*datum.UserID = "some-user-id"
					*datum.DeviceID = "some-device-id"
					theTime, _ := time.Parse(time.RFC3339, "2016-09-01T11:00:00Z")
					*datum.Time = theTime
					return datum
				}

				DescribeTable("should",
					func(getInput func() bson.M, expected []bson.M, expectError bool) {
						input := getInput()
						_, actual, err := utils.GetDatumUpdates(input)
						if expectError {
							Expect(err).ToNot(BeNil())
							Expect(actual).To(BeNil())
							return
						}
						Expect(err).To(BeNil())
						if expected != nil {
							Expect(actual).To(BeEquivalentTo(expected))
						} else {
							Expect(actual).To(BeNil())
						}
					},

					Entry("do nothing when not normailzed",
						func() bson.M {
							mmolL := glucose.MmolL
							cbg := newContinuous(&mmolL)
							cbgData := getBSONData(cbg)
							cbgData["_id"] = expectedID
							cbgData["value"] = 9.5
							return cbgData
						},
						[]bson.M{{"$set": bson.M{
							"_deduplicator": bson.M{"hash": "lLCOZJMLvNaBx7dMc31bbX4zwSfPvxcUd0Z1uU/YIAs="},
						}}},
						false,
					),
					Entry("update value when normailzed",
						func() bson.M {
							mgdL := glucose.MgdL
							cbg := newContinuous(&mgdL)
							cbgData := getBSONData(cbg)
							cbgData["_id"] = expectedID
							cbgData["value"] = 180

							return cbgData
						},
						[]bson.M{{"$set": bson.M{
							"_deduplicator": bson.M{"hash": "FZtVRkliues5vAt25ZK+WDAqa4Q6tAAe9h2PdKM15Q4="},
							"value":         pointer.FromFloat64(9.99135),
							"units":         pointer.FromString(glucose.MmolL),
						}}},
						false,
					),
				)
			})
			Context("Historic datum", func() {
				It("g5 dexcom", func() {
					actualID, actual, err := utils.GetDatumUpdates(getBSONData(test.CBGDexcomG5MobDatum))
					Expect(err).To(BeNil())
					Expect(actual).ToNot(BeNil())
					Expect(actual).To(Equal([]bson.M{{"$set": bson.M{"_deduplicator": bson.M{"hash": "TKJurm+/SuA5tarn/nATa7Nw0LXgwGel67lgJihUctM="}}}}))
					Expect(actualID).ToNot(BeEmpty())
				})

				It("carelink medtronic pumpSettings", func() {
					actualID, actual, err := utils.GetDatumUpdates(getBSONData(test.PumpSettingsCarelink))
					Expect(err).To(BeNil())
					Expect(actual).ToNot(BeNil())
					Expect(actual).To(Equal([]bson.M{{"$set": bson.M{"_deduplicator": bson.M{"hash": "NC17pw1UAaab50iChhQXJ+N9dTi6GduTy9UjsMHolow="}}}}))
					Expect(actualID).ToNot(BeEmpty())
				})

				It("tandem pumpSettings", func() {
					actualID, actual, err := utils.GetDatumUpdates(getBSONData(test.PumpSettingsTandem))
					Expect(err).To(BeNil())
					Expect(actual).ToNot(BeNil())
					Expect(len(actual)).To(Equal(1))
					Expect(actual).To(Equal([]bson.M{{"$set": bson.M{"_deduplicator": bson.M{"hash": "bpKLJbi5JfqD7N0WJ1vj0ck03c9EZ3U0H09TCLhdd3k="}}}}))
					Expect(actualID).ToNot(BeEmpty())
				})

				It("omnipod pumpSettings", func() {
					actualID, actual, err := utils.GetDatumUpdates(getBSONData(test.PumpSettingsOmnipod))
					Expect(err).To(BeNil())
					Expect(actual).ToNot(BeNil())
					Expect(actual).To(Equal([]bson.M{{"$set": bson.M{"_deduplicator": bson.M{"hash": "oH7/6EEgUjRTeafEpm74fVTYMBvMdQ65/rhg0oFoev8="}}}}))
					Expect(actualID).ToNot(BeEmpty())
				})

			})
		})
	})
})