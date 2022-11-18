package integrationtests

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/kube-pangu/kube-pangu/core/kubepartitioner"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"testing"
)

const testConfigMapResourceWithSomeAnnotation = `{
	"apiVersion": "v1",
	"data": {
		"key": "value"
	},
	"kind": "ConfigMap",
	"metadata": {
		"annotations": {
			"annotation": "value"
		},
		"name": "integration-test-map",
		"namespace": "default"
	}
}`

const testConfigMapResourceWithoutAnnotations = `{
	"apiVersion": "v1",
	"data": {
		"key": "value"
	},
	"kind": "ConfigMap",
	"metadata": {
		"name": "integration-test-map",
		"namespace": "default"
	}
}`

const testConfigMapResourceWithPartitionAnnotation = `{
	"apiVersion": "v1",
	"data": {
		"key": "value"
	},
	"kind": "ConfigMap",
	"metadata": {
		"annotations": {
			"partition-key-partitioner-1234": "pkey"
		},
		"name": "integration-test-map",
		"namespace": "default"
	}
}`

func TestResource(t *testing.T) {
	homeDir, _ := os.UserHomeDir()
	kubeConfig, err := clientcmd.BuildConfigFromFlags("", fmt.Sprintf("%s/%s", homeDir, ".kube/config"))
	if err != nil {
		fmt.Printf("error getting Kubernetes config: %v\n", err)
		os.Exit(1)
	}

	dynamicClient, err := dynamic.NewForConfig(kubeConfig)

	createTestAndDeleteResource(t, kubeConfig, dynamicClient, testConfigMapResourceWithoutAnnotations, func(u *unstructured.Unstructured) {
		resource := kubepartitioner.NewKubeResource(u, kubeConfig, dynamicClient)
		err, _ := resource.GetResourcePartitionKeyForPartitionerId("1234")
		shouldNotBeNil(t, err)
	})

	createTestAndDeleteResource(t, kubeConfig, dynamicClient, testConfigMapResourceWithSomeAnnotation, func(u *unstructured.Unstructured) {
		resource := kubepartitioner.NewKubeResource(u, kubeConfig, dynamicClient)
		err, _ := resource.GetResourcePartitionKeyForPartitionerId("1234")
		shouldNotBeNil(t, err)
	})

	createTestAndDeleteResource(t, kubeConfig, dynamicClient, testConfigMapResourceWithPartitionAnnotation, func(u *unstructured.Unstructured) {
		resource := kubepartitioner.NewKubeResource(u, kubeConfig, dynamicClient)
		err, key := resource.GetResourcePartitionKeyForPartitionerId("1234")
		shouldBeNil(t, err)
		shouldBeTrue(t, key == "pkey")
	})

	createTestAndDeleteResource(t, kubeConfig, dynamicClient, testConfigMapResourceWithPartitionAnnotation, func(u *unstructured.Unstructured) {
		resource := kubepartitioner.NewKubeResource(u, kubeConfig, dynamicClient)
		ownerId := resource.GetOwnerNodeForPartitionerId("1234")
		shouldBeTrue(t, ownerId == "")

		err := resource.SetOwnerNodeForPartitionerId("1234", "owner1")
		shouldBeNil(t, err)

		ownerId = resource.GetOwnerNodeForPartitionerId("1234")
		shouldBeTrue(t, ownerId == "owner1")
	})
}

func createTestAndDeleteResource(t *testing.T, kubeConfig *rest.Config, dynamicClient dynamic.Interface, testConfigMapResource string, testFunc func(*unstructured.Unstructured)) {
	dc, err := discovery.NewDiscoveryClientForConfig(kubeConfig)
	mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(dc))
	shouldBeNil(t, err)

	gvk := schema.GroupVersionKind{
		Group:   "",
		Version: "v1",
		Kind:    "ConfigMap",
	}
	mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	shouldBeNil(t, err)

	unstructuredConfigMap := unstructured.Unstructured{}
	json.Unmarshal([]byte(testConfigMapResource), &unstructuredConfigMap)
	resourceObj, err := dynamicClient.Resource(mapping.Resource).Namespace("default").Create(context.Background(), &unstructuredConfigMap, v1.CreateOptions{})
	shouldBeNil(t, err)

	testFunc(resourceObj)

	err = dynamicClient.Resource(mapping.Resource).Namespace("default").Delete(context.Background(), resourceObj.GetName(), v1.DeleteOptions{})
	shouldBeNil(t, err)
}
