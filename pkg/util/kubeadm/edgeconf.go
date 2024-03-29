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

package kubeadm

import (
	"github.com/pkg/errors"
	"github.com/superedge/edgeadm/pkg/edgeadm/cmd"
	"io/ioutil"
	"k8s.io/klog/v2"
	phases "k8s.io/kubernetes/cmd/kubeadm/app/cmd/phases/join"
	"k8s.io/kubernetes/cmd/kubeadm/app/cmd/phases/workflow"
	cmdutil "k8s.io/kubernetes/cmd/kubeadm/app/cmd/util"
	"os"
	"path/filepath"
)

const (
	IngressYaml = `
apiVersion: apiserver.k8s.io/v1beta1
kind: EgressSelectorConfiguration
egressSelections:
- name: cluster
  connection:
    proxyProtocol: HTTPConnect
    transport:
      tcp:
        url: https://tunnel-cloud.edge-system.svc.cluster.local:8000
        tlsConfig:
          caBundle: /etc/kubernetes/pki/ca.crt
          clientCert: /etc/kubernetes/pki/tunnel-anp-client.crt
          clientKey: /etc/kubernetes/pki/tunnel-anp-client.key
`
	IngressYamlPath = "/etc/kubernetes/kube-apiserver-conf/egress-selector-configuration.yaml"
)

// NewCertsPhase returns the phase for the certs
func NewEdgeConfPhase(config *cmd.EdgeadmConfig) workflow.Phase {
	return workflow.Phase{
		Name:  "edge-config",
		Short: "Config generation",
		Phases: []workflow.Phase{
			{
				Name:  "ingress-config",
				Short: "IngressSelector Config generation",
				RunIf: func(data workflow.RunData) (bool, error) {
					return config.IsEnableEdge, nil
				},
				Run: func(data workflow.RunData) error {
					if err := os.MkdirAll(filepath.Dir(IngressYamlPath), os.FileMode(0755)); err != nil {
						klog.Errorf("Failed create % path, error:%v", IngressYamlPath, err)
						return err
					}
					err := ioutil.WriteFile(IngressYamlPath, []byte(IngressYaml), os.FileMode(0600))
					if err != nil {
						klog.Errorf("Failed to write %s, error:%v", IngressYaml, err)
						return err
					}
					return nil
				},
			},
		},
		Run:  runCerts,
		Long: cmdutil.MacroCommandLongDescription,
	}
}

func newControlPlanePrepareEdgeConfPhase() workflow.Phase {
	return workflow.Phase{
		Name:         "kubeconfig [egress]",
		Short:        "Generate the kubeconfig for the new control plane components",
		Run:          runControlPlanePrepareKubeconfigPhaseLocal, //NB. eventually in future we would like to break down this in sub phases for each kubeconfig
		InheritFlags: []string{},
	}
}
func runControlPlanePrepareKubeconfigPhaseLocal(c workflow.RunData) error {
	data, ok := c.(phases.JoinData)
	if !ok {
		return errors.New("control-plane-prepare phase invoked with an invalid data struct")
	}

	// Skip if this is not a control plane
	if data.Cfg().ControlPlane == nil {
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(IngressYamlPath), os.FileMode(0755)); err != nil {
		klog.Errorf("Failed create % path, error:%v", IngressYamlPath, err)
		return err
	}
	err := ioutil.WriteFile(IngressYamlPath, []byte(IngressYaml), os.FileMode(0600))
	if err != nil {
		klog.Errorf("Failed to write %s, error:%v", IngressYaml, err)
		return err
	}
	return nil
}
