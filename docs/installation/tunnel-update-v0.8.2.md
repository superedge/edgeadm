# Tunnel 新版本升级流程(v0.8.2)

现在最新版本 v0.8.2 版本对 Tunnel 进行了较大服务重构，云边支持能力更加全面。本文会详细描述如何从旧版本（v0.8.0之前的版本）升级到 v0.8.2 最新版本的操作步骤

## 1. 云端 tunnel-cloud升级

### 1.1 原始信息分析

tunnel-cloud 主要在云端节点（Master 和 Worker 节点）部署了 `deployment/tunnel-cloud` ，同时主要依赖下面几个 Configmap

- `tunnel-cloud-token`：这里会记录之前版本 token 信息，用于 tunnel-edge 连接云端，这里的信息最好做一下备份；如果不需要修改 token，这个 cm 的信息保持不变

- `tunnel-node`：这里会记录边缘 edge 节点接入时的链接信息，包括边缘节点的 name 以及对应的 tunnel-cloud 的 pod ip；这个信息会在 tunnel-cloud 的 pod 重启后自动 Update，也可以不用单独处理，保持不变即可

- `tunnel-cloud-conf`：这个cm 记录了 tunnel-cloud 的配置文件信息，这里会更新，请参考下面的章节

- `tunnel-coredns`：这个 cm  之前用于进行云——边链路选择，现在重构后不再使用，可以删除

### 1.2 更新步骤

- 删除不需要的 Configmap：`edge-system/tunnel-coredns`

- 创建新的 Configmap：`edge-system/tunnel-cache`，这个 cm 用于同步云边的节点信息和 svc信息，重要！参考下面的 yaml 文件创建：
  
  ```yaml
  ---
  apiVersion: v1
  kind: ConfigMap
  metadata:
    name: tunnel-cache
    namespace: edge-system
  data:
    edge_nodes: ""
    cloud_nodes: ""
    services: ""
    user_services: ""
  ```

- 更新 Confgmap：`edge-system/tunnel-cloud-conf`
  
  ```yaml
  ---
  apiVersion: v1
  data:
    mode.toml: |
      [mode]
          [mode.cloud]
              [mode.cloud.stream]
                  [mode.cloud.stream.server]
                          grpc_port = 9000
                          log_port = 7000
                          metrics_port = 6000
                  [mode.cloud.stream.register]
                          service = "tunnel-cloud"
              [mode.cloud.egress]
                  port = 8000
              [mode.cloud.http_proxy]
                  port = 8080
              [mode.cloud.ssh]
                  port = 22
  kind: ConfigMap
  metadata:
    name: tunnel-cloud-conf
    namespace: edge-system
  ```

