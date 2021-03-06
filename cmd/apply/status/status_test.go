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

package status_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/cli-experimental/cmd/apply/status"
	"sigs.k8s.io/cli-experimental/internal/pkg/wirecli/wiretest"
)

var h string

func TestMain(m *testing.M) {
	c, stop, err := wiretest.NewRestConfig()
	if err != nil {
		os.Exit(1)
	}
	defer stop()
	h = c.Host
	os.Exit(m.Run())
}

func setupKustomize(t *testing.T) string {
	f, err := ioutil.TempDir("/tmp", "TestApplyStatus")
	assert.NoError(t, err)
	err = ioutil.WriteFile(filepath.Join(f, "kustomization.yaml"), []byte(`apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
configMapGenerator:
- name: testmap
namespace: default
`), 0644)
	assert.NoError(t, err)
	return f
}

func TestStatus(t *testing.T) {
	f := setupKustomize(t)
	buf := new(bytes.Buffer)

	cmd := status.GetApplyStatusCommand()
	cmd.SetOutput(buf)
	cmd.SetArgs([]string{fmt.Sprintf("--master=%s", h), f})

	assert.NoError(t, cmd.Execute())
	assert.Equal(t, "Doing `cli-experimental apply status`\nResources: 1\n", buf.String())
}
