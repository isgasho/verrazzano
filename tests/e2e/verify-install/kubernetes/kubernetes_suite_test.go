package kubernetes_test

import (
	"fmt"
	"testing"

	"github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/reporters"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestKubernetes(t *testing.T) {
	RegisterFailHandler(Fail)
	junitReporter := reporters.NewJUnitReporter(fmt.Sprintf("kubernetes-%d-test-result.xml", config.GinkgoConfig.ParallelNode))
	RunSpecsWithDefaultAndCustomReporters(t, "Kubernetes Suite", []Reporter{junitReporter})
}