- 更新 Deployment：`edge-system/tunnel-cloud` 以及 Service：`edge-system/tunnel-cloud`，参考下面的 yaml 文件：
  
  ```yaml
  ---
  apiVersion: rbac.authorization.k8s.io/v1
  kind: ClusterRole
  metadata:
   name: tunnel-cloud
  rules:
   - apiGroups: [ "" ]
     resources: [ "configmaps" ]
     verbs: [ "get", "update" ]
   - apiGroups: [ "" ]
     resources: [ "services","pods","nodes" ]
     verbs: [ "get","list","watch" ]
   - apiGroups: [ "" ]
     resources: [ "endpoints" ]
     verbs: [ "get","list","watch","create","update" ]
  ---
  apiVersion: apps/v1
  kind: Deployment
  metadata:
   labels:
     app: tunnel-cloud
   name: tunnel-cloud
   namespace: edge-system
  spec:
   selector:
     matchLabels:
       app: tunnel-cloud
   template:
     metadata:
       labels:
         app: tunnel-cloud
     spec:
       serviceAccount: tunnel-cloud
       serviceAccountName: tunnel-cloud
       containers:
         - name: tunnel-cloud
           image: superedge.tencentcloudcr.com/superedge/tunnel:v0.8.2
           imagePullPolicy: IfNotPresent
           livenessProbe:
             httpGet:
               path: /cloud/healthz
               port: 51010
             initialDelaySeconds: 10
             periodSeconds: 60
             timeoutSeconds: 3
             successThreshold: 1
             failureThreshold: 1
           command:
             - /usr/local/bin/tunnel
           args:
             - --m=cloud
             - --c=/etc/tunnel/conf/mode.toml
             - --log-dir=/var/log/tunnel
             - --alsologtostderr
           env:
             - name: POD_IP
               valueFrom:
                 fieldRef:
                   apiVersion: v1
                   fieldPath: status.podIP
             - name: POD_NAMESPACE
               valueFrom:
                 fieldRef:
                   apiVersion: v1
                   fieldPath: metadata.namespace
             - name: POD_NAME
               valueFrom:
                 fieldRef:
                   apiVersion: v1
                   fieldPath: metadata.name
             - name: USER_NAMESPACE
               value: edge-system
           volumeMounts:
             - name: token
               mountPath: /etc/tunnel/token
             - name: certs
               mountPath: /etc/tunnel/certs
             - name: hosts
               mountPath: /etc/tunnel/nodes
             - name: conf
               mountPath: /etc/tunnel/conf
             - name: cache
               mountPath: /etc/tunnel/cache
           ports:
             - containerPort: 9000
               name: tunnel
               protocol: TCP
           resources:
             limits:
               cpu: 50m
               memory: 100Mi
             requests:
               cpu: 10m
               memory: 20Mi
       volumes:
         - name: token
           configMap:
             name: tunnel-cloud-token
         - name: certs
           secret:
             items:
             - key: tunnel-cloud-server.crt
               path: cloud.crt
             - key: tunnel-cloud-server.key
               path: cloud.key
             - key: tunnel-anp-server.crt
               path: egress.crt
             - key: tunnel-anp-server.key
               path: egress.key
             secretName: tunnel-cloud-cert
         - name: hosts
           configMap:
             name: tunnel-nodes
         - name: cache
           configMap:
             name: tunnel-cache
         - name: conf
           configMap:
             name: tunnel-cloud-conf
       nodeSelector:
         node-role.kubernetes.io/master: ""
       tolerations:
         - key: "node-role.kubernetes.io/master"
           operator: "Exists"
           effect: "NoSchedule"
  ---
  apiVersion: v1
  kind: Service
  metadata:
    name: tunnel-cloud
    namespace: edge-system
  spec:
    ports:
      - name: grpc
        port: 9000
        protocol: TCP
        targetPort: 9000
      - name: ssh
        port: 22
        protocol: TCP
        targetPort: 22
      - name: tunnel-metrics
        port: 6000
        protocol: TCP
        targetPort: 6000
      - name: egress
        port: 8000
        protocol: TCP
        targetPort: 8000
      - name: http-proxy
        port: 8080
        protocol: TCP
        targetPort: 8080
    selector:
      app: tunnel-cloud
    sessionAffinity: None
    type: NodePort
  ```

### 1.3 给节点打标签

新版本的 Tunnel 支持同时访问云上 worker node 以及边缘 edge node，因此需要通过 label 区分不同地域的 node，如下：

- 云上节点，包括 Master node 和 Worker node，都保证需要打上如下标签（edgeadm 部署的时候会自动添加；如果没有自动添加，请手动 label）    
  
  ```shell
  kubectl  label nodes masterNodeName   superedge.io/node-cloud=enable
  ```

- 边缘节点，通过 edgeadm --enable-edge=true 添加进来的节点，认为是边缘节点，保证需要打上如下标签（edgeadm join 添加的节点会自动标签；如果没有请手动添加）
  
  ```shell
  kubectl  label nodes masterNodeName   superedge.io/node-edge=enable
  ```

## 2. 边缘端 tunnel-edge 升级

### 2.1 原始信息备份

