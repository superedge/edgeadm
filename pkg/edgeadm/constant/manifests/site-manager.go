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
      - persistentvolumeclaims
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
apiVersion: v1
data:
  webhook_server.crt: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUMrVENDQWVHZ0F3SUJBZ0lVTC9TTjJOVlBDWS9yZVJOSFBleGgvZzBSaWVZd0RRWUpLb1pJaHZjTkFRRUwKQlFBd0ZURVRNQkVHQTFVRUF4TUthM1ZpWlhKdVpYUmxjekFlRncweU16QXpNVFl3T1RJMk1UVmFGdzB5TkRBegpNVFV3T1RJMk1UVmFNQmN4RlRBVEJnTlZCQU1NREhOcGRHVXRiV0Z1WVdkbGNqQ0NBU0l3RFFZSktvWklodmNOCkFRRUJCUUFEZ2dFUEFEQ0NBUW9DZ2dFQkFNSCtvLy9GMzRsSmFvMUcyd1dGZ2IvZGRLb0lOaDlSVUtDWnZMWGkKaUtORlpmMnlJcFFKbFc2UUhYWHdNVFBQdkNpbkxHUm10YzRGaFlQN0E5YUREdEdjNDIvSUErN2Q5Yk9KekJRQQpqaW9kTmszK0lvbFg1NFVLT2VSUDNDZWZNTFFSSWJpTmJrTkd0RlNPcGpIZzQ3T0JIMFZXNjJreEdJZmdLK0llCjdDY0hkU2doQ1AyREM2cW92ZjRvTjJLdHRySlNzenVBaFpQWUhJQ0FJUWMzYU1ha3JyUjVLRjNqU1VtajhNOGQKWCtqajFnQzRZUnVDbEFhb2ZDVzJJczFkZFZUb0NhSGNlTzA4RE1razRtdkdLVWsyM2x3S3hwUlRKOVNNQTNKOQpaVFhjeit2aWtZaEI2N3Q0MmV6S2M3bDUvOXJURUFiR3E1YnM5YkhMUTZpZlU0MENBd0VBQWFNL01EMHdPd1lEClZSMFJCRFF3TW9JcWMybDBaUzF0WVc1aFoyVnlMbVZrWjJVdGMzbHpkR1Z0TG5OMll5NWpiSFZ6ZEdWeUxteHYKWTJGc2h3Ui9BQUFCTUEwR0NTcUdTSWIzRFFFQkN3VUFBNElCQVFBT0ZkL09iZU16am1ScEhFM3NhV01neDN4YQowQ0xGTUhiYlB5MVNrZzJwVzlQR2M0djNRS1NyRU0xeDJlNysxR1VXd3dxRUZENVNSN3k0UFl0QUxTejNGcHNuCmp0YTFyNkp5ZUpObE1DdFNKSDJwZnhOTUtRQXIybWZYK2oyRFg0aFY2MGhNeFNySUZnTXZ1Q1hMQ0pWbC9yQUYKNVovdExLQVFXWldVOUI3SEx5S0lremFBemZZV1M1V1hXcmFBNXg0UlFBQ0dITEQvdzBtWTBZczJTRlVhVE5zZwpWYmZqYkxKelBCK2VVOHlBb0xWQ2hEcjJBdkw1NlRsSW5xd2t0N3kwQTREN1h4MUM4Ry92MUUwS1pQSHoxVFg1CkRvUXVuWDdBR1NQc1FEY21kS0NKVlNDTVBMWnBvTHk4K1EydUU5bGpKdVNISlNpd2tyZFpMdHpKeTFyZAotLS0tLUVORCBDRVJUSUZJQ0FURS0tLS0tCg==
  webhook_server.key: LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlFb3dJQkFBS0NBUUVBd2Y2ai84WGZpVWxxalViYkJZV0J2OTEwcWdnMkgxRlFvSm04dGVLSW8wVmwvYklpCmxBbVZicEFkZGZBeE04KzhLS2NzWkdhMXpnV0ZnL3NEMW9NTzBaempiOGdEN3QzMXM0bk1GQUNPS2gwMlRmNGkKaVZmbmhRbzU1RS9jSjU4d3RCRWh1STF1UTBhMFZJNm1NZURqczRFZlJWYnJhVEVZaCtBcjRoN3NKd2QxS0NFSQovWU1McXFpOS9pZzNZcTIyc2xLek80Q0ZrOWdjZ0lBaEJ6ZG94cVN1dEhrb1hlTkpTYVB3engxZjZPUFdBTGhoCkc0S1VCcWg4SmJZaXpWMTFWT2dKb2R4NDdUd015U1RpYThZcFNUYmVYQXJHbEZNbjFJd0RjbjFsTmR6UDYrS1IKaUVIcnUzalo3TXB6dVhuLzJ0TVFCc2FybHV6MXNjdERxSjlUalFJREFRQUJBb0lCQUVocGs2L3luWWt5WlZxTQoyMDZKVWpCYktxUVpZcEo0R04rSjQxNEZURG1kdXY5aTBmTnVUR0F6M1R0YnlCSHQ5ZTg2ejZBK2twaHZpVElGCnJaNFIxNk00cTlEYTJWVDlkeXhvUUV3ckZPWDFkNExQWFBibFlCOVIvT0FUU3p0aStad05WUWY5aXU0RDQyWTQKeFNLVExvdWZwQnVPNFZxbm45K0FOd0UxeDdLZE9CN0t3S2ZqaUhFSm1Cc1NiMU9MS3E5ajd6RGZHRWhVczdYUwo2T05oeThXVWc0ckg5VVVpSGdZS1JROTJKdkFpcW5aNGpRcHB1Sm9VZGVlRE9iTzZUOGpTbHY3TjNOdC9aVTJSClJDQy9jZUxWWjJheUdnczdyNWNOZjRaVHJaV3hXSFlrWWhIWXZXaVNuODZiRnBxRWxNaFN2Q2xTZmNoTUxEcE8KYVBTWEdRRUNnWUVBOFlCdlJpRXY2N292UTR1S1VPSmk3eXlKSkd3eHd1Y1lLcHRFN1VXMWx6dnpnbjhZQmlsMgpoNkQ3NXNtb0lOR3NzY1Nrajg1amxPNy9BaU9FZnU3WU1Jc1daNllVSmRWbXB3Q09wdUNBTnN4NWZ0bm5IU05vCi9rOGJNa29mSmNmM0lTQ2F1M1ZJYTg4TUt0eHMwd3puWEMrbm81TVEwYkQ0eG5rbGNubW4rMEVDZ1lFQXphUVYKMmlZR3RSdmZJYmhtbWEzc2dwM2NQZXRHNjdQb2FXc3ZPUmNCOS9tWVNVRnpWNEEvcENMUnVDWE5heVQ0WTNESApKYkNNcXMrbk4yUFprNk9LZVl6bVFyLzNvUUtXWVhqeG9weU9HR05Hb0hENHg2anN6OGNyU0cvblBrRGJoUlk0CmRoUUlsbFg0MWEvaHE0cHA4RXJ6Yy9HMXdZSE9iQW1mYmswOWdVMENnWUJFNjRGV0F6eVl5azZZdVNibEJHWjEKbVVFZUt0NWNuL1RPbS9ja3U2TWlJTkxTcUJDa0dZc2hFN2t2Mk5icFhzMHBBbFJ3VWRjcmRyVkIxLzhFOW9hdAorOU9PQ3VCdkY2S3ZBRUsxcnhZSURYeVN6ZjdkMnZBb3UweW9vbXlYTEtVRFNEbkFTNjA2VHlGS3poTWtlK2MrCjhMNm51TjJ3Nmc5bEhNZXFEcnY0d1FLQmdERTF1ZkQ1TnBPeWRzUDMvNzE0N0djWlpiSC9rbm9uRkUvZDBYQWsKL0ZpZUJ2NUl4bFJESVhlaXlYTDZ3TnlKL1ZLMmswR0dyVExXL0ZuNThBQXZtNXlZeGlWbEVOb2I1MmF0N1kwUApUOFd3UkI5eXlXWG1HNzFoR1E5OWorWEsyWDFRb3ZSR3VRTlkwWEk1WTVTTVMrdXYwL0NFQUEydGhYcy9Ga0xzCkF2TGxBb0dCQU5ZeThPVjkzMmRSUDg1YThKeC9aOGNDYXNROEJIVVVTUitpM21zR3hCTlFXTWEvQ3BCRTFpK0wKZ3orOGdZbytMT0dCcmZNQXVKZS9xbWhGK3gwblZEbXVhNVBZaXFvR01QaFlQV2VLV05NQTdMTzhXRnN0UktTRwplMVNZL3dtdXVuUFVVYmd3S0R1LzRpL0ZBWFJpMzFsT2tJZGNNa2d2ODRrN0tUYVBaTHB3Ci0tLS0tRU5EIFJTQSBQUklWQVRFIEtFWS0tLS0tCg==
