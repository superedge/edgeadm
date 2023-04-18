## Edgeadm使用教程

### 源码编译

您可以选择使用SuperEdge Release的版本，也可以根据需要使用源代码编译出符合您需求的版本

#### 1. 选择Release版本

- [最新版本](https://github.com/superedge/edgeadm/releases)
- [历史版本](https://github.com/superedge/superedge/releases)

#### 2. 我要编译源代码

- deal with dependency: `go mod tidy`
- build: `make build`
- clean: `make clean`

> edgeadm 输出的二进制文件在`output`文件夹下

#### 3. 版本适配列表

由于 edgeadm 依赖的 kubeadm 和 kubernetes 版本有强依赖关系，请检查您需要的 edgeadm 版本：现阶段 main 主分支支持部署 Kubernetes 1.22 版本；如果需要部署更低版本的Kubernetes ，请 checkout 到对应的分支进行编译

| Branch         | Kubernetes 1.18.2 | Kubernetes 1.20.6 | Kubernetes 1.22.6 |
| -------------- | ----------------- | ----------------- | ----------------- |
| `release-1.18` | ✓                 | -                 | -                 |
| `release-1.20` | -                 | ✓                 | -                 |
| `HEAD`         | -                 | -                 | ✓                 |

> 注意：最新的v0.9.0版本仅支持 Kubernetes 1.22.6 版本

### 开始部署

#### 1. 两条指令从零搭建一个边缘集群

- 下载安装包
  > edgeadm 最近两个版本[v0.9.0,v0.8.2]支持的体系结构 arch[amd64, arm64]以及kubernetes 版本[1.22.6, 1.20.6]组合如下，请大家按需下载：
  > - CPU arch [amd64, arm64], kubernetes version [1.22.6], version: v0.9.0
  > - CPU arch [amd64, arm64], kubernetes version [1.22.6, 1.20.6], version: v0.8.2
  > 注意修改 `arch/version/kubernetesVersion` 变量参数来下载 tgz 包：  
  ```
  arch=amd64 version=v0.9.0 kubernetesVersion=1.22.6 && rm -rf edgeadm-linux-* && wget https://superedge-1253687700.cos.ap-guangzhou.myqcloud.com/$version/$arch/edgeadm-linux-$arch-$version-k8s-$kubernetesVersion.tgz && tar -xzvf edgeadm-linux-* && cd edgeadm-linux-$arch-$version-k8s-$kubernetesVersion && ./edgeadm
  ```
  此静态安装包也可以从 [Github Release页面](https://github.com/superedge/edgeadm/releases) 下载

- 安装边缘 Kubernetes master 节点
  将下载的压缩包解压后，进入目录，执行下面的命令：
  
  ```shell
  ./edgeadm init --kubernetes-version=1.22.6 --image-repository superedge.tencentcloudcr.com/superedge --service-cidr=10.244.0.0/16 --pod-network-cidr=10.233.0.0/16 --install-pkg-path ./kube-linux-*.tar.gz --apiserver-cert-extra-sans=<Master节点外网 IP/域名等> --apiserver-advertise-address=<Master节点内网 IP> --enable-edge=true --edge-version=0.9.0
  ```

> --apiserver-cert-extra-sans=<Master节点外网 IP/域名等>：这里的外网 IP 指的是边缘节点需要接入的云端控制面的公网 IP以及外网域名，apiserver 会签发相应的证书供边缘节点访问
> 
> --apiserver-advertise-address=<Master节点内网 IP>：这里的内网 IP 指的是 edgeadm 用于初始化 etcd 和 apiserver 需要绑定的节点内部 IP
> 
> --edge-version=0.9.0：如果需要使用最新的 [Kins](https://github.com/superedge/superedge/blob/main/docs/components/kins_CN.md) 能力, 这里需要指定最新`v0.9.0`的版本（仅支持 Kubernetes 1.22.6）；如果不需要 Kins 能力，同时又希望能够使用类似 1.20 的低 K8s 版本，可以使用 `v0.8.2`版本，支持最新的云边隧道能力，支持云端 master、worker 和边缘节点三种类型节点的 7 层协议互通，适配更加完善。

- Join 边缘节点

```shell
./edgeadm join <Master节点外网IP/域名>:Port --token xxxx --discovery-token-ca-cert-hash sha256:xxxxxxxxxx --install-pkg-path <edgeadm kube-*静态安装包地址> --enable-edge=true 
```

> --enable-edge=true: true 代表是边缘节点，会部署 lite-apiserver 等边缘组件；false 代表是云上 worker 节点，会按照标准 kubeadm 方式部署，不会部署边缘组件

详情见：[从零搭建边缘集群](./docs/installation/install_edge_kubernetes_CN.md)

#### 2. 一键将已有集群转换成边缘集群

- 将普通集群转换成边缘集群: `edgeadm change --kubeconfig admin.kubeconfig`

- 将边缘集群回退成普通集群: `edgeadm revert --kubeconfig admin.kubeconfig`

- [edgeadm 一键转换](./docs/installation/install_via_edgeadm_CN.md)

#### 3. 以Addon方式部署SuperEdge

- [Addon方式部署](./docs/installation/addon_superedge_CN.md)

#### 4. 我是高手，想一个个组件手工部署

- [手工部署](./docs/installation/install_manually_CN.md)
