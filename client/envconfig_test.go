package client

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

// pretend we have testing.T.Setenv, which ginkgo didn't bother to implement.
func setenv(t GinkgoTInterface, key, val string) {
	t.Cleanup(func() {
		os.Setenv(key, os.Getenv(key))
	})
	os.Setenv(key, val)
}

var _ = Describe("envconfigReporter", func() {
	t := GinkgoT()

	It("does", func() {
		setenv(t, "TIDEPOOL_ADDRESS", "1")
		r := NewEnvconfigReporter("TIDEPOOL")
		foo, err := r.Get("ADDRESS")
		Expect(err).ShouldNot(HaveOccurred())
		Expect(foo).To(Equal("1"))
	})
})
