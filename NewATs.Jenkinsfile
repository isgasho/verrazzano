// Copyright (c) 2020, 2021, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.

def DOCKER_IMAGE_TAG

pipeline {
    options {
        skipDefaultCheckout true
    }

    agent {
       docker {
            image "${RUNNER_DOCKER_IMAGE}"
            args "${RUNNER_DOCKER_ARGS}"
            registryUrl "${RUNNER_DOCKER_REGISTRY_URL}"
            registryCredentialsId 'ocir-pull-and-push-account'
        }
    }

    parameters {
        booleanParam (description: 'Whether to run example tests', name: 'RUN_EXAMPLE_TESTS', defaultValue: true)
        booleanParam (description: 'Whether to dump k8s cluster on success (off by default can be useful to capture for comparing to failed cluster)', name: 'DUMP_K8S_CLUSTER_ON_SUCCESS', defaultValue: false)
    }

    environment {
        DOCKER_REPO = 'ghcr.io'
        DOCKER_CREDS = credentials('github-packages-credentials-rw')
        GOPATH = '/home/opc/go'
        GO_REPO_PATH = "${GOPATH}/src/github.com/verrazzano"
        IMAGE_PULL_SECRET = 'verrazzano-container-registry'
        OCR_CREDS = credentials('ocr-pull-and-push-account')
        OCR_REPO = 'container-registry.oracle.com'
        DOCKER_PLATFORM_CI_IMAGE_NAME = 'verrazzano-platform-operator-jenkins'
        DOCKER_PLATFORM_PUBLISH_IMAGE_NAME = 'verrazzano-platform-operator'
        DOCKER_PLATFORM_IMAGE_NAME = "${env.BRANCH_NAME == 'develop' || env.BRANCH_NAME == 'master' ? env.DOCKER_PLATFORM_PUBLISH_IMAGE_NAME : env.DOCKER_PLATFORM_CI_IMAGE_NAME}"
        INSTALL_CONFIG_FILE_KIND = "./tests/e2e/config/scripts/install-verrazzano-kind.yaml"
        CLUSTER_NAME = 'verrazzano'
        POST_DUMP_FAILED_FILE = "${WORKSPACE}/post_dump_failed_file.tmp"
        NETRC_FILE = credentials('netrc')
        KUBECONFIG = "${WORKSPACE}/test_kubeconfig"


        SERVICE_KEY = credentials('PAGERDUTY_SERVICE_KEY')
    }

    stages {
        stage('Clean workspace and checkout') {
            steps {
                sh """
                    echo "${NODE_LABELS}"
                """

                script {
                    checkout scm
                }
                sh """
                    cp -f "${NETRC_FILE}" $HOME/.netrc
                    chmod 600 $HOME/.netrc
                """

                script {
                try {
                sh """
                    echo "${DOCKER_CREDS_PSW}" | docker login ${env.DOCKER_REPO} -u ${DOCKER_CREDS_USR} --password-stdin
                """
            } catch(error) {
                echo "docker login failed, retrying after sleep"
                retry(4) {
                sleep(30)
                sh """
                    echo "${DOCKER_CREDS_PSW}" | docker login ${env.DOCKER_REPO} -u ${DOCKER_CREDS_USR} --password-stdin
                """
                }
            }
            }
                script {
                try {
                sh """
                    echo "${OCR_CREDS_PSW}" | docker login -u ${OCR_CREDS_USR} ${OCR_REPO} --password-stdin
                """
            } catch(error) {
                echo "OCR docker login failed, retrying after sleep"
                retry(4) {
                sleep(30)
                sh """
                    echo "${OCR_CREDS_PSW}" | docker login -u ${OCR_CREDS_USR} ${OCR_REPO} --password-stdin
                """
                }
            }
            }
                sh """
                    rm -rf ${GO_REPO_PATH}/verrazzano
                    mkdir -p ${GO_REPO_PATH}/verrazzano
                    tar cf - . | (cd ${GO_REPO_PATH}/verrazzano/ ; tar xf -)
                """

                script {
                    def props = readProperties file: '.verrazzano-development-version'
                    VERRAZZANO_DEV_VERSION = props['verrazzano-development-version']
                    TIMESTAMP = sh(returnStdout: true, script: "date +%Y%m%d%H%M%S").trim()
                    SHORT_COMMIT_HASH = sh(returnStdout: true, script: "git rev-parse --short HEAD").trim()
                    DOCKER_IMAGE_TAG = "${VERRAZZANO_DEV_VERSION}-${TIMESTAMP}-${SHORT_COMMIT_HASH}"
                }
            }
        }

        stage('New Acceptance Tests') {
            stages {
                 stage('Create Kind Cluster for Test Suite') {
                    steps {
                        sh """
                            cd ${GO_REPO_PATH}/verrazzano/platform-operator
                            make create-cluster
                        """
                    }
                }

                stage('Create Image Pull Secrets') {
                    steps {
                        sh """
                            # Create image pull secret for Verrazzano docker images
                            cd ${GO_REPO_PATH}/verrazzano
                            ./tests/e2e/config/scripts/create-image-pull-secret.sh "${IMAGE_PULL_SECRET}" "${DOCKER_REPO}" "${DOCKER_CREDS_USR}" "${DOCKER_CREDS_PSW}"
                            ./tests/e2e/config/scripts/create-image-pull-secret.sh github-packages "${DOCKER_REPO}" "${DOCKER_CREDS_USR}" "${DOCKER_CREDS_PSW}"
                            ./tests/e2e/config/scripts/create-image-pull-secret.sh ocr "${OCR_REPO}" "${OCR_CREDS_USR}" "${OCR_CREDS_PSW}"
                        """
                    }
                }

                stage('Install Platform Operator') {
                    steps {
                        sh """
                            cd ${GO_REPO_PATH}/verrazzano

                            # Configure the deployment file to use an image pull secret for branches that have private images
                            if [ "${env.BRANCH_NAME}" == "master" ] || [ "${env.BRANCH_NAME}" == "develop" ]; then
                                echo "Using operator.yaml from Verrazzano repo"
                                cp platform-operator/deploy/operator.yaml /tmp/operator.yaml
                            else
                                echo "Generating operator.yaml based on image name provided: ${DOCKER_PLATFORM_IMAGE_NAME}:${DOCKER_IMAGE_TAG}"
                                ./tests/e2e/config/scripts/process_operator_yaml.sh platform-operator "${DOCKER_PLATFORM_IMAGE_NAME}:${DOCKER_IMAGE_TAG}"
                            fi

                            # Install the verrazzano-platform-operator
                            cat /tmp/operator.yaml
                            kubectl apply -f /tmp/operator.yaml

                            # make sure ns exists
                            ./tests/e2e/config/scripts/check_verrazzano_ns_exists.sh verrazzano-install

                            # create secret in verrazzano-install ns
                            ./tests/e2e/config/scripts/create-image-pull-secret.sh "${IMAGE_PULL_SECRET}" "${DOCKER_REPO}" "${DOCKER_CREDS_USR}" "${DOCKER_CREDS_PSW}" "verrazzano-install"

                            # Configure the custom resource to install verrazzano on Kind
                            echo "Installing yq"
                            GO111MODULE=on go get github.com/mikefarah/yq/v4
                            export PATH=${HOME}/go/bin:${PATH}
                            ./tests/e2e/config/scripts/process_kind_install_yaml.sh ${INSTALL_CONFIG_FILE_KIND}
                        """
                    }
                }

                stage('Install Verrazzano') {
                    steps {
                        sh """
                            echo "Waiting for Operator to be ready"
                            cd ${GO_REPO_PATH}/verrazzano
                            kubectl -n verrazzano-install rollout status deployment/verrazzano-platform-operator

                            echo "Installing Verrazzano on Kind"
                            kubectl apply -f ${INSTALL_CONFIG_FILE_KIND}

                            # wait for Verrazzano install to complete
                            ./tests/e2e/config/scripts/wait-for-verrazzano-install.sh

                            # Hack
                            # OCIR images don't work with KIND.
                            # Coherence image doesn't get pulled correctly in KIND.
                            docker pull container-registry.oracle.com/middleware/coherence:12.2.1.4.0
                            kind load docker-image --name ${CLUSTER_NAME} container-registry.oracle.com/middleware/coherence:12.2.1.4.0
                            # The ToDoList example image currently cannot be pulled in KIND.
                            docker pull container-registry.oracle.com/verrazzano/example-todo:0.8.0
                            kind load docker-image --name ${CLUSTER_NAME} container-registry.oracle.com/verrazzano/example-todo:0.8.0
                        """
                    }
                    post {
                        always {
                            sh """
                                ## dump out install logs
                                mkdir -p ${WORKSPACE}/verrazzano/platform-operator/scripts/install/build/logs
                                kubectl logs --selector=job-name=verrazzano-install-my-verrazzano > ${WORKSPACE}/verrazzano/platform-operator/scripts/install/build/logs/verrazzano-install.log --tail -1
                                kubectl describe pod --selector=job-name=verrazzano-install-my-verrazzano > ${WORKSPACE}/verrazzano/platform-operator/scripts/install/build/logs/verrazzano-install-job-pod.out
                                echo "Verrazzano Installation logs dumped to verrazzano-install.log"
                                echo "Verrazzano Install pod description dumped to verrazzano-install-job-pod.out"
                                echo "------------------------------------------"
                            """
                        }
                    }
                }

                stage('Run Acceptance Tests') {
                    environment {
                        TEST_ENV = "KIND"
                    }
                    parallel {
                        stage('verify-install') {
                            steps {
                                runGinkgoRandomize('verify-install')
                            }
                        }
                        stage('verify-infra restapi') {
                            steps {
                                runGinkgo('verify-infra/restapi')
                            }
                        }
                        stage('verify-infra oam') {
                            steps {
                                runGinkgo('verify-infra/oam')
                            }
                        }
                        stage('verify-infra vmi') {
                            steps {
                                runGinkgo('verify-infra/vmi')
                            }
                        }
                        stage('examples') {
                            when {
                                expression {params.RUN_EXAMPLE_TESTS == true}
                            }
                            steps {
                                runGinkgo('examples/todo-list')
                                runGinkgo('examples/sock-shop')
                            }
                        }
                    }
                }

            }
            post {
                failure {
                    dumpK8sCluster('new-acceptance-tests-cluster-dump.tar.gz')
                }
                success {
                    script {
                        if (params.DUMP_K8S_CLUSTER_ON_SUCCESS == true) {
                            dumpK8sCluster('new-acceptance-tests-cluster-dump.tar.gz')
                        }
                    }
                }
            }
        }
    }

    post {
        always {
            dumpVerrazzanoSystemPods()
            dumpCattleSystemPods()
            dumpNginxIngressControllerLogs()
            dumpVerrazzanoPlatformOperatorLogs()
            dumpVerrazzanoApplicationOperatorLogs()
            dumpOamKubernetesRuntimeLogs()

            archiveArtifacts artifacts: '**/coverage.html,**/logs/**,**/verrazzano_images.txt,**/*cluster-dump.tar.gz', allowEmptyArchive: true
            junit testResults: '**/*test-result.xml', allowEmptyResults: true

            sh """
                cd ${GO_REPO_PATH}/verrazzano/platform-operator
                make delete-cluster
                if [ -f ${POST_DUMP_FAILED_FILE} ]; then
                  echo "Failures seen during dumping of artifacts, treat post as failed"
                  exit 1
                fi
            """
            deleteDir()
        }
        failure {
            mail to: "${env.BUILD_NOTIFICATION_TO_EMAIL}", from: "${env.BUILD_NOTIFICATION_FROM_EMAIL}",
            subject: "Verrazzano: ${env.JOB_NAME} - Failed",
            body: "Job Failed - \"${env.JOB_NAME}\" build: ${env.BUILD_NUMBER}\n\nView the log at:\n ${env.BUILD_URL}\n\nBlue Ocean:\n${env.RUN_DISPLAY_URL}"
            script {
                if (env.JOB_NAME == "verrazzano/master" || env.JOB_NAME == "verrazzano/develop") {
                    pagerduty(resolve: false, serviceKey: "$SERVICE_KEY", incDescription: "Verrazzano: ${env.JOB_NAME} - Failed", incDetails: "Job Failed - \"${env.JOB_NAME}\" build: ${env.BUILD_NUMBER}\n\nView the log at:\n ${env.BUILD_URL}\n\nBlue Ocean:\n${env.RUN_DISPLAY_URL}")
                    slackSend ( message: "Job Failed - \"${env.JOB_NAME}\" build: ${env.BUILD_NUMBER}\n\nView the log at:\n ${env.BUILD_URL}\n\nBlue Ocean:\n${env.RUN_DISPLAY_URL}" )
                }
            }
        }
    }
}

