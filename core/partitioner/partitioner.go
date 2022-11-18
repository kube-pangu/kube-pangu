package partitioner

import (
	"github.com/kube-pangu/kube-pangu/core/models"
	"sort"
)

type ConsistentHashPartitioner struct {
	partitionNodesById map[string]models.PartitionNode
	resourceClient     models.ResourcesClient
	partitionerId      string
	hashFunc           func(string) uint64
}

func NewConsistentHashPartitioner(partitionerId string, resourceClient models.ResourcesClient, hashFunc func(string) uint64) *ConsistentHashPartitioner {
	return &ConsistentHashPartitioner{
		partitionNodesById: map[string]models.PartitionNode{},
		resourceClient:     resourceClient,
		partitionerId:      partitionerId,
		hashFunc:           hashFunc,
	}
}

func (p *ConsistentHashPartitioner) AddPartition(node models.PartitionNode) error {
	p.partitionNodesById[node.GetId()] = node
	return nil
}

func (p *ConsistentHashPartitioner) RemovePartition(nodeId string) error {
	delete(p.partitionNodesById, nodeId)
	return nil
}

func (p *ConsistentHashPartitioner) DoPartition() error {
	if len(p.partitionNodesById) == 0 {
		return nil
	}

	seedNodeMapper := map[uint64]models.PartitionNode{}
	var allSeeds []uint64
	for _, node := range p.partitionNodesById {
		for _, seed := range node.GetSeeds() {
			seedNodeMapper[seed] = node
			allSeeds = append(allSeeds, seed)
		}
	}

	sort.Slice(allSeeds, func(i, j int) bool { return allSeeds[i] < allSeeds[j] })

	resources, err := p.resourceClient.QueryResources()
	if err != nil {
		return err
	}

	for _, resource := range resources {
		err, partitionKey := resource.GetResourcePartitionKeyForPartitionerId(p.partitionerId)
		if err != nil {
			continue
		}
		point := p.hashFunc(partitionKey)

		partitionAssigned := false
		for i := 0; i < len(allSeeds)-1; i++ {
			if point >= allSeeds[i] && point < allSeeds[i+1] {
				partitionAssigned = true
				err, ownerId := resource.GetOwnerNodeForPartitionerId(p.partitionerId)
				if err != nil {
					resource.SetOwnerNodeForPartitionerId(p.partitionerId, seedNodeMapper[allSeeds[i]].GetId())
				} else if ownerId != seedNodeMapper[allSeeds[i]].GetId() {
					resource.SetOwnerNodeForPartitionerId(p.partitionerId, seedNodeMapper[allSeeds[i]].GetId())
				}

				break
			}
		}

		if !partitionAssigned {
			err, ownerId := resource.GetOwnerNodeForPartitionerId(p.partitionerId)
			if err != nil {
				resource.SetOwnerNodeForPartitionerId(p.partitionerId, seedNodeMapper[allSeeds[len(allSeeds)-1]].GetId())
			} else if ownerId != seedNodeMapper[allSeeds[len(allSeeds)-1]].GetId() {
				resource.SetOwnerNodeForPartitionerId(p.partitionerId, seedNodeMapper[allSeeds[len(allSeeds)-1]].GetId())
			}
		}
	}

	return nil
}
