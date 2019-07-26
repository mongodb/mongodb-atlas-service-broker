# ATLAS Service Broker
Setting up the service broker by using a local Kubernetes cluster through Minikube. This guide is meant for demo purposes, and the docs team.

## Getting Started
These instructions are split in two parts. One for installing all the necessary tools and the other for running it. After following all the steps, the project is up and running on your local machine for development and testing purposes. 

### Installing the tools

##### Install Homebrew on local machine
1. Open up a terminal, copy the following command: `/usr/bin/ruby -e "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/master/install)"` 
press enter, wait for a moment until the installation is completed. You now have installed homebrew successfully.

##### Install Minikube and Kubernetes
Install the kubectl command line tool by following these instructions:
1. [Install kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/#install-kubectl-on-macos) 
2. [Install Minikube](https://kubernetes.io/docs/tasks/tools/install-minikube/). For the hypervisor we recommend installing [Virtualbox](https://www.virtualbox.org/wiki/Downloads).
3. Verify that Minikube was installed by running`minikube start` from the command line.

##### Install Helm with Homebrew
1. `brew install kubernetes-helm`
3. Verify in the terminal that Helm was installed by running `helm help` and
make sure Kubernetes is running by doing `minikube start`and then run the command `helm init`.

##### Install the Kubernetes Service Catalog 
Kubernetes doesn’t include support for service brokers by default. For this we need to install an extension called the Service Catalog.

1. Clone the `https://github.com/10gen/atlas-service-broker`.
2. Run the script at `scripts/install-service-catalog.sh` to install the service catalog. If you get an error message stating _"could not find a ready tiller pod"_ wait a few minutes and try again.

##### Install Docker
1. Download Docker for Mac from [here](https://download.docker.com/mac/stable/Docker.dmg)
2. Start Docker for Mac, on the top right corner you should see the docker symbol signifying that it is running.


### Setting up the Service Broker

##### Setting up Atlas
The service broker needs a set of API keys to be able to communicate with Atlas. We recommend using cloud-qa.mongodb.com for testing.

1. Sign in to cloud-qa.mongodb.com. You should be redirected to the general dashboard page. On your left side you should see a sidebar with many options to choose from. Click on the button directly underneath Context, go to your organization and click on it. Now click on Access Management (or access) -> API Keys and create a new API key with full permissions, by clicking on Manage which is found on the right side. Make note of the public and private key after creation.
2. Create a new project by clicking on Projects, which is found on the left sidebar. Create a new project and take note of the project/group ID. The project ID can be found in Settings.

##### Install the Service Broker
1. Run the following command and replace GROUPID with your project ID from the previous step. Also replace PUBLICKEY and PRIVATEKEY with the API public and private key:
`kubectl create secret generic atlas-api --from-literal=group-id="GROUPID" --from-literal=public-key="PUBLICKEY" --from-literal=private-key="PRIVATEKEY" --from-literal=base-url="https://cloud-qa.mongodb.com/api/atlas/v1.0" --namespace atlas`
2. Run `scripts/kubernetes-deploy.sh atlas`

##### Deploying a cluster
Now the setup should be all done and Kubernetes is ready to provision clusters.
1. A new cluster can be deployed by running `kubectl create -filename scripts/kubernetes/instance.yaml -n atlas`.
2. The progress can be checked inside of cloud-qa.
3. Once the cluster has been provisioned a binding can be created using `kubectl create -f scripts/kubernetes/binding.yaml -n atlas`.
4. The provisioned cluster and binding can be removed again using `kubectl delete serviceinstance atlas-cluster-instance`.

More on Kubernetes commands can be found [here](https://kubernetes.io/docs/reference/generated/kubectl/kubectl-commands).
