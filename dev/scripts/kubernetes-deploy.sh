#!/bin/bash

NAMESPACE=${1:-default}
echo "Using namespace $NAMESPACE"

kubectl apply -f samples/kubernetes/deployment.yaml --namespace $NAMESPACE
kubectl apply -f samples/kubernetes/service.yaml --namespace $NAMESPACE
kubectl apply -f samples/kubernetes/auth-secret.yaml --namespace $NAMESPACE
kubectl apply -f samples/kubernetes/service-broker.yaml --namespace $NAMESPACE
