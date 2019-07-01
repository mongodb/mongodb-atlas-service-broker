#!/bin/bash

NAMESPACE=${1:-default}
echo "Using namespace $NAMESPACE"

kubectl apply -f scripts/kubernetes/deployment.yaml --namespace $NAMESPACE
kubectl apply -f scripts/kubernetes/service.yaml --namespace $NAMESPACE
kubectl apply -f scripts/kubernetes/auth-secret.yaml --namespace $NAMESPACE
kubectl apply -f scripts/kubernetes/service-broker.yaml --namespace $NAMESPACE
