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
package common

import (
	"github.com/superedge/edgeadm/pkg/edgeadm/cmd"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"path/filepath"
	"time"

	"github.com/superedge/edgeadm/pkg/edgeadm/constant"
	"github.com/superedge/edgeadm/pkg/edgeadm/constant/manifests"
	"github.com/superedge/edgeadm/pkg/util/kubeclient"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

// DeployEdgeCorednsAddon installs edge node CoreDNS addon to a Kubernetes cluster
func DeployEdgeCorednsAddon(client kubernetes.Interface, manifestsDir string, edgeadmConf *cmd.EdgeadmConfig) error {

	if err := EnsureEdgeSystemNamespace(client); err != nil {
		return err
	}

	// Deploy edge-coredns config
	option := map[string]interface{}{
		"Namespace":    constant.NamespaceEdgeSystem,
		"CoreDnsImage": GetEdgeDnsImage(edgeadmConf),
	}
	userEdgeCorednsConfig := filepath.Join(manifestsDir, manifests.APPEdgeCorednsConfig)
	edgeCorednsConfig := ReadYaml(userEdgeCorednsConfig, manifests.EdgeCorednsConfigYaml)
	// Waiting DeploymentGrid apply success
	err := kubeclient.CreateResourceWithFile(client, edgeCorednsConfig, option)
	if err != nil {
		klog.Errorf("Deploy edge-coredns config error: %v", err)
		return err
	}

	// Deploy edge-coredns deploymentGrid
	err = wait.PollImmediate(3*time.Second, 5*time.Minute, func() (bool, error) {
		err = kubeclient.CreateOrDeleteResourceWithFile(client, nil, manifests.EdgeCorednsDeploymentGridYaml, option, true)
		if err != nil {
			klog.V(2).Infof("Waiting deploy edge-coredns DeploymentGrid, system message: %v", err)
			return false, nil
		}
		return true, nil
	})
	klog.Infof("Deploy %s success!", manifests.APPEdgeCorednsDeploymentGrid)

	// Deploy edge-coredns serviceGrid
	err = wait.PollImmediate(3*time.Second, 5*time.Minute, func() (bool, error) {
		err = kubeclient.CreateOrDeleteResourceWithFile(client, nil, manifests.EdgeCorednsServiceGridYaml, option, true)
		if err != nil {
			klog.V(2).Infof("Waiting deploy edge-coredns ServiceGrid, system message: %v", err)
			return false, nil
		}
		return true, nil
	})
	klog.Infof("Deploy %s success!", manifests.APPEdgeCorednsServiceGrid)

	return err
}

// DeleteEdgeCoredns uninstalls edge node CoreDNS addon to a Kubernetes cluster
func DeleteEdgeCoredns(kubeconfigFile string, manifestsDir string) error {
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
	userEdgeCorednsConfig := filepath.Join(manifestsDir, manifests.APPEdgeCorednsConfig)
	edgeCorednsConfig := ReadYaml(userEdgeCorednsConfig, manifests.EdgeCorednsConfigYaml)
	// Waiting DeploymentGrid apply success
	err = kubeclient.DeleteResourceWithFile(client, edgeCorednsConfig, option)
	if err != nil {
		klog.Errorf("Deploy edge-coredns config error: %v", err)
		return err
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
		err = kubeclient.CreateOrDeleteResourceWithFile(kubeClient, dynamicClient, manifests.EdgeCorednsDeploymentGridYaml, option, false)
		if err != nil {
			klog.Warningf("Waiting deploy edge-coredns DeploymentGrid, system message: %v", err)
			return false, nil
		}
		return true, nil
	})
	klog.Infof("Delete %s success!", manifests.APPEdgeCorednsDeploymentGrid)

	// Delete edge-coredns serviceGrid
	err = wait.PollImmediate(time.Second, 5*time.Minute, func() (bool, error) {
		err = kubeclient.CreateOrDeleteResourceWithFile(kubeClient, dynamicClient, manifests.EdgeCorednsServiceGridYaml, option, false)
		if err != nil {
			klog.Warningf("Waiting deploy edge-coredns ServiceGrid, system message: %v", err)
			return false, nil
		}
		return true, nil
	})
	klog.Infof("Delete %s success!", manifests.APPEdgeCorednsServiceGrid)

	return nil
}