边缘侧通过通过一个 Daemonset：`edge-system/tunnel-edge`在边缘节点上拉起 tunnel-edge 来和云端 tunnel-cloud 创建连接。同时包括一个配置文件 Configmap：`edge-system/tunnel-edge-conf`，升级的用户需要从原始的这个 cm 中获取两个信息，如下：

```yaml
从原始 yaml 中获取 token 和 servername 两个信息保存
[mode]
    [mode.edge]
        [mode.edge.stream]
            [mode.edge.stream.client]
                token = "hFcu71fDEZZ8wY6jyJXWnk9h2RADrqgH"  #备份 token 信息
                cert = "/etc/superedge/tunnel/certs/cluster-ca.crt"
                dns = "tunnel.cloud.io"
                servername = "xxx.yyy.zzz.mmm:31641"   #备份 servername，云端 tunnel-cloud 的服务地址
                logport = 51010
            [mode.edge.https]
                cert= "/etc/superedge/tunnel/certs/apiserver-kubelet-client.crt"
                key=  "/etc/superedge/tunnel/certs/apiserver-kubelet-client.key"
        [mode.edge.httpproxy]
            proxyip = "0.0.0.0"
            proxyport = "51009"
```

### 2.2 更新步骤

- 更新 Configmap：`edge-system/tunnel-edge-conf`：
  
  ```yaml
  apiVersion: v1
  kind: ConfigMap
  metadata:
    name: tunnel-edge-conf
    namespace: edge-system
  data:
    mode.toml: |
      [mode]
          [mode.edge]
              [mode.edge.stream]
                  [mode.edge.stream.client]
                      token = "{token}"
                      dns = "127.0.0.1"
                      server_name = "{ip:port}"
                      log_port = 51010
              [mode.edge.http_proxy]
                  ip = "0.0.0.0"
                  port = "51009"
  ```
  
  使用之前保存的 tunnel-edge-conf 的**token**和**servername**进行替换，通过上述配置，tunnel-edge 会在边缘节点上使用 `mode.edge.http_proxy.port=51009`开启边缘侧代理服务，用户通过 `http_proxy=http://169.254.20.11:51009`即可通过代理访问云端 service 或者 pod

- 更新 Daemonset：`edge-system/tunnel-edge`：
  
  ```yaml
  apiVersion: apps/v1
  kind: DaemonSet
  metadata:
    name: tunnel-edge
    namespace: edge-system
  spec:
    selector:
      matchLabels:
        app: tunnel-edge
    template:
      metadata:
        labels:
          app: tunnel-edge
      spec:
        hostNetwork: true
        nodeSelector:
          superedge.io/node-edge: enable
        containers:
          - name: tunnel-edge
            image: superedge.tencentcloudcr.com/superedge/tunnel:v0.8.2
            imagePullPolicy: IfNotPresent
            livenessProbe:
              httpGet:
                path: /edge/healthz
                port: 51010
              initialDelaySeconds: 10
              periodSeconds: 180
              timeoutSeconds: 3
              successThreshold: 1
              failureThreshold: 3
            resources:
              limits:
                cpu: 20m
                memory: 40Mi
              requests:
                cpu: 10m
                memory: 10Mi
            command:
              - /usr/local/bin/tunnel
            env:
              - name: NODE_NAME
                valueFrom:
                  fieldRef:
                    apiVersion: v1
                    fieldPath: spec.nodeName
            args:
              - --m=edge
              - --c=/etc/tunnel/conf/mode.toml
              - --log-dir=/var/log/tunnel
              - --alsologtostderr
            volumeMounts:
              - name: certs
                mountPath: /etc/tunnel/certs
              - name: conf
                mountPath: /etc/tunnel/conf
        volumes:
          - secret:
              secretName: tunnel-edge-cert
              items:
              - key: cluster-ca.crt
                path: ca.crt
            name: certs
          - configMap:
              name: tunnel-edge-conf
            name: conf
  ```
