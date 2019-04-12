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

package resourceconfig

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/kustomize/pkg/fs"
	"sigs.k8s.io/kustomize/pkg/ifc/transformer"
	"sigs.k8s.io/kustomize/pkg/loader"
	"sigs.k8s.io/kustomize/pkg/resmap"
	"sigs.k8s.io/kustomize/pkg/target"
)

// ConfigProvider provides runtime.Objects for a path
type ConfigProvider interface {
	// IsSupported returns true if the ConfigProvider supports the given path
	IsSupported(path string) bool

	// GetConfig returns the Resource Config as runtime.Objects
	GetConfig(path string) ([]runtime.Object, error)
}

// KustomizeProvider provides configs from Kusotmize targets
type KustomizeProvider struct {
	RF *resmap.Factory
	TF transformer.Factory
	FS fs.FileSystem
}

// IsSupported checks if the path is supported by KustomizeProvider
func (p *KustomizeProvider) IsSupported(path string) bool {
	return true
}

// GetConfig returns the resource configs
func (p *KustomizeProvider) GetConfig(path string) ([]runtime.Object, error) {
	ldr, err := loader.NewLoader(path, p.FS)
	if err != nil {
		return nil, err
	}
	defer ldr.Cleanup()

	kt, err := target.NewKustTarget(ldr, p.RF, p.TF)
	if err != nil {
		return nil, err
	}
	allResources, err := kt.MakeCustomizedResMap()
	if err != nil {
		return nil, err
	}
	var results []runtime.Object
	for _, r := range allResources {
		results = append(results, &unstructured.Unstructured{Object: r.Kunstructured.Map()})
	}

	return results, nil
}

// RawConfigFileProvider provides configs from raw K8s configuration files
type RawConfigFileProvider struct{}

// IsSupported checks if a path is a raw K8s configuration file
func (p *RawConfigFileProvider) IsSupported(path string) bool {
	return false
}

// GetConfig returns the resource configs
func (p *RawConfigFileProvider) GetConfig(path string) ([]runtime.Object, error) {
	return nil, nil
}

// RawConfigHTTPProvider provides configs from HTTP urls
type RawConfigHTTPProvider struct{}

// IsSupported returns if the path points to a HTTP url target
func (p *RawConfigHTTPProvider) IsSupported(path string) bool {
	return false
}

// GetConfig returns the resource configs
func (p *RawConfigHTTPProvider) GetConfig(path string) ([]runtime.Object, error) {
	return nil, nil
}
