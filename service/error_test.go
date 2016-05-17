package service_test

import (
	"encoding/json"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/tidepool-org/platform/service"
)

var _ = Describe("Error", func() {
	Context("Source struct", func() {
		Context("encoded as JSON", func() {
			It("is an empty object if no fields are specified", func() {
				source := &service.Source{}
				Expect(json.Marshal(source)).To(MatchJSON("{}"))
			})

			It("is a populated object if fields are specified", func() {
				source := &service.Source{
					Parameter: "test-parameter",
					Pointer:   "test-pointer",
				}
				Expect(json.Marshal(source)).To(MatchJSON("{" +
					"\"parameter\":\"test-parameter\"," +
					"\"pointer\":\"test-pointer\"" +
					"}"))
			})
		})
	})

	Context("Error struct", func() {
		Context("encoded as JSON", func() {
			It("is an empty object if no fields are specified", func() {
				err := &service.Error{}
				Expect(json.Marshal(err)).To(MatchJSON("{}"))
			})

			It("is a populated object if fields are specified", func() {
				err := &service.Error{
					Code:   "test-code",
					Detail: "test-detail",
					Source: &service.Source{
						Parameter: "test-parameter",
						Pointer:   "test-pointer",
					},
					Status: 400,
					Title:  "test-title",
				}
				Expect(json.Marshal(err)).To(MatchJSON("{" +
					"\"code\":\"test-code\"," +
					"\"detail\":\"test-detail\"," +
					"\"source\":{" +
					"\"parameter\":\"test-parameter\"," +
					"\"pointer\":\"test-pointer\"" +
					"}," +
					"\"status\":\"400\"," +
					"\"title\":\"test-title\"" +
					"}"))
			})
		})
	})
})