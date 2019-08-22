#!/usr/bin/env bash

if [ -z "$2" ]; then
  echo 'Usage: ./dev/scripts/evergreen-kubernetes-setup.sh BIN_PATH DOCKER_IMAGE' > /dev/stderr
  exit 1
fi

bin=$1
docker_image=$2

export PATH="$bin:$PATH"

set -eou pipefail

echo "Install kubectl"
curl -L -o $bin/kubectl https://storage.googleapis.com/kubernetes-release/release/$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)/bin/linux/amd64/kubectl
chmod +x $bin/kubectl

echo "Install kind"
curl -L -o $bin/kind https://github.com/kubernetes-sigs/kind/releases/download/v0.4.0/kind-linux-amd64
chmod +x $bin/kind

echo "Start cluster"
kind create cluster
KUBECONFIG=$(kind get kubeconfig-path)
export KUBECONFIG

echo "Install Helm"
curl -LO https://get.helm.sh/helm-v2.14.3-linux-amd64.tar.gz
tar -zxvf helm-v2.14.3-linux-amd64.tar.gz
mv linux-amd64/helm $bin/helm

echo "Setup helm"
kubectl --namespace kube-system create serviceaccount tiller
kubectl create clusterrolebinding tiller --clusterrole cluster-admin --serviceaccount=kube-system:tiller
helm init --service-account tiller --wait

echo "Install Service Catalog extension"
helm repo add svc-cat https://svc-catalog-charts.storage.googleapis.com
helm install svc-cat/catalog --name service-catalog --namespace catalog
kubectl rollout status --watch deployment/service-catalog-catalog-apiserver --namespace=catalog

echo "Add Docker image to cluster"
docker image tag $docker_image $docker_image:e2e-test
kind load docker-image $docker_image:e2e-test
