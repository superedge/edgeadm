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
	"fmt"
	"github.com/superedge/edgeadm/pkg/edgeadm/cmd"
	certsphase "k8s.io/kubernetes/cmd/kubeadm/app/phases/certs"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/pflag"
	kubeadmapi "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm"
	kubeadmscheme "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/scheme"
	kubeadmapiv1beta2 "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm/v1beta2"
	"k8s.io/kubernetes/cmd/kubeadm/app/cmd/options"
	kubeadminit "k8s.io/kubernetes/cmd/kubeadm/app/cmd/phases/init"
	phases "k8s.io/kubernetes/cmd/kubeadm/app/cmd/phases/join"
	"k8s.io/kubernetes/cmd/kubeadm/app/cmd/phases/workflow"
	cmdutil "k8s.io/kubernetes/cmd/kubeadm/app/cmd/util"
	kubeadmconstants "k8s.io/kubernetes/cmd/kubeadm/app/constants"
	"k8s.io/kubernetes/cmd/kubeadm/app/util/pkiutil"
)

var (
	saKeyLongDesc = fmt.Sprintf(cmdutil.LongDesc(`
		Generate the private key for signing service account tokens along with its public key, and save them into
		%s and %s files.
		If both files already exist, kubeadm skips the generation step and existing files will be used.
		`+cmdutil.AlphaDisclaimer), kubeadmconstants.ServiceAccountPrivateKeyName, kubeadmconstants.ServiceAccountPublicKeyName)

	genericLongDesc = cmdutil.LongDesc(`
		Generate the %[1]s, and save them into %[2]s.cert and %[2]s.key files.%[3]s

		If both files already exist, kubeadm skips the generation step and existing files will be used.
		` + cmdutil.AlphaDisclaimer)
)

var (
	csrOnly bool
	csrDir  string
)

// NewCertsPhase returns the phase for the certs
func NewEdgeCertsPhase(config *cmd.EdgeadmConfig) workflow.Phase {
	return workflow.Phase{
		Name:   "certs",
		Short:  "Certificate generation",
		Phases: newEdgeCertSubPhases(),
		RunIf: func(data workflow.RunData) (bool, error) {
			return config.IsEnableEdge, nil
		},
		Run:  runCerts,
		Long: cmdutil.MacroCommandLongDescription,
	}
}

func localFlags() *pflag.FlagSet {
	set := pflag.NewFlagSet("csr", pflag.ExitOnError)
	options.AddCSRFlag(set, &csrOnly)
	options.AddCSRDirFlag(set, &csrDir)
	return set
}

// newCertSubPhases returns sub phases for certs phase
func newEdgeCertSubPhases() []workflow.Phase {
	subPhases := []workflow.Phase{}

	// All subphase
	allPhase := workflow.Phase{
		Name:           "all",
		Short:          "Generate all certificates",
		InheritFlags:   getCertPhaseFlags("all"),
		RunAllSiblings: true,
	}

	subPhases = append(subPhases, allPhase)

	// This loop assumes that GetDefaultCertList() always returns a list of
	// certificate that is preceded by the CAs that sign them.
	var lastCACert *KubeadmCert
	for _, cert := range GetEdgeCertList() {
		var phase workflow.Phase
		if cert.CAName == "" {
			phase = newCertSubPhase(cert, runCAPhase(cert))
			lastCACert = cert
		} else {
			phase = newCertSubPhase(cert, runCertPhase(cert, lastCACert))
			phase.LocalFlags = localFlags()
		}
		subPhases = append(subPhases, phase)
	}

	// SA creates the private/public key pair, which doesn't use x509 at all
	saPhase := workflow.Phase{
		Name:         "sa",
		Short:        "Generate a private key for signing service account tokens along with its public key",
		Long:         saKeyLongDesc,
		Run:          runCertsSa,
		InheritFlags: []string{options.CertificatesDir},
	}

	subPhases = append(subPhases, saPhase)

	return subPhases
}

