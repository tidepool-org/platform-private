package mongo_test

import (
	"fmt"
	"testing"

	. "github.com/onsi/ginkgo/v2"

	"github.com/tidepool-org/platform/test"
)

func TestSuite(t *testing.T) {
	test.Test(t)
}

var _ = BeforeSuite(func() {
	fmt.Println("before suite store/mongo")
})
