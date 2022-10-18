package constant

const (
	ISEnableEdge           = "enable-edge"
	DefaultHA              = "default-ha"
	DefaultHAKubeVIP       = "kube-vip"
	InstallPkgPath         = "install-pkg-path"
	ManifestsDir           = "manifests-dir"
	HANetworkInterface     = "interface"
	ContainerRuntime       = "runtime"
	EdgeVersion            = "edge-version"
	PodInfraContainerImage = "pod-infra-container-image"
	EdgeImageRepository    = "edge-image-repository"
	EdgeVirtualAddr        = "edge-virtual-addr"
)

const (
	ControlFormat              = "    "
	InstallPkgPathNote         = "Path of edgeadm kube-* install package"
	InstallPkgNetworkLocation  = ""
	HANetworkDefaultInterface  = "eth0"
	ContainerRuntimeDocker     = "docker"
	ContainerRuntimeContainerd = "containerd"
	ContainerRuntimeNone       = "none"

	DefaultDockerCRISocket     = "/var/run/dockershim.sock"
	DefaultContainerdCRISocket = "/run/containerd/containerd.sock"
	DefaultEdgeVirtualAddr     = "169.254.20.11"
)
