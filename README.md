# MongoDB Atlas Service Broker

WIP implementation of the [Open Service Broker API](https://www.openservicebrokerapi.org/) for [MongoDB Atlas](https://www.mongodb.com/cloud/atlas).

## Development

**Do not clone this project to your $GOPATH**

This project uses Go modules which will be disabled if the project is built from the `$GOPATH`. If
the project is built inside the `$GOPATH` then Go will fetch the dependencies from there as well. This
could lead to incorrect versions and unreliable builds. When placed outside the `$GOPATH` dependencies will
automatically be installed when the project is built.

## Configuration

Configuration is handled with environment variables.

### Broker API Server

- `BROKER_HOST`
- `BROKER_PORT`

### Atlas API

- `ATLAS_BASE_URL`
- `ATLAS_GROUP_ID`
- `ATLAS_PUBLIC_KEY`
- `ATLAS_PRIVATE_KEY`

## Deploying to Kubernetes

The service broker can be deployed to Kubernetes by following these steps:

1. Run `scripts/install-service-catalog.sh` to install the service catalog extension in Kubernetes.
   Make sure you have Helm installed and configured before running.
2. Make sure the Service Catalog extension is installed in Kubernetes. Installation instructions can
   be found in the [Kubernetes docs](https://kubernetes.io/docs/tasks/service-catalog/install-service-catalog-using-helm/).
3. Build the Dockerfile and make the resulting image available in your cluster. If you are using
   Minikube `scripts/minikube-build.sh` can be used to build the image using Minikube's Docker
   daemon.
4. Create a secret called `atlas-api` containing the following keys:
   - `base-url`for the Atlas API
   - `group-id` for the project under which clusters should be deployed
   - `public-key`for the API key
   - `private-key`for the API key
5. Deploy the service broker by running `scripts/kubernetes-deploy.sh <namespace>`. This will create
   a new deployment and a service of the image from step 2. The script will also deploy the actual service broker resource with the
   name `atlas-service-broker`.
6. Make sure the broker is ready by running `svcat get brokers`.
7. A new instance can be provisioned by running `kubectl create -f
   scripts/kubernetes/instance.yaml`. The instance will be given the name `atlas-cluster-instance`
   and its status can be checked using `svcat get instances`.
8. Once the instance is up and running a binding can be created to gain access. A binding named
   `atlas-cluster-binding` can be created by running `kubectl create -f
   script/kubernetes/binding.yaml`. The binding credentials will automatically be stored in a secret
   of the same name.
9. After use all bindings can be removed by running `svcat unbind atlas-cluser-instance` and the
   cluster can be deprovisioned using `svcat deprovision atlas-cluster-instance`.
10. Run `scripts/kubernetes-teardown.sh <namespace>` to fully remove the service broker.

