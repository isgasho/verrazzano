// Copyright (c) 2020, 2021, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

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
