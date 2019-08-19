#!/bin/bash

eval $(minikube docker-env)
docker build . -t quay.io/mongodb/mongodb-atlas-service-broker
