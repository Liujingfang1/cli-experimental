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

package pkg

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/kustomize/pkg/inventory"
	"sigs.k8s.io/yaml"
)

func TestCmdWithEmptyInput(t *testing.T) {
	buf := new(bytes.Buffer)
	cmd, done, err := InitializeFakeCmd(buf, nil)
	defer done()
	assert.NoError(t, err)
	assert.NotNil(t, cmd)

	err = cmd.Apply(nil)
	assert.NoError(t, err)

	err = cmd.Prune(nil)
	assert.NoError(t, err)

	err = cmd.Delete(nil)
	assert.NoError(t, err)
}

// setupResourcesV1 create a slice of unstructured
// with two ConfigMaps
// 	one with the inventory annotation
// 	one without the inventory annotation
func setupResourcesV1() []*unstructured.Unstructured {
	r1 := &unstructured.Unstructured{}
	r1.SetGroupVersionKind(schema.GroupVersionKind{
		Version: "v1",
		Kind:    "ConfigMap",
	})
	r1.SetName("cm1")
	r1.SetNamespace("default")
	r2 := &unstructured.Unstructured{}
	r2.SetGroupVersionKind(schema.GroupVersionKind{
		Version: "v1",
		Kind:    "ConfigMap",
	})
	r2.SetName("inventory")
	r2.SetNamespace("default")
	r2.SetAnnotations(map[string]string{
		inventory.InventoryAnnotation:
		"{\"current\":{\"~G_v1_ConfigMap|default|cm1\":null}}",
		inventory.InventoryHashAnnotation: "1234567",
	})
	return []*unstructured.Unstructured{r1, r2}
}

// setupResourcesV2 create a slice of unstructured
// with two ConfigMaps
// 	one with the inventory annotation
// 	one without the inventory annotation
func setupResourcesV2() []*unstructured.Unstructured {
	r1 := &unstructured.Unstructured{}
	r1.SetGroupVersionKind(schema.GroupVersionKind{
		Version: "v1",
		Kind:    "ConfigMap",
	})
	r1.SetName("cm2")
	r1.SetNamespace("default")
	r2 := &unstructured.Unstructured{}
	r2.SetGroupVersionKind(schema.GroupVersionKind{
		Version: "v1",
		Kind:    "ConfigMap",
	})
	r2.SetName("inventory")
	r2.SetNamespace("default")
	r2.SetAnnotations(map[string]string{
		inventory.InventoryAnnotation:
		"{\"current\":{\"~G_v1_ConfigMap|default|cm2\":null}}",
		inventory.InventoryHashAnnotation: "7654321",
	})
	return []*unstructured.Unstructured{r1, r2}
}

/* TestCmd tests Apply/Prune/Delete functions by
taking the following steps:
	1. Initialize a Cmd object
	2. Create the Resources v1
	3. Check that there no existing ConfigMap.

	Call apply and prune for the first version of Configs
	4. Apply the resources and confirm that there are 2 ConfigMaps
	5. Prune the resources and confirm that there are 2 ConfigMaps

	Call apply and prune for the second version of Configs
	6. Create the Resources v2
	7. Apply the resources and confirm that there are 3 ConfigMaps
	8. Prune the resources and confirm that there are 2 ConfigMaps

	Cleanup
	9. Delete the resources and confirm that there is no ConfigMap
*/
func TestCmd(t *testing.T) {
	buf := new(bytes.Buffer)
	cmd, done, err := InitializeFakeCmd(buf, nil)
	defer done()
	assert.NoError(t, err)
	assert.NotNil(t, cmd)

	cmList := &unstructured.UnstructuredList{}
	cmList.SetGroupVersionKind(schema.GroupVersionKind{
		Kind:    "ConfigMapList",
		Version: "v1",
	})

	c := cmd.Applier.DynamicClient
	err = c.List(context.Background(), cmList, "default", nil)
	assert.NoError(t, err)
	assert.Equal(t, len(cmList.Items), 0)

	resources := setupResourcesV1()
	err = cmd.Apply(resources)
	assert.NoError(t, err)
	err = c.List(context.Background(), cmList, "default", nil)
	assert.NoError(t, err)
	assert.Equal(t, len(cmList.Items), 2)

	err = cmd.Prune(resources)
	assert.NoError(t, err)
	err = c.List(context.Background(), cmList, "default", nil)
	assert.NoError(t, err)
	assert.Equal(t, len(cmList.Items), 2)

	resources = setupResourcesV2()
	err = cmd.Apply(resources)
	assert.NoError(t, err)
	err = c.List(context.Background(), cmList, "default", nil)
	assert.NoError(t, err)
	assert.Equal(t, len(cmList.Items), 3)

	err = cmd.Prune(resources)
	assert.NoError(t, err)
	err = c.List(context.Background(), cmList, "default", nil)
	assert.NoError(t, err)
	assert.Equal(t, len(cmList.Items), 2)

	err = cmd.Delete(resources)
	assert.NoError(t, err)
	err = c.List(context.Background(), cmList, "default", nil)
	assert.NoError(t, err)
	assert.Equal(t, len(cmList.Items), 0)
}

