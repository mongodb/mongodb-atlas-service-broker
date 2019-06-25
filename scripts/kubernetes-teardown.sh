#!/bin/bash

NAMESPACE=${1:-default}
echo "Using namespace $NAMESPACE"

kubectl delete -f scripts/kubernetes/service-broker.yaml --namespace $NAMESPACE
kubectl delete -f scripts/kubernetes/auth-secret.yaml --namespace $NAMESPACE
kubectl delete -f scripts/kubernetes/service.yaml --namespace $NAMESPACE
kubectl delete -f scripts/kubernetes/deployment.yaml --namespace $NAMESPACE
