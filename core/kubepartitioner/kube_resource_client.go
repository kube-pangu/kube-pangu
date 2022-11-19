package kubepartitioner

import (
	"context"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"

	"github.com/kube-pangu/kube-pangu/core/models"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
)

type kubeResourceClient struct {
	dynamicClient     dynamic.Interface
	groupVersionKinds []schema.GroupVersionKind
	config            *rest.Config
}

func NewKubeResourceClient(dynamicClient dynamic.Interface, gvks []schema.GroupVersionKind, config *rest.Config) models.ResourcesClient {
	return &kubeResourceClient{
		dynamicClient:     dynamicClient,
		groupVersionKinds: gvks,
		config:            config,
	}
}

func (k *kubeResourceClient) QueryResources() ([]models.Resource, error) {
	var unstructuredList []unstructured.Unstructured
	for _, gvk := range k.groupVersionKinds {
		dc, err := discovery.NewDiscoveryClientForConfig(k.config)
		if err != nil {
			return nil, err
		}
		mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(dc))

		mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
		if err != nil {
			return nil, err
		}

		list, err := k.dynamicClient.Resource(mapping.Resource).List(context.Background(), v1.ListOptions{})
		if err != nil {
			return nil, err
		}
		unstructuredList = append(unstructuredList, list.Items...)
	}

	var resources []models.Resource
	for _, unstructuredObject := range unstructuredList {
		var innerUnstructuredObject unstructured.Unstructured
		unstructuredObject.DeepCopyInto(&innerUnstructuredObject)
		resources = append(resources, NewKubeResource(&innerUnstructuredObject, k.config, k.dynamicClient))
	}
	return resources, nil
}
