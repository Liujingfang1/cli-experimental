//+build wireinject

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

package wirecli

import (
	"io"

	"github.com/google/wire"
	"sigs.k8s.io/cli-experimental/internal/pkg/apply"
	"sigs.k8s.io/cli-experimental/internal/pkg/clik8s"
	"sigs.k8s.io/cli-experimental/internal/pkg/status"
)

// InitializeApplyStatus creates a new *status.Status object
func InitializeStatus(clik8s.ResourceConfigPath, io.Writer) (*status.Status, error) {
	panic(wire.Build(ProviderSet))
}

// InitializeApply creates a new *apply.Apply object
func InitializeApply(clik8s.ResourceConfigPath, io.Writer) (*apply.Apply, error) {
	panic(wire.Build(ProviderSet))
}

// DoStatus creates a new Status object and runs it
func DoStatus(clik8s.ResourceConfigPath, io.Writer) (status.Result, error) {
	panic(wire.Build(ProviderSet))
}

// DoApply creates a new Apply object and runs it
func DoApply(clik8s.ResourceConfigPath, io.Writer) (apply.Result, error) {
	panic(wire.Build(ProviderSet))
}
