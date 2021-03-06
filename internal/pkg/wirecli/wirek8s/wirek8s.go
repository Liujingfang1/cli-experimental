/*
Copyright 2019 The Kubernetes Authors.
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

package wirek8s

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/google/wire"
	"github.com/spf13/cobra"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/cli-experimental/internal/pkg/clik8s"
	"sigs.k8s.io/cli-experimental/internal/pkg/resourceconfig"
	"sigs.k8s.io/cli-experimental/internal/pkg/util"
	"sigs.k8s.io/kustomize/k8sdeps"
	"sigs.k8s.io/kustomize/pkg/factory"
	"sigs.k8s.io/kustomize/pkg/fs"
	"sigs.k8s.io/kustomize/pkg/ifc/transformer"
	"sigs.k8s.io/kustomize/pkg/resmap"
	"sigs.k8s.io/yaml"

	// for connecting to various types of hosted clusters
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

// ConfigProviderSet defines dependencies for initializing ConfigProvider
var ConfigProviderSet = wire.NewSet(
	NewKustomizeFactory, NewResMapFactory, NewTransformerFactory,
	NewFileSystem, NewConfigProvider)

// ProviderSet defines dependencies for initializing Kubernetes objects
var ProviderSet = wire.NewSet(NewKubernetesClientSet, NewKubeConfigPathFlag, NewRestConfig,
	NewMasterFlag, NewResourceConfig, ConfigProviderSet)
var kubeConfigPathFlag string
var master string

// Flags registers flags for talkig to a Kubernetes cluster
func Flags(command *cobra.Command) {
	var path string
	if len(util.HomeDir()) > 0 {
		path = filepath.Join(util.HomeDir(), ".kube", "config")
	} else {
		path = ""
		command.MarkFlagRequired("kubeconfig")
	}
	command.Flags().StringVar(&kubeConfigPathFlag,
		"kubeconfig", path, "absolute path to the kubeconfig file")
	command.Flags().StringVar(&master,
		"master", "", "address of master")
}

// NewKubeConfigPathFlag provides the path to the kubeconfig file
func NewKubeConfigPathFlag() clik8s.KubeConfigPath {
	return clik8s.KubeConfigPath(kubeConfigPathFlag)
}

// NewMasterFlag returns the MasterURL parsed from the `--master` flag
func NewMasterFlag() clik8s.MasterURL {
	return clik8s.MasterURL(master)
}

// NewRestConfig returns a new rest.Config parsed from --kubeconfig and --master
func NewRestConfig(master clik8s.MasterURL, path clik8s.KubeConfigPath) (*rest.Config, error) {
	return clientcmd.BuildConfigFromFlags(string(master), string(path))
}

// NewKustomizeFactory returns a new factory need by Kustomize
func NewKustomizeFactory() *factory.KustFactory {
	return k8sdeps.NewFactory()
}

// NewResMapFactory returns a new resmap factory
func NewResMapFactory(f *factory.KustFactory) *resmap.Factory {
	return f.ResmapF
}

// NewTransformerFactory returns a new transformer factory
func NewTransformerFactory(f *factory.KustFactory) transformer.Factory {
	return f.TransformerF
}

// NewFileSystem returns a new filesystem
func NewFileSystem() fs.FileSystem {
	return fs.MakeRealFS()
}

// NewConfigProvider returns a new ConfigProvider
func NewConfigProvider(rf *resmap.Factory, fSys fs.FileSystem, tf transformer.Factory) resourceconfig.ConfigProvider {
	return &resourceconfig.KustomizeProvider{
		RF: rf,
		TF: tf,
		FS: fSys,
	}
}

// NewKubernetesClientSet provides a clientset for talking to k8s clusters
func NewKubernetesClientSet(c *rest.Config) (*kubernetes.Clientset, error) {
	return kubernetes.NewForConfig(c)
}

// NewResourceConfig provides ResourceConfigs read from the ResourceConfigPath and FileSystem.
func NewResourceConfig(rcp clik8s.ResourceConfigPath, cp resourceconfig.ConfigProvider) (clik8s.ResourceConfigs, error) {
	p := string(rcp)
	var values clik8s.ResourceConfigs

	if cp.IsSupported(p) {
		return cp.GetConfig(p)
	}

	r, err := doFile(p)
	if err != nil {
		return nil, err
	}
	values = append(values, r...)

	return values, nil
}

func doFile(p string) (clik8s.ResourceConfigs, error) {
	var values clik8s.ResourceConfigs

	// Don't allow running on kustomization.yaml, prevents weird things like globbing
	if filepath.Base(p) == "kustomization.yaml" {
		return nil, fmt.Errorf(
			"cannot run on kustomization.yaml - use the directory (%v) instead", filepath.Dir(p))
	}

	// Resource file
	b, err := ioutil.ReadFile(p)
	if err != nil {
		return nil, err
	}
	objs := strings.Split(string(b), "---")
	for _, o := range objs {
		body := map[string]interface{}{}

		if err := yaml.Unmarshal([]byte(o), &body); err != nil {
			return nil, err
		}
		values = append(values, &unstructured.Unstructured{Object: body})
	}

	return values, nil
}
