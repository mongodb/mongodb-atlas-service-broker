# MongoDB Atlas Service Broker

Implementation of the [Open Service Broker API](https://www.openservicebrokerapi.org/) for [MongoDB Atlas](https://www.mongodb.com/cloud/atlas). Deploy this service to easily manage Atlas instances!


## Configuration

Configuration is handled with environment variables.

### Atlas API

| Variable | Default | Description |
| -------- | ------- | ----------- |
| ATLAS_GROUP_ID | **Required** | Group in which to provision new clusters |
| ATLAS_PUBLIC_KEY | **Required** | Public part of the Atlas API key |
| ATLAS_PRIVATE_KEY | **Required** | Private part of the Atlas API key |
| ATLAS_BASE_URL | `https://cloud.mongodb.com` | Base URL used for Atlas API connections |

### Broker API Server

| Variable | Default | Description |
| -------- | ------- | ----------- |
| BROKER_HOST | `127.0.0.1` | Address which the broker server listens on |
| BROKER_PORT | `4000` | Port which the broker server listens on |
| BROKER_USERNAME | **Required** | Username for basic auth against broker |
| BROKER_PASSWORD | **Required** | Password for basic auth against broker |

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


## Development

**Do not clone this project to your $GOPATH**

This project uses Go modules which will be disabled if the project is built from the `$GOPATH`. If
the project is built inside the `$GOPATH` then Go will fetch the dependencies from there as well. This
could lead to incorrect versions and unreliable builds. When placed outside the `$GOPATH` dependencies will
automatically be installed when the project is built.


## Testing

The project contains both unit tests and integration tests against Atlas. The unit tests can be
found inside each package in `pkg/` and can be run with `go test ./pkg/...`.

The integration tests are also implemented as Go tests and are found in `test/`. Credentials
for connecting to the Atlas API should be passed as environment variables as specified in
[Configuration](#configuration). These tests can be run with `go test -timeout 1h ./test`. The
default timeout is 10 minutes which is normally too short for some of the tests, hence it's
recommended to raise the timeout to 1 hour. As part of the integration tests a MongoDB connection is
set up to test the generated credentials. For this test to not fail the testing host needs to be
whitelisted in Atlas.

Unit and integration tests can be run at once using `go test -timeout 1h ./...`. Remember
to pass the necessary environment variables and raise the timeout limit.

## Releasing

The release process consists of publishing a new Github release with attached binaries as well as publishing a Docker image to [quay.io](https://quay.io). Evergreen can automatically build and publish the artifacts based on a tagged commit.

1. Add a new annotated tag using `git tag -a vX.X.X`. Git will prompt for a message which later will be used for the Github release message.
2. Push the tag using `git push <remote> vX.X.X`.
3. Run `evergreen patch -v release -t release_github -t release_docker -y -f` and Evergreen will automatically complete the release.

## Adding third-party dependencies

Please include their license in the notices/ directory.
