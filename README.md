# i-operator
// TODO(user): Add simple overview of use/purpose

## Description
// TODO(user): An in-depth paragraph about your project and overview of use

## Getting Started

### Prerequisites
- go version v1.23.0
- docker version 26.14.
- kubectl version v1.28.0.
- Access to a Kubernetes v1.28.0+ cluster.

### 环境安装

go 安装

```
wget https://go.dev/dl/go1.23.0.linux-amd64.tar.gz
tar -C /usr/local -xzf go1.23.0.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc 
go version
```

kubebuilder 安装

```
# download kubebuilder and install locally.
curl -L -o kubebuilder "https://go.kubebuilder.io/dl/latest/$(go env GOOS)/$(go env GOARCH)"
chmod +x kubebuilder && sudo mv kubebuilder /usr/local/bin/
```

项目初始化

```bash
mkdir i-operator
cd i-operator
# 创建一个名为 Application 的 CRD 对象
kubebuilder init --domain crd.genlang.cn --repo github.com/qinlanggen001/i-operator
```

创建API对象

```bash
kubebuilder create api --group core --version v1 --kind Application --namespaced=true
# 新增了 /api/v1 目录
# 新增 /bin 目录
# config 目录下新增 /config/crd 和 /config/samples
# 新增 /internal/controllers 目录
```

创建webhook

```bash
chmod +x bin/controller-gen-v0.17.2
kubebuilder create webhook --group core --version v1 --kind Application --defaulting --programmatic-validation
#Config 目录下增加了 Webhook 相关配置
#internal/webhook 目录下增加了 Webhook 默认实现
```

部署crd

```bash
# 执行 make manifests 命令，会根据我们定义的 CRD 生成对应的 yaml 文件，以及其他部署相关的 yaml 文件
make manifests
# CRD 部署到集群
make install
# 本地启动 Controller
make run
# 当有报错{"error": "open /tmp/k8s-webhook-server/serving-certs/tls.crt: no such file or directory"}，生成证书
openssl req -x509 -newkey rsa:4096 -keyout /tmp/k8s-webhook-server/serving-certs/tls.key -out /tmp/k8s-webhook-server/serving-certs/tls.crt -days 365 -nodes -subj "/CN=localhost"
# 重新执行make run
# make run
/mnt/go/i-operator/bin/controller-gen rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases
/mnt/go/i-operator/bin/controller-gen object:headerFile="hack/boilerplate.go.txt" paths="./..."
go fmt ./...
go vet ./...
go run ./cmd/main.go
2025-04-08T10:46:21+08:00	INFO	controller-runtime.builder	Registering a mutating webhook	{"GVK": "core.crd.genlang.cn/v1, Kind=Application", "path": "/mutate-core-crd-genlang-cn-v1-application"}
2025-04-08T10:46:21+08:00	INFO	controller-runtime.webhook	Registering webhook	{"path": "/mutate-core-crd-genlang-cn-v1-application"}
2025-04-08T10:46:21+08:00	INFO	controller-runtime.builder	Registering a validating webhook	{"GVK": "core.crd.genlang.cn/v1, Kind=Application", "path": "/validate-core-crd-genlang-cn-v1-application"}
2025-04-08T10:46:21+08:00	INFO	controller-runtime.webhook	Registering webhook	{"path": "/validate-core-crd-genlang-cn-v1-application"}
2025-04-08T10:46:21+08:00	INFO	setup	starting manager
2025-04-08T10:46:21+08:00	INFO	starting server	{"name": "health probe", "addr": "[::]:8081"}
2025-04-08T10:46:21+08:00	INFO	controller-runtime.webhook	Starting webhook server
2025-04-08T10:46:21+08:00	INFO	setup	disabling http/2
2025-04-08T10:46:21+08:00	INFO	Starting EventSource	{"controller": "application", "controllerGroup": "core.crd.genlang.cn", "controllerKind": "Application", "source": "kind source: *v1.Application"}
2025-04-08T10:46:21+08:00	INFO	controller-runtime.certwatcher	Updated current TLS certificate
2025-04-08T10:46:21+08:00	INFO	controller-runtime.webhook	Serving webhook server	{"host": "", "port": 9443}
2025-04-08T10:46:21+08:00	INFO	controller-runtime.certwatcher	Starting certificate poll+watcher	{"interval": "10s"}
2025-04-08T10:46:22+08:00	INFO	Starting Controller	{"controller": "application", "controllerGroup": "core.crd.genlang.cn", "controllerKind": "Application"}
2025-04-08T10:46:22+08:00	INFO	Starting workers	{"controller": "application", "controllerGroup": "core.crd.genlang.cn", "controllerKind": "Application", "worker count": 1}
2025/04/08 10:48:02 http: TLS handshake error from 20.163.14.5:42678: tls: first record does not look like a TLS handshake
```

部署完整的operator 到集群

```
make deploy
```



### To Deploy on the cluster

**Build and push your image to the location specified by `IMG`:**

```sh
make docker-build docker-push IMG=<some-registry>/i-operator:tag
```

**NOTE:** This image ought to be published in the personal registry you specified.
And it is required to have access to pull the image from the working environment.
Make sure you have the proper permission to the registry if the above commands don’t work.

**Install the CRDs into the cluster:**

```sh
make install
```

**Deploy the Manager to the cluster with the image specified by `IMG`:**

```sh
make deploy IMG=<some-registry>/i-operator:tag
```

> **NOTE**: If you encounter RBAC errors, you may need to grant yourself cluster-admin
privileges or be logged in as admin.

**Create instances of your solution**
You can apply the samples (examples) from the config/sample:

```sh
kubectl apply -k config/samples/
```

>**NOTE**: Ensure that the samples has default values to test it out.

### To Uninstall
**Delete the instances (CRs) from the cluster:**

```sh
kubectl delete -k config/samples/
```

**Delete the APIs(CRDs) from the cluster:**

```sh
make uninstall
```

**UnDeploy the controller from the cluster:**

```sh
make undeploy
```

## Project Distribution

Following the options to release and provide this solution to the users.

### By providing a bundle with all YAML files

1. Build the installer for the image built and published in the registry:

```sh
make build-installer IMG=<some-registry>/i-operator:tag
```

**NOTE:** The makefile target mentioned above generates an 'install.yaml'
file in the dist directory. This file contains all the resources built
with Kustomize, which are necessary to install this project without its
dependencies.

2. Using the installer

Users can just run 'kubectl apply -f <URL for YAML BUNDLE>' to install
the project, i.e.:

```sh
kubectl apply -f https://raw.githubusercontent.com/<org>/i-operator/<tag or branch>/dist/install.yaml
```

### By providing a Helm Chart

1. Build the chart using the optional helm plugin

```sh
kubebuilder edit --plugins=helm/v1-alpha
```

2. See that a chart was generated under 'dist/chart', and users
can obtain this solution from there.

**NOTE:** If you change the project, you need to update the Helm Chart
using the same command above to sync the latest changes. Furthermore,
if you create webhooks, you need to use the above command with
the '--force' flag and manually ensure that any custom configuration
previously added to 'dist/chart/values.yaml' or 'dist/chart/manager/manager.yaml'
is manually re-applied afterwards.

## Contributing
// TODO(user): Add detailed information on how you would like others to contribute to this project

**NOTE:** Run `make help` for more information on all potential `make` targets

More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)

## License

Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

