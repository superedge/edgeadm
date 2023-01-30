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
package steps

import (
	"context"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm"
	phases "k8s.io/kubernetes/cmd/kubeadm/app/cmd/phases/join"
	"k8s.io/kubernetes/cmd/kubeadm/app/cmd/phases/workflow"
	kubeadmconstants "k8s.io/kubernetes/cmd/kubeadm/app/constants"

	"github.com/superedge/edgeadm/pkg/edgeadm/cmd"
	"github.com/superedge/edgeadm/pkg/edgeadm/constant"
	"github.com/superedge/edgeadm/pkg/util"
)

func NewJoinPreparePhase(config *cmd.EdgeadmConfig) workflow.Phase {
	return workflow.Phase{
		Name:  "join-prepare",
		Short: "join prepare for master or edge node",
		Run:   joinPreparePhase,
		RunIf: func(c workflow.RunData) (bool, error) {
			return config.IsEnableEdge, nil
		},
		InheritFlags: []string{},
	}
}

// joinMasterPreparePhase prepare join master logic.
func joinPreparePhase(c workflow.RunData) error {
	data, ok := c.(phases.JoinData)
	if !ok {
		return errors.New("installLiteAPIServer phase invoked with an invalid data struct")
	}

	masterDomain, err := configControlPlaneInfo(data.Cfg())
	if err != nil {
		klog.Errorf("Config ControlPlaneInfo, error: %v")
		return err
	}

	tlsBootstrapCfg, err := data.TLSBootstrapCfg()
	if err != nil {
		return err
	}

	kubeClient, err := initKubeClient(data, tlsBootstrapCfg)
	if err != nil {
		klog.Errorf("Get kube client error: %v", err)
		return err
	}

	// Deletes the bootstrapKubeConfigFile, so the credential used for TLS bootstrap is removed from disk
	defer func() {
		os.Remove(kubeadmconstants.GetBootstrapKubeletKubeConfigPath())
		os.Remove(constant.KubeadmCertPath)
	}()

	// prepare join edge node
	if data.Cfg().ControlPlane == nil {
		if err := prepareJoinEdgeNode(kubeClient, data, masterDomain); err != nil {
			klog.Errorf("Prepare Join edge node, error: %v", err)
			return nil
		}
	}

	return nil
}

func configControlPlaneInfo(joinConfiguration *kubeadm.JoinConfiguration) (string, error) {
	endpoint := joinConfiguration.Discovery.BootstrapToken.APIServerEndpoint
	host, port, err := util.SplitHostPortIgnoreMissingPort(endpoint)
	if err != nil {
		return "", errors.Errorf("Invalid APIServerEndpoint: %s", endpoint)
	}
	if port != "" {
		endpoint = net.JoinHostPort(constant.AddonAPIServerDomain, port)
	} else {
		endpoint = constant.AddonAPIServerDomain
	}
	// if domain instead of ipv4 address was provided, we won't update control plane info
	if net.ParseIP(host) == nil {
		return "", nil
	}
	joinConfiguration.Discovery = kubeadm.Discovery{
		BootstrapToken: &kubeadm.BootstrapTokenDiscovery{
			APIServerEndpoint:        endpoint,
			Token:                    joinConfiguration.Discovery.BootstrapToken.Token,
			CACertHashes:             joinConfiguration.Discovery.BootstrapToken.CACertHashes,
			UnsafeSkipCAVerification: joinConfiguration.Discovery.BootstrapToken.UnsafeSkipCAVerification,
		},
		File:              joinConfiguration.Discovery.File,
		TLSBootstrapToken: joinConfiguration.Discovery.TLSBootstrapToken,
		Timeout:           joinConfiguration.Discovery.Timeout,
	}
	if net.ParseIP(host) != nil {
		hostEntry := host + " " + constant.AddonAPIServerDomain
		ensureHostDNS(hostEntry)
	}
	return host, nil
}

func ensureHostDNS(hostEntry string) error {
	cmds := []string{
		constant.ResetDNSCmd,
		fmt.Sprintf("cat << EOF >>%s \n%s\n%s\n%s\nEOF", constant.HostsFilePath, constant.HostDNSBeginMark, hostEntry, constant.HostDNSEndMark),
	}
	for _, cmd := range cmds {
		if _, _, err := util.RunLinuxCommand(cmd); err != nil {
			klog.Errorf("Running linux command: %s error: %v", cmd, err)
			return err
		}
	}
	return nil
}

