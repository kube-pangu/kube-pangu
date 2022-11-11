package partitioner

import "github.com/kube-pangu/kube-pangu/core/models"

type inMemoryResourceClient struct {
	resources []models.Resource
}

func (i *inMemoryResourceClient) QueryResources() ([]models.Resource, error) {
	return i.resources, nil
}

var _ models.ResourcesClient = &inMemoryResourceClient{}
