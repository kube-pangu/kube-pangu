package kubepartitioner

import (
	"fmt"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type KubeResource struct {
	resource unstructured.Unstructured
}

func NewKubeResource(resource unstructured.Unstructured) *KubeResource {
	return &KubeResource{
		resource: resource,
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

func (k *KubeResource) SetOwnerNodeForPartitionerId(string, string) {

}
func (k *KubeResource) GetOwnerNodeForPartitionerId(string) (error, string) {
	return nil, ""
}

func (k *KubeResource) GetUnstructured() unstructured.Unstructured {
	return k.resource
}