def runGinkgoRandomize(testSuitePath) {
    catchError(buildResult: 'FAILURE', stageResult: 'FAILURE') {
        sh """
            cd ${GO_REPO_PATH}/verrazzano/tests/e2e
            ginkgo -p --randomizeAllSpecs -v -keepGoing --noColor ${testSuitePath}/...
        """
    }
}

def runGinkgo(testSuitePath) {
    catchError(buildResult: 'FAILURE', stageResult: 'FAILURE') {
        sh """
            cd ${GO_REPO_PATH}/verrazzano/tests/e2e
            ginkgo -v -keepGoing --noColor ${testSuitePath}/...
        """
    }
}

def dumpK8sCluster(archiveFilePath) {
    sh """
        ${GO_REPO_PATH}/verrazzano/tools/scripts/k8s-dump-cluster.sh -z ${archiveFilePath}
    """
}

def dumpVerrazzanoSystemPods() {
    sh """
        cd ${GO_REPO_PATH}/verrazzano/platform-operator
        export DIAGNOSTIC_LOG="${WORKSPACE}/verrazzano/platform-operator/scripts/install/build/logs/verrazzano-system-pods.log"
        ./scripts/install/k8s-dump-objects.sh -o pods -n verrazzano-system -m "verrazzano system pods" || echo "failed" > ${POST_DUMP_FAILED_FILE}
        export DIAGNOSTIC_LOG="${WORKSPACE}/verrazzano/platform-operator/scripts/install/build/logs/verrazzano-system-certs.log"
        ./scripts/install/k8s-dump-objects.sh -o cert -n verrazzano-system -m "verrazzano system certs" || echo "failed" > ${POST_DUMP_FAILED_FILE}
        export DIAGNOSTIC_LOG="${WORKSPACE}/verrazzano/platform-operator/scripts/install/build/logs/verrazzano-system-kibana.log"
        ./scripts/install/k8s-dump-objects.sh -o pods -n verrazzano-system -r "vmi-system-kibana-*" -m "verrazzano system kibana log" -l -c kibana || echo "failed" > ${POST_DUMP_FAILED_FILE}
        export DIAGNOSTIC_LOG="${WORKSPACE}/verrazzano/platform-operator/scripts/install/build/logs/verrazzano-system-es-master.log"
        ./scripts/install/k8s-dump-objects.sh -o pods -n verrazzano-system -r "vmi-system-es-master-*" -m "verrazzano system kibana log" -l -c es-master || echo "failed" > ${POST_DUMP_FAILED_FILE}
    """
}

