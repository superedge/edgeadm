package common

import (
	"github.com/superedge/edgeadm/pkg/edgeadm/cmd"
	"github.com/superedge/edgeadm/pkg/edgeadm/constant"
	"github.com/superedge/edgeadm/pkg/edgeadm/constant/manifests"
	"github.com/superedge/edgeadm/pkg/util/kubeclient"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	"path/filepath"
	"time"
)

func DeploySiteManager(clientSet kubernetes.Interface, manifestsDir string, edgeadmConf *cmd.EdgeadmConfig) error {

	if err := EnsureEdgeSystemNamespace(clientSet); err != nil {
		return err
	}

	siteManagerImage, err := GetSuperEdgeImage("site-manager", edgeadmConf)
	if err != nil {
		return err
	}
	option := map[string]interface{}{
		"Namespace":        constant.NamespaceEdgeSystem,
		"SiteManagerImage": siteManagerImage,
	}
	userSiteManagerConfig := filepath.Join(manifestsDir, manifests.APP_SITE_MANAGER)
	siteManagerConfig := ReadYaml(userSiteManagerConfig, manifests.SiteManagerYaml)

	// Deploy site-manager
	err = kubeclient.CreateOrDeleteResourceWithFile(clientSet, nil, siteManagerConfig, option, true)
	if err != nil {
		klog.Errorf("Waiting deploy edge-coredns ds, system message: %v", err)
		return err
	}
	klog.Infof("Deploy %s success!", manifests.APP_SITE_MANAGER)
	return err
}

// DeleteSiteManager uninstalls edge node CoreDNS addon to a Kubernetes cluster
func DeleteSiteManager(kubeconfigFile string, manifestsDir string) error {
	client, err := kubeclient.GetClientSet(kubeconfigFile)
	if err != nil {
		return err
	}

	if err := EnsureEdgeSystemNamespace(client); err != nil {
		return err
	}

	// Delete edge-coredns
	option := map[string]interface{}{
		"Namespace": constant.NamespaceEdgeSystem,
	}

	restCfg, err := clientcmd.BuildConfigFromFlags("", kubeconfigFile)

	dynamicClient, err := dynamic.NewForConfig(restCfg)
	if err != nil {
		klog.Errorf("Failed to get rest kubeclient, error: %v", err)
		return err
	}
	kubeClient, err := kubernetes.NewForConfig(restCfg)
	if err != nil {
		klog.Errorf("Failed to get dynamic kubeclient, error: %v", err)
		return err
	}

	// Delete edge-coredns deploymentGrid
	err = wait.PollImmediate(time.Second, 5*time.Minute, func() (bool, error) {
		err = kubeclient.CreateOrDeleteResourceWithFile(kubeClient, dynamicClient, manifests.SiteManagerYaml, option, false)
		if err != nil {
			klog.Warningf("Waiting deploy edge-coredns DeploymentGrid, system message: %v", err)
			return false, nil
		}
		return true, nil
	})
	klog.Infof("Delete %s success!", manifests.APP_SITE_MANAGER)

	return nil
}
