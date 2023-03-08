package manifests

const APP_SITE_MANAGER = "site-manager.yaml"
const SiteManagerYaml = `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: site-manager-service-account
  namespace: {{.Namespace}}
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: site-manager-cluster-role
rules:
  - apiGroups:
      - apiextensions.k8s.io
    resources:
      - customresourcedefinitions
    verbs:
      - "*"
  - apiGroups:
      - ""
    resources:
      - pods
      - nodes
      - services
      - secrets
      - namespaces
      - events
      - configmaps
      - persistentvolumes
    verbs:
      - "*"
  - apiGroups:
      - site.superedge.io
    resources:
      - "*"
    verbs:
      - "*"
  - apiGroups:
      - apps
    resources:
      - daemonsets
      - statefulsets
    verbs:
      - "*"
  - apiGroups:
      - storage.k8s.io
    resources:
      - storageclasses
    verbs:
      - "*"
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: site-manager-cluster-role-binding
  namespace: {{.Namespace}}
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: site-manager-cluster-role
subjects:
  - kind: ServiceAccount
    name: site-manager-service-account
    namespace: {{.Namespace}}
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: site-manager
  namespace: {{.Namespace}}
spec:
  replicas: 1
  selector:
    matchLabels:
      app: site-manager
  strategy:
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 1
    type: RollingUpdate
  template:
    metadata:
      labels:
        app: site-manager
    spec:
      containers:
        - name: site-manager
          image: {{.SiteManagerImage}}
          imagePullPolicy: Always
          command:
            - /usr/local/bin/site-manager
            - v=4
          resources:
            limits:
              cpu: 50m
              memory: 100Mi
            requests:
              cpu: 10m
              memory: 20Mi
          terminationMessagePath: /dev/termination-log
          terminationMessagePolicy: File
          securityContext:
            privileged: true
      dnsPolicy: ClusterFirst
      restartPolicy: Always
      serviceAccountName: site-manager-service-account
      schedulerName: default-scheduler
      terminationGracePeriodSeconds: 30
      tolerations:
        - key: "node-role.kubernetes.io/master"
          operator: "Exists"
          effect: "NoSchedule"
`