def dumpCattleSystemPods() {
    sh """
        cd ${GO_REPO_PATH}/verrazzano/platform-operator
        export DIAGNOSTIC_LOG="${WORKSPACE}/verrazzano/platform-operator/scripts/install/build/logs/cattle-system-pods.log"
        ./scripts/install/k8s-dump-objects.sh -o pods -n cattle-system -m "cattle system pods" || echo "failed" > ${POST_DUMP_FAILED_FILE}
        export DIAGNOSTIC_LOG="${WORKSPACE}/verrazzano/platform-operator/scripts/install/build/logs/rancher.log"
        ./scripts/install/k8s-dump-objects.sh -o pods -n cattle-system -r "rancher-*" -m "Rancher logs" -l || echo "failed" > ${POST_DUMP_FAILED_FILE}
    """
}

def dumpNginxIngressControllerLogs() {
    sh """
        cd ${GO_REPO_PATH}/verrazzano/platform-operator
        export DIAGNOSTIC_LOG="${WORKSPACE}/verrazzano/platform-operator/scripts/install/build/logs/nginx-ingress-controller.log"
        ./scripts/install/k8s-dump-objects.sh -o pods -n ingress-nginx -r "nginx-ingress-controller-*" -m "Nginx Ingress Controller" -l || echo "failed" > ${POST_DUMP_FAILED_FILE}
    """
}