func prepareJoinEdgeNode(kubeClient *kubernetes.Clientset, data phases.JoinData, masterDomain string) error {
	joinCfg, err := data.InitCfg()
	if err != nil {
		return err
	}

	// Set kubelet cluster-dns
	edgeInfoConfigMap, err := kubeClient.CoreV1().ConfigMaps(constant.NamespaceEdgeSystem).Get(context.TODO(), constant.EdgeCertCM, metav1.GetOptions{})
	if err != nil {
		klog.Errorf("Get configMap: %s, error: %v", constant.EdgeCertCM, err)
		return err
	}
	edgeCoreDNSIP, ok := edgeInfoConfigMap.Data[constant.EdgeVirtualAddr]
	if !ok {
		return fmt.Errorf("Get lite-apiserver configMap %s value nil\n", constant.LiteAPIServerTLSJSON)
	}
	edgeCoreDNSIP = strings.Replace(edgeCoreDNSIP, "\n", "", -1)

	if joinCfg.NodeRegistration.KubeletExtraArgs == nil {
		joinCfg.NodeRegistration.KubeletExtraArgs = make(map[string]string)
	}
	joinCfg.NodeRegistration.KubeletExtraArgs["cluster-dns"] = edgeCoreDNSIP
	klog.V(4).Infof("Get edge-coredns clusterIP %s", edgeCoreDNSIP)

	// Splicing node delay domain config
	delayDomainHostConfig := ""
	if net.ParseIP(masterDomain) != nil {
		nodeDelayDomain, ok := edgeInfoConfigMap.Data[constant.EdgeNodeDelayDomain]
		if !ok {
			klog.Warningf("Get cluster-info configMap %s value nil\n", constant.EdgeNodeDelayDomain)
		}
		delayDomains := []string{constant.AddonAPIServerDomain}
		tempDelayDomains := strings.SplitAfter(strings.TrimSpace(nodeDelayDomain), "\n")
		for _, domain := range tempDelayDomains {
			if domain != "" {
				delayDomains = append(delayDomains, domain)
			}
		}
		klog.V(4).Infof("Get node delay domain config: %v", delayDomains)
		for _, delayDomain := range delayDomains {
			delayDomainHostConfig += fmt.Sprintf("%s %s\n", masterDomain, delayDomain)
		}
	}

	// Set node host
	nodeHostConfig, ok := edgeInfoConfigMap.Data[constant.EdgeNodeHostConfig]
	if !ok {
		klog.Warningf("Get cluster-info configMap %s value nil\n", constant.EdgeNodeHostConfig)
	}
	nodeHostConfig = fmt.Sprintf("%s\n%s\n", delayDomainHostConfig, nodeHostConfig)
	if err := ensureHostDNS(nodeHostConfig); err != nil {
		klog.Errorf("Set node hosts err: %v", err)
		return err
	}

	// Set registries config
	insecureRegistriesCfg, ok := edgeInfoConfigMap.Data[constant.InsecureRegistries]
	if !ok {
		klog.Warningf("Get cluster-info configMap %s value nil\n", constant.InsecureRegistries)
	}

	var insecureRegistry []string
	insecureRegistryCfg := strings.SplitAfter(strings.TrimSpace(insecureRegistriesCfg), "\n")
	for _, registry := range insecureRegistryCfg {
		if registry != "" {
			registryDomain := strings.ReplaceAll(registry, "\n", "")
			insecureRegistry = append(insecureRegistry, registryDomain)
		}
	}
	insecureRegistries := map[string][]string{
		"insecure-registries": insecureRegistry,
	}
	if len(insecureRegistry) > 0 {
		util.RemoveFile(constant.UserRegistryCfg)
		os.MkdirAll(constant.UserNodeConfigDir, 0755)
		if err := util.WriteWithBufio(constant.UserRegistryCfg, util.ToJson(insecureRegistries)); err != nil {
			klog.Errorf("Write file: %s, err: %v", constant.UserRegistryCfg, err)
			return err
		}
	}

	return nil
}
