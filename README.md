## Edgeadm使用教程

### 源码编译
您可以选择使用SuperEdge Release的版本，也可以根据需要使用源代码编译出符合您需求的版本

#### 1. 选择Release版本
- [版本列表](https://github.com/superedge/superedge/releases)

#### 2. 我要编译源代码

- deal with dependency: `go mod tidy`
- build: `make build`
- clean: `make clean`

> edgeadm 输出的二进制文件在`output`文件夹下

#### 3. 版本适配列表
由于 edgeadm 依赖的 kubeadm 和 kubernetes 版本有强依赖关系，请检查您需要的 edgeadm 版本：现阶段 main 主分支支持部署 Kubernetes 1.22 版本；如果需要部署更低版本的Kubernetes ，请 checkout 到对应的分支进行编译

| Branch         | Kubernetes 1.18.2 | Kubernetes 1.20.6 | Kubernetes 1.22.6 |
| -------------- | ----------------- | ----------------- | ----------------- |
| `release-1.20` | ✓(v0.7.0)         | ✓(v0.7.0)         | -                 |
| `HEAD`         | -                 | -                 | ✓(v0.8.0)         |


### 开始部署

#### 1. 两条指令从零搭建一个边缘集群
-   下载安装包
> main 主分支现在仅支持部署 Kubernetes 1.22.6版本，注意修改"arch=amd64"参数，目前支持[amd64, arm64], 下载自己机器对应的体系结构，其他参数不变
```
arch=amd64 version=v0.8.0 kubernetesVersion=1.22.6 && rm -rf edgeadm-linux-* && wget https://superedge-1253687700.cos.ap-guangzhou.myqcloud.com/$version/$arch/edgeadm-linux-$arch-$version-k8s-$kubernetesVersion.tgz && tar -xzvf edgeadm-linux-* && cd edgeadm-linux-$arch-$version-k8s-$kubernetesVersion && ./edgeadm
```

-   安装边缘 Kubernetes master 节点
```shell
./edgeadm init --kubernetes-version=1.22.6 --image-repository superedge.tencentcloudcr.com/superedge --service-cidr=10.244.0.0/16 --pod-network-cidr=10.233.0.0/16 --install-pkg-path ./kube-linux-*.tar.gz --apiserver-cert-extra-sans=<Master节点 内网/外网IP> --apiserver-advertise-address=<Master节点内网/外网 IP> --enable-edge=true
```

-   Join 边缘节点
```shell
./edgeadm join <Master节点外网IP/Master节点内网IP/域名>:Port --token xxxx --discovery-token-ca-cert-hash sha256:xxxxxxxxxx --install-pkg-path <edgeadm kube-*静态安装包地址> --enable-edge=true 
```

详情见：[从零搭建边缘集群](./docs/installation/install_edge_kubernetes_CN.md)

#### 2. 一键将已有集群转换成边缘集群

- 将普通集群转换成边缘集群: `edgeadm change --kubeconfig admin.kubeconfig`

- 将边缘集群回退成普通集群: `edgeadm revert --kubeconfig admin.kubeconfig`

- [edgeadm 一键转换](./docs/installation/install_via_edgeadm_CN.md)

#### 3. 以Addon方式部署SuperEdge
- [Addon方式部署](./docs/installation/addon_superedge_CN.md)

#### 4. 我是高手，想一个个组件手工部署

- [手工部署](./docs/installation/install_manually_CN.md)
