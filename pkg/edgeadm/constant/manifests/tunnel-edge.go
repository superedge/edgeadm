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

const APP_TUNNEL_EDGE = "tunnel-edge.yaml"

const TunnelEdgeYaml = `
apiVersion: v1
kind: ConfigMap
metadata:
  name: tunnel-edge-conf
  namespace: {{.Namespace}}
data:
  mode.toml: |
    [mode]
        [mode.edge]
            [mode.edge.stream]
                [mode.edge.stream.client]
                    token = "{{.TunnelCloudEdgeToken}}"
                    dns = "127.0.0.1"
                    server_name = "{{.MasterIP}}:{{.TunnelPersistentConnectionPort}}"
                    log_port = 51010
            [mode.edge.http_proxy]
                ip = "0.0.0.0"
                port = "51009"
                
---
apiVersion: v1
data:
  ca.crt: '{{.KubernetesCaCert}}'
kind: Secret
metadata:
  name: tunnel-edge-cert
  namespace: {{.Namespace}}
type: Opaque
---
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: tunnel-edge
  namespace: {{.Namespace}}
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
          image: {{.TunnelImage}}
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
          name: certs
        - configMap:
            name: tunnel-edge-conf
          name: conf
`
