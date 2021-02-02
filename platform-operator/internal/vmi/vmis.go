// Copyright (C) 2020, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

// Handles creation/deletion of VMI CRs, based on a VerrazzanoBinding

package vmi

import (
	"context"
	"errors"
	"fmt"
	v1beta1v8o "github.com/verrazzano/verrazzano-crd-generator/pkg/apis/verrazzano/v1beta1"
	vmov1 "github.com/verrazzano/verrazzano-monitoring-operator/pkg/apis/vmcontroller/v1"
	vmoclientset "github.com/verrazzano/verrazzano-monitoring-operator/pkg/client/clientset/versioned"
	vmolisters "github.com/verrazzano/verrazzano-monitoring-operator/pkg/client/listers/vmcontroller/v1"
	"github.com/verrazzano/verrazzano/platform-operator/internal/constants"
	"github.com/verrazzano/verrazzano/platform-operator/internal/util/diff"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"strconv"
)

var createInstanceFunc = createInstance

// GetProfileBindingName will return the binding name based on the profile
// if the profile doesn't have a special binding name, the binding name supplied is returned
// for Dev profile the VMI system binding name is returned
func GetProfileBindingName(bindingName string) string {
	//if SharedVMIDefault() {
		return constants.VmiSystemBindingName
	//}
	//return bindingName
}

// GetManagedBindingLabels returns binding labels for managed cluster.
func GetManagedBindingLabels(binding *v1beta1v8o.VerrazzanoBinding, managedClusterName string) map[string]string {
	return map[string]string{constants.K8SAppLabel: constants.VerrazzanoGroup, constants.VerrazzanoBinding: binding.Name, constants.VerrazzanoCluster: managedClusterName}
}

// GetManagedLabelsNoBinding return labels for managed cluster with no binding.
func GetManagedLabelsNoBinding(managedClusterName string) map[string]string {
	return map[string]string{constants.K8SAppLabel: constants.VerrazzanoGroup, constants.VerrazzanoCluster: managedClusterName}
}

// GetManagedNamespaceForBinding return the namespace for a given binding.
func GetManagedNamespaceForBinding(binding *v1beta1v8o.VerrazzanoBinding) string {
	return fmt.Sprintf("%s-%s", constants.VerrazzanoPrefix, binding.Name)
}

// GetLocalBindingLabels returns binding labels for local cluster.
func GetLocalBindingLabels(binding *v1beta1v8o.VerrazzanoBinding) map[string]string {
	return map[string]string{constants.K8SAppLabel: constants.VerrazzanoGroup, constants.VerrazzanoBinding: binding.Name}
}

// GetManagedClusterNamespaceForSystem returns the system namespace for Verrazzano.
func GetManagedClusterNamespaceForSystem() string {
	return constants.VerrazzanoSystem
}

// GetVmiNameForBinding returns a Verrazzano Monitoring Instance name.
func GetVmiNameForBinding(bindingName string) string {
	return bindingName
}

// GetVmiURI returns a Verrazzano Monitoring Instance URI.
func GetVmiURI(bindingName string, verrazzanoURI string) string {
	return fmt.Sprintf("vmi.%s.%s", bindingName, verrazzanoURI)
}

// GetServiceAccountNameForSystem return the system service account for Verrazzano.
func GetServiceAccountNameForSystem() string {
	return constants.VerrazzanoSystem
}

// CreateUpdateVmi creates/updates Verrazzano Monitoring Instances for a given binding.
func CreateUpdateVmi(binding *v1beta1v8o.VerrazzanoBinding, vmoClientSet vmoclientset.Interface, vmiLister vmolisters.VerrazzanoMonitoringInstanceLister, verrazzanoURI string, enableMonitoringStorage string) error {
	zap.S().Debugf("Creating/updating Local (Management Cluster) VMI for VerrazzanoBinding %s", binding.Name)

	//if util.SharedVMIDefault() && !util.IsSystemProfileBindingName(binding.Name) {
	//	zap.S().Infof("Using shared VMI for binding %s", binding.Name)
	//	return nil
	//}

	// Construct the expected VMI
	newVmi, err := createInstanceFunc(binding, verrazzanoURI, enableMonitoringStorage)
	if err != nil {
		return err
	}

	// Create or update VMIs
	existingVmi, err := vmiLister.VerrazzanoMonitoringInstances(newVmi.Namespace).Get(newVmi.Name)
	if existingVmi != nil {
		newVmi.Spec.Grafana.Storage.PvcNames = existingVmi.Spec.Grafana.Storage.PvcNames
		newVmi.Spec.Prometheus.Storage.PvcNames = existingVmi.Spec.Prometheus.Storage.PvcNames
		newVmi.Spec.Elasticsearch.Storage.PvcNames = existingVmi.Spec.Elasticsearch.Storage.PvcNames
		specDiffs := diff.CompareIgnoreTargetEmpties(existingVmi, newVmi)
		if specDiffs != "" {
			zap.S().Infof("VMI %s : Spec differences %s", newVmi.Name, specDiffs)
			zap.S().Infof("Updating VMI %s", newVmi.Name)
			newVmi.ResourceVersion = existingVmi.ResourceVersion
			_, err = vmoClientSet.VerrazzanoV1().VerrazzanoMonitoringInstances(newVmi.Namespace).Update(context.TODO(), newVmi, metav1.UpdateOptions{})
		}
	} else {
		zap.S().Infof("Creating VMI %s", newVmi.Name)
		_, err = vmoClientSet.VerrazzanoV1().VerrazzanoMonitoringInstances(newVmi.Namespace).Create(context.TODO(), newVmi, metav1.CreateOptions{})
	}
	if err != nil {
		return err
	}
	return nil
}

