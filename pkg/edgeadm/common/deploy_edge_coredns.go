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
		"Namespace":       constant.NamespaceEdgeSystem,
		"CoreDnsImage":    GetEdgeDnsImage(edgeadmConf),
		"EdgeVirtualAddr": edgeadmConf.EdgeVirtualAddr,
	}
	userEdgeCorednsConfig := filepath.Join(manifestsDir, manifests.APP_Edge_Coredns)
	edgeCorednsConfig := ReadYaml(userEdgeCorednsConfig, manifests.EdgeCorednsYaml)

	// Deploy edge-coredns
	err := kubeclient.CreateOrDeleteResourceWithFile(client, nil, edgeCorednsConfig, option, true)
	if err != nil {
		klog.Errorf("Waiting deploy edge-coredns ds, system message: %v", err)
		return err
	}
	klog.Infof("Deploy %s success!", manifests.APP_Edge_Coredns)
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
		err = kubeclient.CreateOrDeleteResourceWithFile(kubeClient, dynamicClient, manifests.EdgeCorednsYaml, option, false)
		if err != nil {
			klog.Warningf("Waiting deploy edge-coredns DeploymentGrid, system message: %v", err)
			return false, nil
		}
		return true, nil
	})
	klog.Infof("Delete %s success!", manifests.APP_Edge_Coredns)

	return nil
}
