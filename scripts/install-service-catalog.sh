#!/bin/bash

# Add the service catalog repo to helm
helm repo add svc-cat https://svc-catalog-charts.storage.googleapis.com

# Install the service catalog
helm install svc-cat/catalog \
    --name catalog --namespace catalog
