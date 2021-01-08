// Copyright (c) 2020, 2021, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package verrazzano_test

import (
	"fmt"
	"github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/reporters"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestVerrazzano(t *testing.T) {
	RegisterFailHandler(Fail)
	junitReporter := reporters.NewJUnitReporter(fmt.Sprintf("verrazzano-%d-test-result.xml", config.GinkgoConfig.ParallelNode))
	RunSpecsWithDefaultAndCustomReporters(t, "Verrazzano Suite", []Reporter{junitReporter})
}
