package validator_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"regexp"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/structure"
	structureBase "github.com/tidepool-org/platform/structure/base"
	structureValidator "github.com/tidepool-org/platform/structure/validator"
)

var _ = Describe("StringArray", func() {
	var base *structureBase.Base

	BeforeEach(func() {
		base = structureBase.New()
	})

	Context("NewStringArray", func() {
		It("returns successfully", func() {
			value := []string{"one", "two"}
			Expect(structureValidator.NewStringArray(base, &value)).ToNot(BeNil())
		})
	})

	Context("with new validator with nil value", func() {
		var validator *structureValidator.StringArray
		var result structure.StringArray

		BeforeEach(func() {
			validator = structureValidator.NewStringArray(base, nil)
			Expect(validator).ToNot(BeNil())
		})

		Context("Exists", func() {
			BeforeEach(func() {
				result = validator.Exists()
			})

			It("reports the expected error", func() {
				Expect(base.Error()).To(HaveOccurred())
				Expect(errors.Sanitize(base.Error())).To(Equal(errors.Sanitize(structureValidator.ErrorValueNotExists())))
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("NotExists", func() {
			BeforeEach(func() {
				result = validator.NotExists()
			})

			It("does not report an error", func() {
				Expect(base.Error()).ToNot(HaveOccurred())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("Empty", func() {
			BeforeEach(func() {
				result = validator.Empty()
			})

			It("does not report an error", func() {
				Expect(base.Error()).ToNot(HaveOccurred())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("NotEmpty", func() {
			BeforeEach(func() {
				result = validator.NotEmpty()
			})

			It("does not report an error", func() {
				Expect(base.Error()).ToNot(HaveOccurred())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("LengthEqualTo", func() {
			BeforeEach(func() {
				result = validator.LengthEqualTo(1)
			})

			It("does not report an error", func() {
				Expect(base.Error()).ToNot(HaveOccurred())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("LengthNotEqualTo", func() {
			BeforeEach(func() {
				result = validator.LengthNotEqualTo(4)
			})

			It("does not report an error", func() {
				Expect(base.Error()).ToNot(HaveOccurred())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("LengthLessThan", func() {
			BeforeEach(func() {
				result = validator.LengthLessThan(3)
			})

			It("does not report an error", func() {
				Expect(base.Error()).ToNot(HaveOccurred())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("LengthLessThanOrEqualTo", func() {
			BeforeEach(func() {
				result = validator.LengthLessThan(1)
			})

			It("does not report an error", func() {
				Expect(base.Error()).ToNot(HaveOccurred())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("LengthGreaterThan", func() {
			BeforeEach(func() {
				result = validator.LengthGreaterThan(3)
			})

			It("does not report an error", func() {
				Expect(base.Error()).ToNot(HaveOccurred())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("LengthGreaterThanOrEqualTo", func() {
			BeforeEach(func() {
				result = validator.LengthGreaterThanOrEqualTo(4)
			})

			It("does not report an error", func() {
				Expect(base.Error()).ToNot(HaveOccurred())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("LengthInRange", func() {
			BeforeEach(func() {
				result = validator.LengthInRange(0, 1)
			})

			It("does not report an error", func() {
				Expect(base.Error()).ToNot(HaveOccurred())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("EachOneOf", func() {
			BeforeEach(func() {
				result = validator.EachOneOf("1", "seven")
			})

			It("does not report an error", func() {
				Expect(base.Error()).ToNot(HaveOccurred())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("EachNotOneOf", func() {
			BeforeEach(func() {
				result = validator.EachNotOneOf("seven", "four")
			})

			It("does not report an error", func() {
				Expect(base.Error()).ToNot(HaveOccurred())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("EachMatches", func() {
			var expression *regexp.Regexp

			BeforeEach(func() {
				expression = regexp.MustCompile("^[0-9]*$")
				result = validator.EachMatches(expression)
			})

			It("does not report an error", func() {
				Expect(base.Error()).ToNot(HaveOccurred())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("EachNotMatches", func() {
			var expression *regexp.Regexp

			BeforeEach(func() {
				expression = regexp.MustCompile("^.ou.$")
				result = validator.EachNotMatches(expression)
			})

			It("does not report an error", func() {
				Expect(base.Error()).ToNot(HaveOccurred())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})
	})

	Context("with new validator with empty string array value", func() {
		var validator *structureValidator.StringArray
		var result structure.StringArray
		var value []string

		BeforeEach(func() {
			value = []string{}
			validator = structureValidator.NewStringArray(base, &value)
			Expect(validator).ToNot(BeNil())
		})

		Context("Exists", func() {
			BeforeEach(func() {
				result = validator.Exists()
			})

			It("does not report an error", func() {
				Expect(base.Error()).ToNot(HaveOccurred())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("NotExists", func() {
			BeforeEach(func() {
				result = validator.NotExists()
			})

			It("reports the expected error", func() {
				Expect(base.Error()).To(HaveOccurred())
				Expect(errors.Sanitize(base.Error())).To(Equal(errors.Sanitize(structureValidator.ErrorValueExists())))
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("Empty", func() {
			BeforeEach(func() {
				result = validator.Empty()
			})

			It("does not report an error", func() {
				Expect(base.Error()).ToNot(HaveOccurred())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("NotEmpty", func() {
			BeforeEach(func() {
				result = validator.NotEmpty()
			})

			It("does not report an error", func() {
				Expect(base.Error()).To(HaveOccurred())
				Expect(errors.Sanitize(base.Error())).To(Equal(errors.Sanitize(structureValidator.ErrorValueEmpty())))
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})
	})

	Context("with new validator with value with length of 1", func() {
		var validator *structureValidator.StringArray
		var result structure.StringArray
		var value []string

		BeforeEach(func() {
			value = []string{"1"}
			validator = structureValidator.NewStringArray(base, &value)
			Expect(validator).ToNot(BeNil())
		})

		Context("Exists", func() {
			BeforeEach(func() {
				result = validator.Exists()
			})

			It("does not report an error", func() {
				Expect(base.Error()).ToNot(HaveOccurred())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("NotExists", func() {
			BeforeEach(func() {
				result = validator.NotExists()
			})

			It("reports the expected error", func() {
				Expect(base.Error()).To(HaveOccurred())
				Expect(errors.Sanitize(base.Error())).To(Equal(errors.Sanitize(structureValidator.ErrorValueExists())))
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("Empty", func() {
			BeforeEach(func() {
				result = validator.Empty()
			})

			It("reports the expected error", func() {
				Expect(base.Error()).To(HaveOccurred())
				Expect(errors.Sanitize(base.Error())).To(Equal(errors.Sanitize(structureValidator.ErrorValueNotEmpty())))
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("NotEmpty", func() {
			BeforeEach(func() {
				result = validator.NotEmpty()
			})

			It("does not report an error", func() {
				Expect(base.Error()).ToNot(HaveOccurred())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("LengthEqualTo", func() {
			BeforeEach(func() {
				result = validator.LengthEqualTo(1)
			})

			It("does not report an error", func() {
				Expect(base.Error()).ToNot(HaveOccurred())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("LengthNotEqualTo", func() {
			BeforeEach(func() {
				result = validator.LengthNotEqualTo(4)
			})

			It("does not report an error", func() {
				Expect(base.Error()).ToNot(HaveOccurred())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("LengthLessThan", func() {
			BeforeEach(func() {
				result = validator.LengthLessThan(3)
			})

			It("does not report an error", func() {
				Expect(base.Error()).ToNot(HaveOccurred())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("LengthLessThanOrEqualTo", func() {
			BeforeEach(func() {
				result = validator.LengthLessThanOrEqualTo(1)
			})

			It("does not report an error", func() {
				Expect(base.Error()).ToNot(HaveOccurred())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("LengthGreaterThan", func() {
			BeforeEach(func() {
				result = validator.LengthGreaterThan(3)
			})

			It("reports the expected error", func() {
				Expect(base.Error()).To(HaveOccurred())
				Expect(errors.Sanitize(base.Error())).To(Equal(errors.Sanitize(structureValidator.ErrorLengthNotGreaterThan(1, 3))))
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("LengthGreaterThanOrEqualTo", func() {
			BeforeEach(func() {
				result = validator.LengthGreaterThanOrEqualTo(4)
			})

			It("reports the expected error", func() {
				Expect(base.Error()).To(HaveOccurred())
				Expect(errors.Sanitize(base.Error())).To(Equal(errors.Sanitize(structureValidator.ErrorLengthNotGreaterThanOrEqualTo(1, 4))))
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("LengthInRange", func() {
			BeforeEach(func() {
				result = validator.LengthInRange(0, 3)
			})

			It("does not report an error", func() {
				Expect(base.Error()).ToNot(HaveOccurred())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("EachOneOf", func() {
			BeforeEach(func() {
				result = validator.EachOneOf("1", "seven")
			})

			It("does not report an error", func() {
				Expect(base.Error()).ToNot(HaveOccurred())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("EachNotOneOf", func() {
			BeforeEach(func() {
				result = validator.EachNotOneOf("seven", "four")
			})

			It("does not report an error", func() {
				Expect(base.Error()).ToNot(HaveOccurred())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("EachMatches", func() {
			var expression *regexp.Regexp

			BeforeEach(func() {
				expression = regexp.MustCompile("^[0-9]*$")
				result = validator.EachMatches(expression)
			})

			It("reports the expected error", func() {
				Expect(base.Error()).ToNot(HaveOccurred())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("EachNotMatches", func() {
			var expression *regexp.Regexp

			BeforeEach(func() {
				expression = regexp.MustCompile("^.ou.$")
				result = validator.EachNotMatches(expression)
			})

			It("reports the expected error", func() {
				Expect(base.Error()).ToNot(HaveOccurred())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})
	})

	Context("with new validator with value with length of 4", func() {
		var validator *structureValidator.StringArray
		var result structure.StringArray
		var value []string

		BeforeEach(func() {
			value = []string{"1", "two", "three", "four"}
			validator = structureValidator.NewStringArray(base, &value)
			Expect(validator).ToNot(BeNil())
		})

		Context("Exists", func() {
			BeforeEach(func() {
				result = validator.Exists()
			})

			It("does not report an error", func() {
				Expect(base.Error()).ToNot(HaveOccurred())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("NotExists", func() {
			BeforeEach(func() {
				result = validator.NotExists()
			})

			It("reports the expected error", func() {
				Expect(base.Error()).To(HaveOccurred())
				Expect(errors.Sanitize(base.Error())).To(Equal(errors.Sanitize(structureValidator.ErrorValueExists())))
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("Empty", func() {
			BeforeEach(func() {
				result = validator.Empty()
			})

			It("reports the expected error", func() {
				Expect(base.Error()).To(HaveOccurred())
				Expect(errors.Sanitize(base.Error())).To(Equal(errors.Sanitize(structureValidator.ErrorValueNotEmpty())))
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("NotEmpty", func() {
			BeforeEach(func() {
				result = validator.NotEmpty()
			})

			It("does not report an error", func() {
				Expect(base.Error()).ToNot(HaveOccurred())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("LengthEqualTo", func() {
			BeforeEach(func() {
				result = validator.LengthEqualTo(1)
			})

			It("reports the expected error", func() {
				Expect(base.Error()).To(HaveOccurred())
				Expect(errors.Sanitize(base.Error())).To(Equal(errors.Sanitize(structureValidator.ErrorLengthNotEqualTo(4, 1))))
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("LengthNotEqualTo", func() {
			BeforeEach(func() {
				result = validator.LengthNotEqualTo(4)
			})

			It("reports the expected error", func() {
				Expect(base.Error()).To(HaveOccurred())
				Expect(errors.Sanitize(base.Error())).To(Equal(errors.Sanitize(structureValidator.ErrorLengthEqualTo(4, 4))))
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("LengthLessThan", func() {
			BeforeEach(func() {
				result = validator.LengthLessThan(3)
			})

			It("reports the expected error", func() {
				Expect(base.Error()).To(HaveOccurred())
				Expect(errors.Sanitize(base.Error())).To(Equal(errors.Sanitize(structureValidator.ErrorLengthNotLessThan(4, 3))))
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("LengthLessThanOrEqualTo", func() {
			BeforeEach(func() {
				result = validator.LengthLessThanOrEqualTo(1)
			})

			It("reports the expected error", func() {
				Expect(base.Error()).To(HaveOccurred())
				Expect(errors.Sanitize(base.Error())).To(Equal(errors.Sanitize(structureValidator.ErrorLengthNotLessThanOrEqualTo(4, 1))))
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("LengthGreaterThan", func() {
			BeforeEach(func() {
				result = validator.LengthGreaterThan(3)
			})

			It("does not report an error", func() {
				Expect(base.Error()).ToNot(HaveOccurred())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("LengthGreaterThanOrEqualTo", func() {
			BeforeEach(func() {
				result = validator.LengthGreaterThanOrEqualTo(4)
			})

			It("does not report an error", func() {
				Expect(base.Error()).ToNot(HaveOccurred())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("LengthInRange", func() {
			BeforeEach(func() {
				result = validator.LengthInRange(0, 3)
			})

			It("reports the expected error", func() {
				Expect(base.Error()).To(HaveOccurred())
				Expect(errors.Sanitize(base.Error())).To(Equal(errors.Sanitize(structureValidator.ErrorLengthNotInRange(4, 0, 3))))
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("EachOneOf", func() {
			BeforeEach(func() {
				result = validator.EachOneOf("1", "seven")
			})

			It("reports the expected error", func() {
				Expect(base.Error()).To(HaveOccurred())
				Expect(errors.Sanitize(base.Error())).To(Equal(errors.Sanitize(errors.Append(
					structureValidator.ErrorValueStringNotOneOf("two", []string{"1", "seven"}),
					structureValidator.ErrorValueStringNotOneOf("three", []string{"1", "seven"}),
					structureValidator.ErrorValueStringNotOneOf("four", []string{"1", "seven"}),
				))))
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("EachNotOneOf", func() {
			BeforeEach(func() {
				result = validator.EachNotOneOf("seven", "four")
			})

			It("reports the expected error", func() {
				Expect(base.Error()).To(HaveOccurred())
				Expect(errors.Sanitize(base.Error())).To(Equal(errors.Sanitize(structureValidator.ErrorValueStringOneOf("four", []string{"seven", "four"}))))
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("EachOneOf with no allowed values", func() {
			BeforeEach(func() {
				result = validator.EachOneOf()
			})

			It("reports the expected error", func() {
				Expect(base.Error()).To(HaveOccurred())
				Expect(errors.Sanitize(base.Error())).To(Equal(errors.Sanitize(errors.Append(
					structureValidator.ErrorValueStringNotOneOf("1", []string{}),
					structureValidator.ErrorValueStringNotOneOf("two", []string{}),
					structureValidator.ErrorValueStringNotOneOf("three", []string{}),
					structureValidator.ErrorValueStringNotOneOf("four", []string{}),
				))))
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("EachNotOneOf with no disallowed values", func() {
			BeforeEach(func() {
				result = validator.EachNotOneOf()
			})

			It("does not report an error", func() {
				Expect(base.Error()).ToNot(HaveOccurred())
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("EachMatches", func() {
			var expression *regexp.Regexp

			BeforeEach(func() {
				expression = regexp.MustCompile("^[0-9]*$")
				result = validator.EachMatches(expression)
			})

			It("reports the expected error", func() {
				Expect(base.Error()).To(HaveOccurred())
				Expect(errors.Sanitize(base.Error())).To(Equal(errors.Sanitize(errors.Append(
					structureValidator.ErrorValueStringNotMatches("two", expression),
					structureValidator.ErrorValueStringNotMatches("three", expression),
					structureValidator.ErrorValueStringNotMatches("four", expression),
				))))
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("EachNotMatches", func() {
			var expression *regexp.Regexp

			BeforeEach(func() {
				expression = regexp.MustCompile("^.ou.$")
				result = validator.EachNotMatches(expression)
			})

			It("reports the expected error", func() {
				Expect(base.Error()).To(HaveOccurred())
				Expect(errors.Sanitize(base.Error())).To(Equal(errors.Sanitize(structureValidator.ErrorValueStringMatches("four", expression))))
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("EachMatches with no expression", func() {
			BeforeEach(func() {
				result = validator.EachMatches(nil)
			})

			It("reports the expected error", func() {
				Expect(base.Error()).To(HaveOccurred())
				Expect(errors.Sanitize(base.Error())).To(Equal(errors.Sanitize(errors.Append(
					structureValidator.ErrorValueStringNotMatches("1", nil),
					structureValidator.ErrorValueStringNotMatches("two", nil),
					structureValidator.ErrorValueStringNotMatches("three", nil),
					structureValidator.ErrorValueStringNotMatches("four", nil),
				))))
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})

		Context("EachNotMatches with no expression", func() {
			BeforeEach(func() {
				result = validator.EachNotMatches(nil)
			})

			It("does not report an error", func() {
				Expect(base.Error()).To(HaveOccurred())
				Expect(errors.Sanitize(base.Error())).To(Equal(errors.Sanitize(errors.Append(
					structureValidator.ErrorValueStringMatches("1", nil),
					structureValidator.ErrorValueStringMatches("two", nil),
					structureValidator.ErrorValueStringMatches("three", nil),
					structureValidator.ErrorValueStringMatches("four", nil),
				))))
			})

			It("returns self", func() {
				Expect(result).To(BeIdenticalTo(validator))
			})
		})
	})
})