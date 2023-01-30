package common

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/superedge/edgeadm/pkg/edgeadm/cmd"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	kubeadmapi "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm"
	"path/filepath"
	"time"

	"github.com/superedge/edgeadm/pkg/edgeadm/constant"
	"github.com/superedge/edgeadm/pkg/edgeadm/constant/manifests"
	"github.com/superedge/edgeadm/pkg/util"
	"github.com/superedge/edgeadm/pkg/util/kubeclient"
)

// runCoreDNSAddon installs CoreDNS addon to a Kubernetes cluster
func DeployTunnelAddon(kubeconfigFile string, client kubernetes.Interface, manifestsDir, caCertFile, caKeyFile, tunnelCloudPublicAddr string, certSANs []string) error {

	// Deploy tunnel-cloud
	tunnelCloudToken := util.GetRandToken(32)
	if err := DeployTunnelCloud(client, manifestsDir,
		caCertFile, caKeyFile, tunnelCloudToken, certSANs, nil, nil); err != nil {
		klog.Errorf("Deploy tunnel-cloud, error: %v", err)
		return err
	}
	klog.Infof("Deploy %s success!", manifests.APP_TUNNEL_CLOUD)

	// GetTunnelCloudPort
	tunnelCloudNodePort, err := GetTunnelCloudPort(client)
	if err != nil {
		klog.Errorf("Get tunnel-cloud port, error: %v", err)
		return err
	}

	// Deploy tunnel-edge
	if err = DeployTunnelEdge(client, manifestsDir,
		caCertFile, caKeyFile, tunnelCloudToken, tunnelCloudPublicAddr, tunnelCloudNodePort, nil, nil); err != nil {
		klog.Errorf("Deploy tunnel-edge, error: %v", err)
		return err
	}

	klog.Infof("Deploy %s success!", manifests.APP_TUNNEL_EDGE)

	return err
}

func DeployTunnelCloud(clientSet kubernetes.Interface, manifestsDir, caCertFile, caKeyFile, tunnelCloudToken string, certSANs []string, cfg *kubeadmapi.InitConfiguration, edgeadmConf *cmd.EdgeadmConfig) error {
	tunnelCloudYaml, option, err := getTunnelCloudResource(clientSet, manifestsDir, caCertFile, caKeyFile, tunnelCloudToken, certSANs)
	if err != nil {
		return err
	}
	tunnelImage, err := GetSuperEdgeImage("tunnel", edgeadmConf)
	if err != nil {
		return err
	}
	option.(map[string]interface{})["TunnelImage"] = tunnelImage
	err = kubeclient.CreateResourceWithFile(clientSet, tunnelCloudYaml, option)
	if err != nil {
		return err
	}

	fmt.Println("Create tunnel-cloud.yaml success!")

	return nil
}

func GetTunnelCloudPort(clientSet kubernetes.Interface) (int32, error) {
	var tunnelCloudNodePort int32 = 0
	wait.PollImmediate(time.Second, 5*time.Minute, func() (bool, error) {
		coredns, err := clientSet.CoreV1().Services(
			constant.NamespaceEdgeSystem).Get(context.TODO(), constant.ServiceTunnelCloud, metav1.GetOptions{})
		if err != nil {
			klog.Errorf("Get tunnel-cloud service error: %v", err)
			return false, nil
		}

		for _, port := range coredns.Spec.Ports {
			if port.Name == constant.TunnelNodePortNameGRPG {
				tunnelCloudNodePort = port.NodePort
				return true, nil
			}
		}
		return false, nil
	})

	if tunnelCloudNodePort == 0 {
		return tunnelCloudNodePort, errors.New("Get tunnel-cloud nodePort nil\n")
	}

	return tunnelCloudNodePort, nil
}

func DeployTunnelEdge(clientSet kubernetes.Interface, manifestsDir,
	caCertFile, caKeyFile, tunnelCloudToken, tunnelCloudNodeAddr string, tunnelCloudNodePort int32, cfg *kubeadmapi.InitConfiguration, edgeadmConf *cmd.EdgeadmConfig) error {

	tunnelEdgeYaml, option, err := getTunnelEdgeResource(clientSet, manifestsDir, caCertFile, caKeyFile, tunnelCloudToken, tunnelCloudNodeAddr, tunnelCloudNodePort)
	if err != nil {
		return err
	}

	tunnelImage, err := GetSuperEdgeImage("tunnel", edgeadmConf)
	if err != nil {
		return err
	}
	// Deploy tunnel-edge
	option.(map[string]interface{})["TunnelImage"] = tunnelImage
	err = wait.PollImmediate(3*time.Second, 5*time.Minute, func() (bool, error) {
		err = kubeclient.CreateOrDeleteResourceWithFile(clientSet, nil, tunnelEdgeYaml, option.(map[string]interface{}), true)
		if err != nil {
			klog.V(2).Infof("Waiting deploy tunnel-edge DeploymentGrid, system message: %v", err)
			return false, nil
		}
		return true, nil
	})
	klog.Infof("Deploy %s success!", manifests.APP_TUNNEL_EDGE)
	return nil
}