// DeleteVmi deletes Verrazzano Monitoring Instances for a given binding.
func DeleteVmi(binding *v1beta1v8o.VerrazzanoBinding, vmoClientSet vmoclientset.Interface, vmiLister vmolisters.VerrazzanoMonitoringInstanceLister) error {
	zap.S().Infof("Deleting Local (Management Cluster) VMIs for VerrazzanoBinding %s", binding.Name)

	selector := labels.SelectorFromSet(map[string]string{constants.VerrazzanoBinding: binding.Name})

	existingVMIsList, err := vmiLister.VerrazzanoMonitoringInstances("").List(selector)
	if err != nil {
		return err
	}
	for _, existingVmi := range existingVMIsList {
		zap.S().Infof("Deleting VMI %s", existingVmi.Name)
		err := vmoClientSet.VerrazzanoV1().VerrazzanoMonitoringInstances(existingVmi.Namespace).Delete(context.TODO(), existingVmi.Name, metav1.DeleteOptions{})
		if err != nil {
			return err
		}
	}
	return nil
}

// Create a storage setting based on the specified value if monitoring storage is enabled
func createStorageOption(envSetting string, enableMonitoringStorageEnvFlag string) vmov1.Storage {
	storageSetting := vmov1.Storage{
		Size: "",
	}
	monitoringStorageEnabled, err := strconv.ParseBool(enableMonitoringStorageEnvFlag)
	if err != nil {
		zap.S().Errorf("Invalid storage setting: %s", enableMonitoringStorageEnvFlag)
	} else if monitoringStorageEnabled && len(envSetting) > 0 {
		storageSetting = vmov1.Storage{
			Size: envSetting,
		}
	}
	return storageSetting
}

// Constructs the necessary VerrazzanoMonitoringInstance for the given VerrazzanoBinding
func createInstance(binding *v1beta1v8o.VerrazzanoBinding, verrazzanoURI string, enableMonitoringStorage string) (*vmov1.VerrazzanoMonitoringInstance, error) {
	if verrazzanoURI == "" {
		return nil, errors.New("verrazzanoURI must not be empty")
	}

	bindingLabels := GetLocalBindingLabels(binding)

	bindingName := GetProfileBindingName(binding.Name)

	return &vmov1.VerrazzanoMonitoringInstance{
		ObjectMeta: metav1.ObjectMeta{
			Name:      GetVmiNameForBinding(binding.Name),
			Namespace: constants.VerrazzanoNamespace,
			Labels:    bindingLabels,
		},
		Spec: vmov1.VerrazzanoMonitoringInstanceSpec{
			URI:             GetVmiURI(bindingName, verrazzanoURI),
			AutoSecret:      true,
			SecretsName:     constants.VmiSecretName,
			CascadingDelete: true,
			Grafana: vmov1.Grafana{
				Enabled:             true,
				Storage:             createStorageOption(GetGrafanaDataStorageSize(), enableMonitoringStorage),
				DashboardsConfigMap: GetVmiNameForBinding(binding.Name) + "-dashboards",
				Resources: vmov1.Resources{
					RequestMemory: GetGrafanaRequestMemory(),
				},
			},
			IngressTargetDNSName: fmt.Sprintf("verrazzano-ingress.%s", verrazzanoURI),
			Prometheus: vmov1.Prometheus{
				Enabled: true,
				Storage: createStorageOption(GetPrometheusDataStorageSize(), enableMonitoringStorage),
				Resources: vmov1.Resources{
					RequestMemory: GetPrometheusRequestMemory(),
				},
			},
			Elasticsearch: vmov1.Elasticsearch{
				Enabled: true,
				Storage: createStorageOption(GetElasticsearchDataStorageSize(), enableMonitoringStorage),
				IngestNode: vmov1.ElasticsearchNode{
					Replicas: GetElasticsearchIngestNodeReplicas(),
					Resources: vmov1.Resources{
						RequestMemory: GetElasticsearchIngestNodeRequestMemory(),
					},
				},
				MasterNode: vmov1.ElasticsearchNode{
					Replicas: GetElasticsearchMasterNodeReplicas(),
					Resources: vmov1.Resources{
						RequestMemory: GetElasticsearchMasterNodeRequestMemory(),
					},
				},
				DataNode: vmov1.ElasticsearchNode{
					Replicas: GetElasticsearchDataNodeReplicas(),
					Resources: vmov1.Resources{
						RequestMemory: GetElasticsearchDataNodeRequestMemory(),
					},
				},
			},
			Kibana: vmov1.Kibana{
				Enabled: true,
				Resources: vmov1.Resources{
					RequestMemory: GetKibanaRequestMemory(),
				},
			},
			ServiceType: "ClusterIP",
		},
	}, nil
}
