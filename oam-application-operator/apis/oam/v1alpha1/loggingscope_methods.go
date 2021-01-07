// Copyright (c) 2020, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package v1alpha1

import (
	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/oam-kubernetes-runtime/pkg/oam"
)

// Ensure that LoggingScope adheres to Scope interface.
var _ oam.Scope = &LoggingScope{}

// GetCondition of this LoggingScope.
func (ls *LoggingScope) GetCondition(ct runtimev1alpha1.ConditionType) runtimev1alpha1.Condition {
	return ls.Status.GetCondition(ct)
}

// SetConditions of this LoggingScope.
func (ls *LoggingScope) SetConditions(c ...runtimev1alpha1.Condition) {
	ls.Status.SetConditions(c...)
}

// GetWorkloadReferences of this LoggingScope.
func (ls *LoggingScope) GetWorkloadReferences() []runtimev1alpha1.TypedReference {
	return ls.Spec.WorkloadReferences
}

// AddWorkloadReference of this LoggingScope.
func (ls *LoggingScope) AddWorkloadReference(r runtimev1alpha1.TypedReference) {
	ls.Spec.WorkloadReferences = append(ls.Spec.WorkloadReferences, r)
}