var (
	crdString = `
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: crontabs.stable.example.com
spec:
  group: stable.example.com
  versions:
    - name: v1
      served: true
      storage: true
  scope: Namespaced
  names:
    plural: crontabs
    singular: crontab
    kind: CronTab
    shortNames:
    - ct
`
	crString = `
apiVersion: "stable.example.com/v1"
kind: CronTab
metadata:
  name: my-new-cron-object
  namespace: default
spec:
  cronSpec: "* * * * */5"
  image: my-awesome-cron-image
`
)

func setupResources(t *testing.T)(
	*unstructured.Unstructured, *unstructured.Unstructured) {
	obj1 := map[string]interface{}{}
	err := yaml.Unmarshal([]byte(crdString), &obj1)
	assert.NoError(t, err)
	crd := &unstructured.Unstructured{Object: obj1}

	obj2 := map[string]interface{}{}
	err = yaml.Unmarshal([]byte(crString), &obj2)
	assert.NoError(t, err)
	cr := &unstructured.Unstructured{Object: obj2}

	return crd, cr
}

/* TestCmdWithCRDs tests Apply/Prune/Delete functions with CRDs and CRs
   by taking the following steps:
	1. Initialize a Cmd object
	2. Create a CRD
	3. Check that the CRD is installed
	4. Create a CR for that CRD type
	5. Check that the CR is created
	6. Delete the CR
	7. Delete the CRD
 */
func TestCmdWithCRDs(t *testing.T) {
	crd, cr := setupResources(t)
	buf := new(bytes.Buffer)
	cmd, done, err := InitializeFakeCmd(buf, nil)
	defer done()

	assert.NoError(t, err)

	crdList := &unstructured.UnstructuredList{}
	crdList.SetGroupVersionKind(schema.GroupVersionKind{
		Kind:    "CustomResourceDefinition",
		Version: "v1beta1",
		Group: "apiextensions.k8s.io",
	})
	ctx := context.Background()
	c := cmd.Applier.DynamicClient
	err = c.List(ctx, crdList, "", nil)
	assert.NoError(t, err)
	assert.Equal(t, len(crdList.Items), 0)

	err = cmd.Apply([]*unstructured.Unstructured{crd})
	assert.NoError(t, err)

	err = c.List(ctx, crdList, "", nil)
	assert.NoError(t, err)
	assert.Equal(t, len(crdList.Items), 1)

	crList := &unstructured.UnstructuredList{}
	crList.SetGroupVersionKind(schema.GroupVersionKind{
		Kind:    "CronTab",
		Version: "v1",
		Group: "stable.example.com",
	})
	err = c.List(ctx, crList, "default", nil)
	assert.NoError(t, err)
	assert.Equal(t, len(crList.Items), 0)

	err = cmd.Apply([]*unstructured.Unstructured{cr})
	assert.NoError(t, err)

	err = c.List(ctx, crList, "default", nil)
	assert.NoError(t, err)
	assert.Equal(t, len(crList.Items), 1)

	err = cmd.Delete([]*unstructured.Unstructured{cr})
	assert.NoError(t, err)

	err = c.List(ctx, crList, "default", nil)
	assert.NoError(t, err)
	assert.Equal(t, len(crList.Items), 0)

	err = cmd.Delete([]*unstructured.Unstructured{crd})
	assert.NoError(t, err)

	err = c.List(ctx, crdList, "", nil)
	assert.NoError(t, err)
	assert.Equal(t, len(crdList.Items), 0)
}