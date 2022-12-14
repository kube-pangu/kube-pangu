package partitioner

import (
	"fmt"
	"github.com/kube-pangu/kube-pangu/core/models"
)

type inMemoryResource struct {
	partitionKeyForPartitioner map[string]string
	ownerNodeForPartitioner    map[string]string
	setOwnerCount              int
}

func (i *inMemoryResource) SetOwnerNodeForPartitionerId(s string, s2 string) error {
	i.ownerNodeForPartitioner[s] = s2
	i.setOwnerCount = i.setOwnerCount + 1
	return nil
}

func (i *inMemoryResource) GetOwnerNodeForPartitionerId(s string) string {
	id, ok := i.ownerNodeForPartitioner[s]
	if !ok {
		return ""
	}

	return id
}

func (i *inMemoryResource) GetResourcePartitionKeyForPartitionerId(partitionerId string) (error, string) {
	id, ok := i.partitionKeyForPartitioner[partitionerId]
	if !ok {
		return fmt.Errorf("PartitionerId not found"), ""
	}

	return nil, id
}

var _ models.Resource = &inMemoryResource{}
