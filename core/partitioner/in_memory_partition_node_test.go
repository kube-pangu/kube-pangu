package partitioner

import "github.com/kube-pangu/kube-pangu/core/models"

type inMemoryPartitionNode struct {
	id    string
	seeds []uint64
}

func (i *inMemoryPartitionNode) GetId() string {
	return i.id
}

func (i *inMemoryPartitionNode) GetSeeds() []uint64 {
	return i.seeds
}

var _ models.PartitionNode = &inMemoryPartitionNode{}
