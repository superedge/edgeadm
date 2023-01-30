/*
Copyright 2020 The SuperEdge Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package manifests

const APP_Edge_Coredns = "edge-coredns.yaml"

const EdgeCorednsYaml = `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: edge-coredns
  namespace: {{.Namespace}}
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  labels:
    kubernetes.io/bootstrapping: rbac-defaults
  name: system:edge-coredns
rules:
  - apiGroups:
    - ""
    resources:
    - endpoints
    - services
    - pods
    - namespaces
    verbs:
    - list
    - watch
  - apiGroups:
    - discovery.k8s.io
    resources:
    - endpointslices
    verbs:
    - list
    - watch
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  annotations:
    rbac.authorization.kubernetes.io/autoupdate: "true"
  labels:
    kubernetes.io/bootstrapping: rbac-defaults
  name: system:edge-coredns
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: system:edge-coredns
subjects:
- kind: ServiceAccount
  name: edge-coredns
  namespace: {{.Namespace}}
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: edge-coredns
  namespace: {{.Namespace}}
data:
  Corefile: |
    .:53 {
        errors
        bind {{.EdgeVirtualAddr}}
        health {
          lameduck 5s
        }
        hosts /etc/edge/hosts {
            reload 300ms
            fallthrough
        }
        ready localhost:8191
        kubernetes cluster.local in-addr.arpa ip6.arpa {
           pods insecure
           fallthrough in-addr.arpa ip6.arpa
           ttl 30
        }
        prometheus :9153
        forward . /etc/resolv.conf
        cache 30
        loop
        reload 2s
        loadbalance
    }
---
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: edge-coredns
  namespace: {{.Namespace}}
spec:
  selector:
    matchLabels:
      k8s-app: edge-coredns
  template:
    metadata:
      labels:
        k8s-app: edge-coredns
    spec:
      containers:
      - args: [ "-conf", "/etc/coredns/Corefile" ]
        image: {{.CoreDnsImage}}
        imagePullPolicy: IfNotPresent
        name: coredns
        ports:
        - containerPort: 53
          hostPort: 53
          name: dns
          protocol: UDP
        - containerPort: 53
          hostPort: 53
          name: dns-tcp
          protocol: TCP
        - containerPort: 9153
          hostPort: 9153
          name: metrics
          protocol: TCP
        resources:
          limits:
            cpu: 50m
            memory: 100Mi
          requests:
            cpu: 10m
            memory: 20Mi
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
            scheme: HTTP
        readinessProbe:
          httpGet:
            host: 127.0.0.1
            path: /ready
            port: 8191
            scheme: HTTP
        securityContext:
          allowPrivilegeEscalation: false
          capabilities:
            add:
            - NET_BIND_SERVICE
            drop:
            - all
          readOnlyRootFilesystem: true
        volumeMounts:
        - name: config-volume
          mountPath: /etc/coredns
          readOnly: true
      dnsPolicy: Default
      hostNetwork: true
      nodeSelector:
        superedge.io/node-edge: enable
      priorityClassName: system-cluster-critical
      restartPolicy: Always
      serviceAccount: edge-coredns
      serviceAccountName: edge-coredns
      tolerations:
      - key: CriticalAddonsOnly
        operator: Exists
      volumes:
      - name: config-volume
        configMap:
          name: edge-coredns
          items:
          - key: Corefile
            path: Corefile
`
