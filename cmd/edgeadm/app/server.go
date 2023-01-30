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

package app

import (
	"flag"
	"io"
	"os"
	"path"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	cliflag "k8s.io/component-base/cli/flag"

	"github.com/superedge/edgeadm/pkg/edgeadm/cmd"
	"github.com/superedge/edgeadm/pkg/edgeadm/cmd/addon"
	"github.com/superedge/edgeadm/pkg/edgeadm/cmd/change"
	"github.com/superedge/edgeadm/pkg/edgeadm/cmd/revert"
	"github.com/superedge/edgeadm/pkg/edgeadm/constant"
	"github.com/superedge/edgeadm/pkg/util/kubeadm"

	"k8s.io/klog/v2"
)

var (
	edgeadmConf = cmd.EdgeadmConfig{
		IsEnableEdge:   true,
		WorkerPath:     "/tmp/edgeadm-tmp",
		Kubeconfig:     "~/.kube/config",
		ManifestsDir:   "/tmp/edgeadm-tmp/edge-manifests",
		InstallPkgPath: "",
	}
)

func NewEdgeadmCommand(in io.Reader, out, err io.Writer) *cobra.Command {
	cmds := &cobra.Command{
		Use:   "edgeadm COMMAND [arg...]",
		Short: "edgeadm use to manage edge kubernetes cluster",
		Run: func(cmd *cobra.Command, _ []string) {
			cmd.Help()
		},
	}

	klog.InitFlags(nil)
	defer klog.Flush()

	// add kubeconfig to persistent flags
	globalFlagSet(nil)
	cmds.ResetFlags()

	// edgeadm about change cluster
	cmds.AddCommand(cmd.NewVersionCMD())
	cmds.AddCommand(cmd.NewManifestsCMD())
	cmds.AddCommand(change.NewChangeCMD())
	cmds.AddCommand(revert.NewRevertCMD())

	// edgeadm create edge cluster
	cmds.AddCommand(kubeadm.NewInitCMD(os.Stdout, &edgeadmConf))
	cmds.AddCommand(kubeadm.NewJoinCMD(os.Stdout, &edgeadmConf))
	cmds.AddCommand(kubeadm.NewCmdToken(os.Stdout, os.Stdout))
	cmds.AddCommand(kubeadm.NewResetCMD(os.Stdin, os.Stdout, &edgeadmConf))
	cmds.AddCommand(addon.NewAddonCMD())
	cmds.AddCommand(addon.NewDetachCMD())

	return cmds
}

func globalFlagSet(flagset *flag.FlagSet) {
	if flagset == nil {
		flagset = flag.CommandLine
	}

	flagset.BoolVar(&edgeadmConf.IsEnableEdge, constant.ISEnableEdge, true, "Enable of install edge kubernetes cluster.")
	flagset.StringVar(&edgeadmConf.WorkerPath, "worker-path", "/tmp/edgeadm-tmp", "Worker path of install edge kubernetes cluster.")
	flagset.StringVar(&edgeadmConf.Kubeconfig, "kubeconfig", "~/.kube/config", "The path to the kubeconfig file. [necessary]")

	os.RemoveAll(edgeadmConf.WorkerPath)
	os.MkdirAll(path.Dir(edgeadmConf.WorkerPath+constant.EdgeClusterLogFile), 0755)
	pflag.Set("log_file", edgeadmConf.WorkerPath+constant.EdgeClusterLogFile)

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.CommandLine.SetNormalizeFunc(cliflag.WordSepNormalizeFunc)

	//flag.Parse()
}
