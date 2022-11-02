# test-sdk

// TODO(user): A simple overview of the project and its purpose.

## Description

// TODO(user): An in-depth paragraph providing more details about the project and its use.

## Getting Started

You’ll need a Kubernetes and optionally a kcp cluster to run against. You can use [KIND](https://sigs.k8s.io/kind) to get a local cluster for testing, or run against a remote cluster.

**Note:** Your controller will automatically use the current context in your kubeconfig file (i.e. whatever cluster `kubectl cluster-info` shows).

### Running on Kubernetes or kcp

1. Build and push your image to the location specified by `REGISTRY` and `IMG`:
	
```sh
make docker-build docker-push REGISTRY=<some-registry> IMG=test-sdk:tag
```
	
2. Deploy the controller to the cluster with the image specified by `REGISTRY` and `IMG`:

```sh
make deploy REGISTRY=<some-registry> IMG=test-sdk:tag
```

### Uninstall resources

To delete the resources from the cluster:

```sh
make uninstall
```

### Undeploy controller

Undeploy the controller from the cluster:

```sh
make undeploy
```

## Contributing

// TODO(user): Add detailed information on how you would like others to contribute to this project.

### How it works

This project aims to follow the Kubernetes [Operator pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/)

It uses [Controllers](https://kubernetes.io/docs/concepts/architecture/controller/) 
which provides a reconcile function responsible for synchronizing resources untile the desired state is reached. 

### Test It Out

1. Install the required resources into the cluster:

```sh
make install
```

2. Run your controller (this will run in the foreground, so switch to a new terminal if you want to leave it running):

```sh
make run
```

**NOTE:** You can also run this in one step by running: `make install run`

### Modifying the API definitions

If you are editing the API definitions, regenerate the manifests using:

```sh
make manifests apiresourceschemas
```

**NOTE:** Run `make --help` for more information on all potential `make` targets

More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)

## License


Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

