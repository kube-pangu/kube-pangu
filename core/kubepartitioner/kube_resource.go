package kubepartitioner

import (
	"context"
	"fmt"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
)

type KubeResource struct {
	resource      *unstructured.Unstructured
	config        *rest.Config
	dynamicClient dynamic.Interface
}

func NewKubeResource(resource *unstructured.Unstructured, config *rest.Config, dynamicClient dynamic.Interface) *KubeResource {
	return &KubeResource{
		resource:      resource,
		config:        config,
		dynamicClient: dynamicClient,
	}
}

func (k *KubeResource) GetResourcePartitionKeyForPartitionerId(partitionerId string) (error, string) {
	annotationsMap := k.resource.GetAnnotations()
	val, ok := annotationsMap[fmt.Sprintf("partition-key-partitioner-%s", partitionerId)]
	if !ok {
		return fmt.Errorf("partition key not defined for the partitioner"), ""
	}
	return nil, val
}

func (k *KubeResource) SetOwnerNodeForPartitionerId(partitionerId string, owner string) error {
	dc, err := discovery.NewDiscoveryClientForConfig(k.config)
	if err != nil {
		return err
	}
	mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(dc))
	gvk := k.resource.GroupVersionKind()
	mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return err
	}

	annotationsMap := k.resource.GetAnnotations()
	annotationsMap[fmt.Sprintf("owner-for-partitioner-%s", partitionerId)] = owner
	k.resource.SetAnnotations(annotationsMap)

	_, err = k.dynamicClient.Resource(mapping.Resource).Namespace(k.resource.GetNamespace()).Update(context.Background(), k.resource, v1.UpdateOptions{})
	return err
}
func (k *KubeResource) GetOwnerNodeForPartitionerId(partitionerId string) string {
	val, ok := k.resource.GetAnnotations()[fmt.Sprintf("owner-for-partitioner-%s", partitionerId)]
	if !ok {
		return ""
	}

	return val
}

func (k *KubeResource) GetUnstructured() *unstructured.Unstructured {
	return k.resource
}
