package util

import (
	"context"
	"github.com/onsi/ginkgo"
	apixv1beta1client "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"os"
	"path/filepath"
	"strings"
)

// GetKubeConfig will get the kubeconfig from the environment variable KUBECONFIG, if set, or else from $HOME/.kube/config
func GetKubeConfig() *restclient.Config {
	kubeconfig := ""
	// if the KUBECONFIG environment variable is set, use that
	kubeconfigEnvVar := os.Getenv("KUBECONFIG")
	if len(kubeconfigEnvVar) > 0 {
		kubeconfig = kubeconfigEnvVar
	} else if home := homedir.HomeDir(); home != "" {
		// next look for $HOME/.kube/config
		kubeconfig = filepath.Join(home, ".kube", "config")
	} else {
		// give up
		ginkgo.Fail("Could not find kube")
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		ginkgo.Fail("Could not get current context from kubeconfig " + kubeconfig)
	}
	return config
}

// DoesCRDExist returns whether a CRD with the given name exists for the cluster
func DoesCRDExist(crdName string) bool {
	// use the current context in the kubeconfig
	config := GetKubeConfig()

	apixClient, err := apixv1beta1client.NewForConfig(config)
	if err != nil {
		ginkgo.Fail("Could not get apix client")
	}

	// ignoring error for now
	crds, _ := apixClient.CustomResourceDefinitions().List(context.TODO(), metav1.ListOptions{})

	for i := range crds.Items {
		if strings.Compare(crds.Items[i].ObjectMeta.Name, crdName) == 0 {
			return true
		}
	}

	return false
}
