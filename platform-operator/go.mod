// Copyright (c) 2020, 2021, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

module github.com/verrazzano/verrazzano/platform-operator

go 1.13

require (
	github.com/golang/mock v1.4.4
	github.com/gordonklaus/ineffassign v0.0.0-20210104184537-8eed68eb605f
	github.com/kylelemons/godebug v1.1.0
	github.com/onsi/ginkgo v1.12.0
	github.com/onsi/gomega v1.9.0
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.5.1
	github.com/verrazzano/verrazzano-crd-generator v0.0.0-20201214161122-0330d094db41
	github.com/verrazzano/verrazzano-monitoring-operator v0.0.0-20201110193547-32e100ccdb72
	go.uber.org/zap v1.16.0
	golang.org/x/lint v0.0.0-20201208152925-83fdc39ff7b5
	golang.org/x/mod v0.4.1 // indirect
	golang.org/x/tools v0.1.0 // indirect
	k8s.io/api v0.18.2
	k8s.io/apiextensions-apiserver v0.18.2
	k8s.io/apimachinery v0.18.2
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/code-generator v0.18.2
	sigs.k8s.io/controller-runtime v0.6.0
	sigs.k8s.io/controller-tools v0.4.1
	sigs.k8s.io/yaml v1.2.0

)

replace k8s.io/client-go => k8s.io/client-go v0.18.2
