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

const APP_TUNNEL_CLOUD = "tunnel-cloud.yaml"

const TunnelCloudYaml = `
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
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: tunnel-cloud
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: tunnel-cloud
subjects:
  - kind: ServiceAccount
    name: tunnel-cloud
    namespace: {{.Namespace}}
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: tunnel-cloud
  namespace: {{.Namespace}}
---
apiVersion: v1
data:
  mode.toml: |
    [mode]
    	[mode.cloud]
        	[mode.cloud.stream]
            	[mode.cloud.stream.server]
                	grpc_port = 9000
                	log_port = 51010
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
  namespace: {{.Namespace}}
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: tunnel-cloud-token
  namespace: {{.Namespace}}
data:
  token: |
    default:{{.TunnelCloudEdgeToken}}
---
apiVersion: v1
data:
  cloud.crt: '{{.TunnelPersistentConnectionServerCrt}}'
  cloud.key: '{{.TunnelPersistentConnectionServerKey}}'
  egress.crt: '{{.TunnelAnpServerCet}}'
  egress.key: '{{.TunnelAnpServerKey}}'
kind: Secret
metadata:
  name: tunnel-cloud-cert
  namespace: {{.Namespace}}
type: Opaque
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: tunnel-nodes
  namespace: {{.Namespace}}
data:
  hosts: ""
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: tunnel-cache
  namespace: {{.Namespace}}
data:
  edge_nodes: ""
  cloud_nodes: ""
  services: ""
  user_services: ""
---
apiVersion: v1
kind: Service
metadata:
  name: tunnel-cloud
  namespace: {{.Namespace}}
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
  type: NodePort
---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: tunnel-cloud
  name: tunnel-cloud
  namespace: {{.Namespace}}
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
          image: {{.TunnelImage}}
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
`
