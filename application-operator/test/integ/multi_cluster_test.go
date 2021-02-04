// Copyright (C) 2021, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.
package integ

import (
	"fmt"
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	"github.com/verrazzano/verrazzano/application-operator/test/integ/util"
)

const (
	multiclusterTestNamespace = "multiclustertest"
	crdDir                    = "../../config/crd/bases"
)

var (
	multiclusterCrds = [2]string{
		fmt.Sprintf("%v/clusters.verrazzano.io_multiclusternamespaces.yaml", crdDir),
		fmt.Sprintf("%v/clusters.verrazzano.io_multiclustersecrets.yaml", crdDir),
	}
)

var _ = ginkgo.Describe("Testing Multi-Cluster Namespace CRD", func() {
	ginkgo.It("MultiCluster CRDs can be applied", func() {
		for _, crd := range multiclusterCrds {
			_, stderr := util.Kubectl(fmt.Sprintf("apply -f %v", crd))
			gomega.Expect(stderr).To(gomega.Equal(""), fmt.Sprintf("Failed to apply CRD %v", crd))
		}
		_, stderr := util.Kubectl("create ns " + multiclusterTestNamespace)
		gomega.Expect(stderr).To(gomega.Equal(""), fmt.Sprintf("failed to create namespace %v", multiclusterTestNamespace))
	})
	ginkgo.It("MultiClusterNamespace can be created ", func() {
		_, stderr := util.Kubectl("apply -f testdata/multi-cluster/multicluster_namespace_sample.yaml")
		gomega.Expect(stderr).To(gomega.Equal(""))
	})
	ginkgo.It("MultiClusterSecret can be created ", func() {
		_, stderr := util.Kubectl("apply -f testdata/multi-cluster/multicluster_secret_sample.yaml")
		gomega.Expect(stderr).To(gomega.Equal(""))
	})
})