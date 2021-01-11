// Copyright (c) 2020, 2021, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package istio_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	. "github.com/verrazzano/verrazzano/tests/e2e/util"
	appsv1 "k8s.io/api/apps/v1"
)

var _ = Describe("Istio", func() {
	const istioNamespace = "istio-system"

	DescribeTable("namespace",
		func(name string) {
			Expect(DoesNamespaceExist(name)).To(BeTrue())
		},
		Entry(fmt.Sprintf("%s namespace should exist", istioNamespace), istioNamespace),
	)

	DescribeTable("deployments",
		func(namespace string) {
			expectedDeployments := []string{
				"grafana",
				"istio-citadel",
				"istio-egressgateway",
				"istio-galley",
				"istio-ingressgateway",
				"istio-pilot",
				"istio-policy",
				"istio-sidecar-injector",
				"istio-telemetry",
				"istiocoredns",
				"prometheus",
			}

			deploymentNames := func(deploymentList *appsv1.DeploymentList) []string {
				deploymentNames := []string{}
				for _, deployment := range deploymentList.Items {
					deploymentNames = append(deploymentNames, deployment.Name)
				}
				return deploymentNames
			}
			deployments := GetDeploymentList(namespace)
			Expect(deployments).Should(
				SatisfyAll(
					Not(BeNil()),
					WithTransform(deploymentNames, ContainElements(expectedDeployments)),
				),
			)
			Expect(len(deployments.Items)).To(Equal(len(expectedDeployments)))
		},
		Entry(fmt.Sprintf("%s namespace should contain expected list of deployments", istioNamespace), istioNamespace),
	)

	const istioJob = "istio-init-crd-14-1.4.6"
	DescribeTable("job",
		func(namespace string, name string) {
			Expect(DoesJobExist(namespace, name)).To(BeTrue())
		},
		Entry(fmt.Sprintf("%s namespace should contain job %s", istioNamespace, istioJob), istioNamespace, istioJob),
	)

})