func newCertSubPhase(certSpec *KubeadmCert, run func(c workflow.RunData) error) workflow.Phase {
	phase := workflow.Phase{
		Name:  certSpec.Name,
		Short: fmt.Sprintf("Generate the %s", certSpec.LongName),
		Long: fmt.Sprintf(
			genericLongDesc,
			certSpec.LongName,
			certSpec.BaseName,
			getSANDescription(certSpec),
		),
		Run:          run,
		InheritFlags: getCertPhaseFlags(certSpec.Name),
	}
	return phase
}

func getCertPhaseFlags(name string) []string {
	flags := []string{
		options.CertificatesDir,
		options.CfgPath,
		options.CSROnly,
		options.CSRDir,
		options.KubernetesVersion,
	}
	if name == "all" || name == "apiserver" {
		flags = append(flags,
			options.APIServerAdvertiseAddress,
			options.ControlPlaneEndpoint,
			options.APIServerCertSANs,
			options.NetworkingDNSDomain,
			options.NetworkingServiceSubnet,
		)
	}
	return flags
}

func getSANDescription(certSpec *KubeadmCert) string {
	//Defaulted config we will use to get SAN certs
	defaultConfig := &kubeadmapiv1beta2.InitConfiguration{
		LocalAPIEndpoint: kubeadmapiv1beta2.APIEndpoint{
			// GetAPIServerAltNames errors without an AdvertiseAddress; this is as good as any.
			AdvertiseAddress: "127.0.0.1",
		},
	}
	defaultInternalConfig := &kubeadmapi.InitConfiguration{}

	kubeadmscheme.Scheme.Default(defaultConfig)
	if err := kubeadmscheme.Scheme.Convert(defaultConfig, defaultInternalConfig, nil); err != nil {
		return ""
	}

	certConfig, err := certSpec.GetConfig(defaultInternalConfig)
	if err != nil {
		return ""
	}

	if len(certConfig.AltNames.DNSNames) == 0 && len(certConfig.AltNames.IPs) == 0 {
		return ""
	}
	// This mutates the certConfig, but we're throwing it after we construct the command anyway
	sans := []string{}

	for _, dnsName := range certConfig.AltNames.DNSNames {
		if dnsName != "" {
			sans = append(sans, dnsName)
		}
	}

	for _, ip := range certConfig.AltNames.IPs {
		sans = append(sans, ip.String())
	}
	return fmt.Sprintf("\n\nDefault SANs are %s", strings.Join(sans, ", "))
}

func runCertsSa(c workflow.RunData) error {
	data, ok := c.(kubeadminit.InitData)
	if !ok {
		return errors.New("certs phase invoked with an invalid data struct")
	}

	// if external CA mode, skip service account key generation
	if data.ExternalCA() {
		fmt.Printf("[certs] Using existing sa keys\n")
		return nil
	}

	// create the new service account key (or use existing)
	return certsphase.CreateServiceAccountKeyAndPublicKeyFiles(data.CertificateWriteDir(), data.Cfg().ClusterConfiguration.PublicKeyAlgorithm())
}

func runCerts(c workflow.RunData) error {
	data, ok := c.(kubeadminit.InitData)
	if !ok {
		return errors.New("certs phase invoked with an invalid data struct")
	}

	fmt.Printf("[certs] Using certificateDir folder %q\n", data.CertificateWriteDir())
	return nil
}

func runCAPhase(ca *KubeadmCert) func(c workflow.RunData) error {
	return func(c workflow.RunData) error {
		data, ok := c.(kubeadminit.InitData)
		if !ok {
			return errors.New("certs phase invoked with an invalid data struct")
		}

		// if using external etcd, skips etcd certificate authority generation
		if data.Cfg().Etcd.External != nil && ca.Name == "etcd-ca" {
			fmt.Printf("[certs] External etcd mode: Skipping %s certificate authority generation\n", ca.BaseName)
			return nil
		}

		if _, err := pkiutil.TryLoadCertFromDisk(data.CertificateDir(), ca.BaseName); err == nil {
			if _, err := pkiutil.TryLoadKeyFromDisk(data.CertificateDir(), ca.BaseName); err == nil {
				fmt.Printf("[certs] Using existing %s certificate authority\n", ca.BaseName)
				return nil
			}
			fmt.Printf("[certs] Using existing %s keyless certificate authority\n", ca.BaseName)
			return nil
		}

		// if dryrunning, write certificates authority to a temporary folder (and defer restore to the path originally specified by the user)
		cfg := data.Cfg()
		cfg.CertificatesDir = data.CertificateWriteDir()
		defer func() { cfg.CertificatesDir = data.CertificateDir() }()

		// create the new certificate authority (or use existing)
		return CreateCACertAndKeyFiles(ca, cfg)
	}
}

