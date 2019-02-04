#!/bin/bash
#
# Copyright 2018, OpenCensus Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http:#www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

[ -z ${CLUSTER_NAME} ] && CLUSTER_NAME=""
[ -z ${IMAGE} ] && IMAGE=gcr.io/prometheus-to-sd/opencensus-operator
[ -z ${NAMESPACE} ] && NAMESPACE=default

export CLUSTER_NAME
export IMAGE
export NAMESPACE

set -e
set -u

dir="$( dirname "${BASH_SOURCE[0]}" )"
tmpdir="${dir}/_tmp"

rm -rf "${tmpdir}"
mkdir "${tmpdir}"

${dir}/../third_party/kube-mutating-webhook-tutorial/create-signed-cert.sh \
    --service opencensus-pod-autoconf \
    --namespace "${NAMESPACE}" \
    --secret opencensus-pod-autoconf

# We have to provide the certificate authority in the webhook configuration. Since we create a cert
# signed by the Kubernetes cluster, we include the CA used in the process.
export CA_BUNDLE=$(kubectl get cm -n kube-system extension-apiserver-authentication \
    -o=jsonpath='{.data.client-ca-file}' | base64 | tr -d '\n')


envsubst < "${dir}/mutatingwebhook.yaml" > "${tmpdir}/mutatingwebhook.yaml"
envsubst < "${dir}/deployment.yaml" > "${tmpdir}/deployment.yaml"
envsubst < "${dir}/service.yaml" > "${tmpdir}/service.yaml"

kubectl apply -f "${tmpdir}"
