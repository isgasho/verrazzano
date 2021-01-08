package verrazzano_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	. "github.com/verrazzano/verrazzano/tests/e2e/util"
)

var _ = Describe("Verrazzano", func() {

	DescribeTable("CRD for",
		func(name string) {
			Expect(DoesCRDExist(name)).To(BeTrue())
		},
		Entry("verrazzanos should exist in cluster", "vverrazzanos.install.verrazzano.io"),
	)

})