func runCertPhase(cert *KubeadmCert, caCert *KubeadmCert) func(c workflow.RunData) error {
	return func(c workflow.RunData) error {
		data, ok := c.(kubeadminit.InitData)
		if !ok {
			return errors.New("certs phase invoked with an invalid data struct")
		}

		// if using external etcd, skips etcd certificates generation
		if data.Cfg().Etcd.External != nil && cert.CAName == "etcd-ca" {
			fmt.Printf("[certs] External etcd mode: Skipping %s certificate generation\n", cert.BaseName)
			return nil
		}

		if certData, _, err := pkiutil.TryLoadCertAndKeyFromDisk(data.CertificateDir(), cert.BaseName); err == nil {
			caCertData, err := pkiutil.TryLoadCertFromDisk(data.CertificateDir(), caCert.BaseName)
			if err != nil {
				return errors.Wrapf(err, "couldn't load CA certificate %s", caCert.Name)
			}

			if err := certData.CheckSignatureFrom(caCertData); err != nil {
				return errors.Wrapf(err, "[certs] certificate %s not signed by CA certificate %s", cert.BaseName, caCert.BaseName)
			}

			fmt.Printf("[certs] Using existing %s certificate and key on disk\n", cert.BaseName)
			return nil
		}

		if csrOnly {
			fmt.Printf("[certs] Generating CSR for %s instead of certificate\n", cert.BaseName)
			if csrDir == "" {
				csrDir = data.CertificateWriteDir()
			}

			return CreateCSR(cert, data.Cfg(), csrDir)
		}

		// if dryrunning, write certificates to a temporary folder (and defer restore to the path originally specified by the user)
		cfg := data.Cfg()
		cfg.CertificatesDir = data.CertificateWriteDir()
		defer func() { cfg.CertificatesDir = data.CertificateDir() }()

		// create the new certificate (or use existing)
		return CreateCertAndKeyFilesWithCA(cert, caCert, cfg)
	}
}

func newControlPlanePrepareEdgeCertsSubphase() workflow.Phase {
	return workflow.Phase{
		Name:         "certs [egress]",
		Short:        "Generate the certificates for the new control plane components",
		Run:          runControlPlanePrepareCertsPhaseLocal, //NB. eventually in future we would like to break down this in sub phases for each cert or add the --csr option
		InheritFlags: []string{},
	}
}

func runControlPlanePrepareCertsPhaseLocal(c workflow.RunData) error {
	data, ok := c.(phases.JoinData)
	if !ok {
		return errors.New("control-plane-prepare phase invoked with an invalid data struct")
	}

	// Skip if this is not a control plane
	if data.Cfg().ControlPlane == nil {
		return nil
	}

	cfg, err := data.InitCfg()
	if err != nil {
		return err
	}

	fmt.Printf("[certs] Using certificateDir folder %q\n", cfg.CertificatesDir)

	// Generate missing certificates (if any)
	return CreateEdgePKIAssets(cfg)
}

func CreateEdgePKIAssets(cfg *kubeadmapi.InitConfiguration) error {

	// This structure cannot handle multilevel CA hierarchies.
	// This isn't a problem right now, but may become one in the future.

	certList := GetEdgeCertList()

	certTree, err := certList.AsMap().CertTree()
	if err != nil {
		return err
	}

	if err := certTree.CreateTree(cfg); err != nil {
		return errors.Wrap(err, "error creating PKI assets")
	}

	fmt.Printf("[certs] Valid certificates and keys now exist in %q\n", cfg.CertificatesDir)

	// Service accounts are not x509 certs, so handled separately
	return CreateServiceAccountKeyAndPublicKeyFiles(cfg.CertificatesDir, cfg.ClusterConfiguration.PublicKeyAlgorithm())
}
