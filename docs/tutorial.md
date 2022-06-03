## Tutorial

### make

- deal with dependency: `go mod tidy`
- build: `make build`
- clean: `make clean`

### Use edgeadm to install SuperEdge

- Convert normal Kubernetes cluster to edge Kubernetes cluster: `edgeadm change --kubeconfig admin.kubeconfig`

- Revert edge Kubernetes cluster to normal Kubernetes cluster: `edgeadm revert --kubeconfig admin.kubeconfig`

- [**More on edgeadm**](./install_via_edgeadm.md)

### Manual installation

- [**Manual installation**](./install_manually.md)
