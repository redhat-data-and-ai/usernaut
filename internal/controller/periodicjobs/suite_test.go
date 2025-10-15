package periodicjobs_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestPeriodicJobs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "PeriodicJobs Suite")
}
