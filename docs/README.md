# Development

The broker is entirely written in Go and consists of a single executable, `main.go`, which makes use of two packages, `pkg/broker` and `pkg/atlas`. The executable runs an HTTP server which conforms to the [Open Service Broker API spec](https://github.com/openservicebrokerapi/servicebroker/blob/master/spec.md).

The server is managed by a third-party library called [`brokerapi`](https://github.com/pivotal-cf/brokerapi). This library exposes a `ServerBroker` interface which we implement with `Broker` in `pkg/broker`. `pkg/atlas` contains a client for the Atlas API and `Broker` uses that client to translate incoming service broker requests to Atlas API calls.

**Do not clone this project to your $GOPATH.** This project uses Go modules which will be disabled if the project is built from the `$GOPATH`. If the project is built inside the `$GOPATH` then Go will fetch the dependencies from there as well. This could lead to incorrect versions and unreliable builds. When placed outside the `$GOPATH` dependencies will automatically be installed when the project is built.

## Testing

The project contains both unit tests and integration tests against Atlas. The unit tests can be found inside each package in `pkg/` and can be run with `go test ./pkg/...`.

The integration tests are also implemented as Go tests and are found in `test/`. Credentials for connecting to the Atlas API should be passed as environment variables `ATLAS_BASE_URL`, `ATLAS_GROUP_ID`, `ATLAS_PUBLIC_KEY`, and `ATLAS_PRIVATE_KEY`. These tests can be run with `go test -timeout 1h ./test`. Go test has a default timeout of 10 minutes which is normally too short for some of the tests, hence it's recommended to raise the timeout to 1 hour. As part of the integration tests a MongoDB connection is set up to test the generated credentials. For this test to not fail the testing host needs to be whitelisted in Atlas.

Unit and integration tests can be run at once using `go test -timeout 1h ./...`. Remember to pass the necessary environment variables and raise the timeout limit.

## Releasing

The release process consists of publishing a new Github release with attached binaries as well as publishing a Docker image to [quay.io](https://quay.io). Evergreen can automatically build and publish the artifacts based on a tagged commit. To perform a release the [Evergreen CLI](https://github.com/evergreen-ci/evergreen/wiki/Using-the-Command-Line-Tool) needs to be installed.

1. Add a new annotated tag using `git tag -a vX.X.X`. Git will prompt for a message which later will be used for the Github release message.
2. Push the tag using `git push <remote> vX.X.X`.
3. Run `evergreen patch -v release -t release_github -t release_docker -y -f` and Evergreen will automatically complete the release.

## Adding third-party dependencies

Please include their license in the notices/ directory.

## Testing in Kubernetes

Follow these steps to test the broker in a Kubernetes cluster. For local testing we recommend using [minikube](https://kubernetes.io/docs/setup/learning-environment/minikube/). We also recommend using the [service catalog CLI](https://github.com/kubernetes-sigs/service-catalog/blob/master/docs/cli.md) (`svcat`) to control the service catalog.

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