def dumpVerrazzanoPlatformOperatorLogs() {
    sh """
        ## dump out verrazzano-platform-operator logs
        mkdir -p ${WORKSPACE}/verrazzano-platform-operator/logs
        kubectl -n verrazzano-install logs --selector=app=verrazzano-platform-operator > ${WORKSPACE}/verrazzano-platform-operator/logs/verrazzano-platform-operator-pod.log --tail -1 || echo "failed" > ${POST_DUMP_FAILED_FILE}
        kubectl -n verrazzano-install describe pod --selector=app=verrazzano-platform-operator > ${WORKSPACE}/verrazzano-platform-operator/logs/verrazzano-platform-operator-pod.out || echo "failed" > ${POST_DUMP_FAILED_FILE}
        echo "verrazzano-platform-operator logs dumped to verrazzano-platform-operator-pod.log"
        echo "verrazzano-platform-operator pod description dumped to verrazzano-platform-operator-pod.out"
        echo "------------------------------------------"
    """
}

def dumpVerrazzanoApplicationOperatorLogs() {
    sh """
        ## dump out verrazzano-application-operator logs
        mkdir -p ${WORKSPACE}/verrazzano-application-operator/logs
        kubectl -n verrazzano-system logs --selector=app=verrazzano-application-operator > ${WORKSPACE}/verrazzano-application-operator/logs/verrazzano-application-operator-pod.log --tail -1 || echo "failed" > ${POST_DUMP_FAILED_FILE}
        kubectl -n verrazzano-system describe pod --selector=app=verrazzano-application-operator > ${WORKSPACE}/verrazzano-application-operator/logs/verrazzano-application-operator-pod.out || echo "failed" > ${POST_DUMP_FAILED_FILE}
        echo "verrazzano-application-operator logs dumped to verrazzano-application-operator-pod.log"
        echo "verrazzano-application-operator pod description dumped to verrazzano-application-operator-pod.out"
        echo "------------------------------------------"
    """
}

def dumpOamKubernetesRuntimeLogs() {
    sh """
        ## dump out oam-kubernetes-runtime logs
        mkdir -p ${WORKSPACE}/oam-kubernetes-runtime/logs
        kubectl -n verrazzano-system logs --selector=app.kubernetes.io/instance=oam-kubernetes-runtime > ${WORKSPACE}/oam-kubernetes-runtime/logs/oam-kubernetes-runtime-pod.log --tail -1 || echo "failed" > ${POST_DUMP_FAILED_FILE}
        kubectl -n verrazzano-system describe pod --selector=app.kubernetes.io/instance=oam-kubernetes-runtime > ${WORKSPACE}/verrazzano-application-operator/logs/oam-kubernetes-runtime-pod.out || echo "failed" > ${POST_DUMP_FAILED_FILE}
        echo "verrazzano-application-operator logs dumped to oam-kubernetes-runtime-pod.log"
        echo "verrazzano-application-operator pod description dumped to oam-kubernetes-runtime-pod.out"
        echo "------------------------------------------"
    """
}