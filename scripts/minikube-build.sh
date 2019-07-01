#!/bin/bash

eval $(minikube docker-env)
docker build . -t atlas-service-broker