kind: Secret
metadata:
  name: site-manager-cert
  namespace: {{.Namespace}}
type: Opaque
---
apiVersion: v1
kind: Service
metadata:
  name: site-manager
  namespace: {{.Namespace}}
spec:
  clusterIP: None
  ports:
    - name: webhook
      port: 9000
      protocol: TCP
      targetPort: 9000
  selector:
    app: site-manager
  sessionAffinity: None
  type: ClusterIP
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
          env:
          - name: CONVERT_WEBHOOK_SERVER
            value: https://site-manager.edge-system.svc.cluster.local:9000/v1
          - name: CA_CRT
            value: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUMvakNDQWVhZ0F3SUJBZ0lCQURBTkJna3Foa2lHOXcwQkFRc0ZBREFWTVJNd0VRWURWUVFERXdwcmRXSmwKY201bGRHVnpNQjRYRFRJek1ETXhNekEzTURVMU9Gb1hEVE16TURNeE1EQTNNRFUxT0Zvd0ZURVRNQkVHQTFVRQpBeE1LYTNWaVpYSnVaWFJsY3pDQ0FTSXdEUVlKS29aSWh2Y05BUUVCQlFBRGdnRVBBRENDQVFvQ2dnRUJBTW94CjlaOERuLy9Yb09wZllIbjZjdk1mS21HbVNMYmtWQTUwT1dOOEFaazVybTRMRFhJeXR6NnlGd2lqbFdXZjRmS2YKRG1jZWgwVFNpbldVYkJhcFNzRUJUc1EvOHBWNk00MTFjQ2RWR1hLUG4zbWxZeVlXNWlIVlZ0cG9RMHh5RGFIcwpZbVNEeHhUakk1ZjlLQnUxOVNwM3JFd3JTdDNDVXJPQUthNnpZZDgvamFuR1lvbDhCK2sxL01YenRWd1pZazB4Cm90VjluN3FHcXZ2Z0NYVWp4RHFsU2JCVjROdmVPalR0N0x5NG5HSXc1UDNFSWxVRm9uL1c3N2NDR3BLVC9rTUkKMGZVOG03aEI2STNEK0dXK0ExUFRiQk9uaGNxQURVaHhZTkVxaFVzZzJVTmw3aXFES0xuQ0Jjd1d5Z1N3UWdhMwp5ZGRYOS90V0d0am0yNEJCMUU4Q0F3RUFBYU5aTUZjd0RnWURWUjBQQVFIL0JBUURBZ0trTUE4R0ExVWRFd0VCCi93UUZNQU1CQWY4d0hRWURWUjBPQkJZRUZCUytkc2ZvQUZEb242RlUrNXZURzg1L3dqeTVNQlVHQTFVZEVRUU8KTUF5Q0NtdDFZbVZ5Ym1WMFpYTXdEUVlKS29aSWh2Y05BUUVMQlFBRGdnRUJBRXpybW1wWk9wUDUxVWR5U2Z3VAptSTBLNGVmc2VSd2NDTGovc0t2OUxzN3EvcktteXJWTUQxSWN4Y1JLUTlUQThqTkdZdGhEZWo1UUM2cDM5cE81CkNrK1JrZ2JIMGx1Mm02bUUzRHVBYkhva1IzMFNHTjNkbUwyV01jTGpPQkRrZVg2UUh2RTcyOWNBSFo1clBtQ2QKUEVoc2l0djFWaEdLVWtESFVieVA0N295U2s1bU9aZ25KcmtZTThaK25xZnJCVS9GYmE5eDRXUzBVREJGRzlXSQo3eGJNd2VLR24rR0EydU1XNHBsNUhCK1JqNHlucHh4cEdxaGR2Rk51c2luc0I5U0xsR25YZ0tFdjJSREpNZ3RBClhrWElFNlF2YlZ6RkVSUXlOb2I5TFNpaWgrbTRKTU8rV01Pak1vREIvQ3BHOWFBekVEOHUvSllHa0xFSnArRVIKQlFJPQotLS0tLUVORCBDRVJUSUZJQ0FURS0tLS0t
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
          volumeMounts:
          - mountPath: /etc/site-manager/certs
            name: certs
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
      volumes:
      - name: certs
        secret:
          secretName: site-manager-cert
`
