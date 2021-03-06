// Copyright (c) 2020, 2021, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

package helm

import (
	"os/exec"
	"strings"

	vz_os "github.com/verrazzano/verrazzano/platform-operator/internal/util/os"
	"go.uber.org/zap"
)

// cmdRunner needed for unit tests
var runner vz_os.CmdRunner = vz_os.DefaultRunner{}

// Upgrade will upgrade a Helm release with the specified charts.
func Upgrade(log *zap.SugaredLogger, releaseName string, namespace string, chartDir string, overwriteYaml string) (stdout []byte, stderr []byte, err error) {
	// Helm upgrade command will apply the new chart, but use all the existing
	// overrides that we used during the install.
	args := []string{"upgrade", releaseName, chartDir}
	if namespace != "" {
		args = append(args, "--namespace")
		args = append(args, namespace)
	}

	if overwriteYaml != "" {
		args = append(args, "--reuse-values")
		args = append(args, "-f")
		args = append(args, overwriteYaml)
	}

	cmd := exec.Command("helm", args...)
	stdout, stderr, err = runner.Run(cmd)
	if err != nil {
		log.Errorf("helm upgrade for release %s failed with stderr: %s\n", releaseName, string(stderr))
		return stdout, stderr, err
	}

	//  Log upgrade output
	log.Infof("helm upgrade for release %s succeeded with stdout: %s\n", releaseName, string(stdout))
	return stdout, stderr, nil
}

// IsReleaseInstalled returns true if the release is installed
func IsReleaseInstalled(releaseName string, namespace string) (found bool, err error) {
	log := zap.S()

	args := []string{"status", releaseName}
	if namespace != "" {
		args = append(args, "--namespace")
		args = append(args, namespace)
	}
	cmd := exec.Command("helm", args...)
	_, stderr, err := runner.Run(cmd)
	if err == nil {
		return true, nil
	}
	if strings.Contains(string(stderr), "not found") {
		return false, nil
	}
	log.Errorf("helm status for release %s failed with stderr: %s\n", releaseName, string(stderr))
	return false, err
}

// SetCmdRunner sets the command runner as needed by unit tests
func SetCmdRunner(r vz_os.CmdRunner) {
	runner = r
}

// SetDefaultRunner sets the command runner to default
func SetDefaultRunner() {
	runner = vz_os.DefaultRunner{}
}
