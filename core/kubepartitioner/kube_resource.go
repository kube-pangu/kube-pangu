package kubepartitioner

import "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

type kubeResource struct {
	resource unstructured.Unstructured
}

func NewKubeResource(resource unstructured.Unstructured) *kubeResource {
	return &kubeResource{
		resource: resource,
	}
}

func (k *kubeResource) GetResourcePartitionKeyForPartitionerId(string) (error, string) {
	return nil, ""
}

func (k *kubeResource) SetOwnerNodeForPartitionerId(string, string) {

}
func (k *kubeResource) GetOwnerNodeForPartitionerId(string) (error, string) {
	return nil, ""
}