func DeleteTunnelAddon(client kubernetes.Interface, manifestsDir, caCertFile, caKeyFile string, tunnelCloudPublicAddr string, certSANs []string) error {
	// GetTunnelCloudPort
	tunnelCloudNodePort, err := GetTunnelCloudPort(client)
	if err != nil {
		klog.Errorf("Get tunnel-cloud port, error: %v", err)
		return err
	}

	// Delete tunnel-edge
	tunnelCloudToken := util.GetRandToken(32)
	if err = DeleteTunnelEdge(client, manifestsDir,
		caCertFile, caKeyFile, tunnelCloudToken, tunnelCloudPublicAddr, tunnelCloudNodePort); err != nil {
		klog.Errorf("Deploy tunnel-edge, error: %v", err)
		return err
	}
	klog.Infof("Delete %s success!", manifests.APP_TUNNEL_EDGE)

	// Delete tunnel-cloud
	if err = DeleteTunnelCloud(client, manifestsDir,
		caCertFile, caKeyFile, tunnelCloudToken, certSANs); err != nil {
		klog.Errorf("Deploy tunnel-cloud, error: %v", err)
		return err
	}
	klog.Infof("Delete %s success!", manifests.APP_TUNNEL_CLOUD)

	return err
}

func DeleteTunnelCloud(clientSet kubernetes.Interface, manifestsDir, caCertFile, caKeyFile, tunnelCloudToken string, certSANs []string) error {
	tunnelCloudYaml, option, err := getTunnelCloudResource(clientSet, manifestsDir, caCertFile, caKeyFile, tunnelCloudToken, certSANs)
	if err != nil {
		return err
	}
	err = kubeclient.DeleteResourceWithFile(clientSet, tunnelCloudYaml, option)
	if err != nil {
		return err
	}

	return nil
}

func DeleteTunnelEdge(clientSet kubernetes.Interface, manifestsDir,
	caCertFile, caKeyFile, tunnelCloudToken string, tunnelCloudNodeAddr string, tunnelCloudNodePort int32) error {
	tunnelEdgeYaml, option, err := getTunnelEdgeResource(clientSet, manifestsDir, caCertFile, caKeyFile, tunnelCloudToken, tunnelCloudNodeAddr, tunnelCloudNodePort)
	if err != nil {
		return err
	}

	err = kubeclient.DeleteResourceWithFile(clientSet, tunnelEdgeYaml, option)
	if err != nil {
		return err
	}

	return nil
}

func getTunnelCloudResource(clientSet kubernetes.Interface, manifestsDir, caCertFile, caKeyFile, tunnelCloudToken string, certSANs []string) (string, interface{}, error) {
	nodes, err := clientSet.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return "", nil, err
	}

	var masterIPs []string
	for _, node := range nodes.Items {
		if _, ok := node.Labels[constant.KubernetesDefaultRoleLabel]; ok {
			for _, address := range node.Status.Addresses {
				if address.Type == v1.NodeInternalIP || address.Type == v1.NodeExternalIP {
					masterIPs = append(masterIPs, address.Address)
				}
			}
		}
	}

	dns := []string{
		"tunnel.cloud.io",
		"tunnelcloud.superedge.io",
	}
	dns = append(dns, certSANs...)
	serviceCert, serviceKey, err := GetServiceCert("TunnelCloudService", caCertFile, caKeyFile, dns, masterIPs)
	if err != nil {
		return "", nil, err
	}

	tunnelAnpServerCrt, tunnelAnpServerKey, err := GetServiceCert("TunnelAnpServer", caCertFile, caKeyFile, []string{"tunnel-cloud.edge-system.svc.cluster.local"}, []string{})
	if err != nil {
		return "", nil, err
	}

	option := map[string]interface{}{
		"Namespace":                           constant.NamespaceEdgeSystem,
		"TunnelCloudEdgeToken":                tunnelCloudToken,
		"TunnelPersistentConnectionServerKey": base64.StdEncoding.EncodeToString(serviceKey),
		"TunnelPersistentConnectionServerCrt": base64.StdEncoding.EncodeToString(serviceCert),
		"TunnelAnpServerCet":                  base64.StdEncoding.EncodeToString(tunnelAnpServerCrt),
		"TunnelAnpServerKey":                  base64.StdEncoding.EncodeToString(tunnelAnpServerKey),
	}

	userManifests := filepath.Join(manifestsDir, manifests.APP_TUNNEL_CLOUD)
	tunnelCloudYaml := ReadYaml(userManifests, manifests.TunnelCloudYaml)
	return tunnelCloudYaml, option, nil
}

func getTunnelEdgeResource(clientSet kubernetes.Interface, manifestsDir,
	caCertFile, caKeyFile, tunnelCloudToken string, tunnelCloudNodeAddr string, tunnelCloudNodePort int32) (string, interface{}, error) {

	caCert, _, err := GetRootCartAndKey(caCertFile, caKeyFile)
	if err != nil {
		return "", nil, err
	}

	masterIps, err := GetMasterIps(clientSet)
	if err != nil {
		return "", nil, err
	}
	if tunnelCloudNodeAddr == "" && len(masterIps) > 0 {
		tunnelCloudNodeAddr = masterIps[0]
	}

	option := map[string]interface{}{
		"Namespace":                      constant.NamespaceEdgeSystem,
		"MasterIP":                       tunnelCloudNodeAddr,
		"KubernetesCaCert":               base64.StdEncoding.EncodeToString(caCert),
		"TunnelCloudEdgeToken":           tunnelCloudToken,
		"TunnelPersistentConnectionPort": tunnelCloudNodePort,
	}

	userManifests := filepath.Join(manifestsDir, manifests.APP_TUNNEL_EDGE)
	tunnelEdgeYaml := ReadYaml(userManifests, manifests.TunnelEdgeYaml)

	return tunnelEdgeYaml, option, nil
}
