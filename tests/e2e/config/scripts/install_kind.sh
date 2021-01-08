#!/bin/bash
#
# Copyright (c) 2020, 2021, Oracle and/or its affiliates.
# Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.
#

set -e
SCRIPT_DIR=$(cd $(dirname "$0"); pwd -P)
ACC_TESTS_ROOT=${SCRIPT_DIR}/..
if [ -z "${KUBECONFIG}" ] ; then
    echo "KUBECONFIG env var must be set!"
    exit 1
fi

if [ -z "${OCR_CREDS_USR}" ] ; then
    echo "OCR_CREDS_USR env var must be set!"
    exit 1
fi

if [ -z "${OCR_CREDS_PSW}" ] ; then
    echo "OCR_CREDS_PSW env var must be set!"
    exit 1
fi

cd $ACC_TESTS_ROOT

echo "${OCR_CREDS_PSW}" | docker login -u ${OCR_CREDS_USR} ${OCR_REPO} --password-stdin

echo "${DOCKER_CREDS_PSW}" | docker login -u ${DOCKER_CREDS_USR} ${DOCKER_REPO} --password-stdin

kind delete cluster --name="${CLUSTER_NAME}" --kubeconfig "${KUBECONFIG}"

KIND_IMAGE="kindest/node:v1.17.0@sha256:9512edae126da271b66b990b6fff768fbb7cd786c7d39e86bdf55906352fdf62"

kind create cluster \
        --wait 30s \
        --image ${KIND_IMAGE} \
        --name ${CLUSTER_NAME} \
        --config ${SCRIPT_DIR}/kind-config.yaml \
        --kubeconfig ${KUBECONFIG}

cat ${KUBECONFIG} | grep server
# this ugly looking line of code will get the ip address of the container running the kube apiserver
# and update the kubeconfig file to point to that address, instead of localhost
sed -i -e "s|127.0.0.1.*|`docker inspect ${CLUSTER_NAME}-control-plane | jq '.[].NetworkSettings.IPAddress' | sed 's/"//g'`:6443|g" ${KUBECONFIG}
cat ${KUBECONFIG} | grep server

# Hack
# OCIR images don't work with KIND.
# Coherence image doesn't get pulled correctly in KIND.
docker pull container-registry.oracle.com/middleware/coherence:12.2.1.4.0
kind load docker-image --name ${CLUSTER_NAME} container-registry.oracle.com/middleware/coherence:12.2.1.4.0
