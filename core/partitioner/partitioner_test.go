package partitioner

import (
	"fmt"
	"github.com/kube-pangu/kube-pangu/core/models"
	"github.com/spaolacci/murmur3"
	"testing"
)

func TestPartitionShouldSucceedWithNoPartitionNodes(t *testing.T) {
	resourceClient := &inMemoryResourceClient{
		resources: []models.Resource{},
	}
	chPartitioner := NewConsistentHashPartitioner("1", resourceClient, murmur3HashFunc)
	err := chPartitioner.DoPartition()
	if err != nil {
		t.Error("Error should not be nil")
	}
}

func TestMultipleResourcesWithMultiplePartitionsAndMultipleSeeds(t *testing.T) {
	hashFunc := func(k string) uint64 {
		lookup := map[string]uint64{
			"key1":  0,
			"key2":  1000,
			"key3":  2000,
			"key4":  3000,
			"key5":  4000,
			"key6":  5000,
			"key7":  6000,
			"key8":  7000,
			"key9":  8000,
			"key10": 9000,
			"key11": 12000,
		}

		return lookup[k]
	}
	partitionerId := "12345"

	var resources []models.Resource
	for i := 1; i <= 10; i++ {
		resources = append(resources, &inMemoryResource{
			partitionKeyForPartitioner: map[string]string{
				partitionerId: fmt.Sprintf("key%d", i),
			},
			ownerNodeForPartitioner: map[string]string{},
		})
	}

	resourceClient := &inMemoryResourceClient{
		resources: resources,
	}
	chPartitioner := NewConsistentHashPartitioner(partitionerId, resourceClient, hashFunc)
	chPartitioner.AddPartition(&inMemoryPartitionNode{
		id:    "1",
		seeds: []uint64{10500, 4500},
	})

	err := chPartitioner.DoPartition()
	shouldBeNil(t, err)
	assignedPartitionsMustMatch(t, partitionerId, resources, []string{"1", "1", "1", "1", "1", "1", "1", "1", "1", "1", "1"})
	assignCallsMustMatch(t, partitionerId, resources, []int{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1})

	chPartitioner.AddPartition(&inMemoryPartitionNode{
		id:    "2",
		seeds: []uint64{1500, 8500},
	})

	/*
		1500+ => 2
		4500+ => 1
		8500+ => 2
		10500+ => 1
	*/
	err = chPartitioner.DoPartition()
	shouldBeNil(t, err)
	assignedPartitionsMustMatch(t, partitionerId, resources, []string{"1", "1", "2", "2", "2", "1", "1", "1", "1", "2", "1"})
	assignCallsMustMatch(t, partitionerId, resources, []int{1, 1, 2, 2, 2, 1, 1, 1, 1, 2, 1})

	chPartitioner.AddPartition(&inMemoryPartitionNode{
		id:    "3",
		seeds: []uint64{0, 7500},
	})

	/*
		0+ => 3
		1500+ => 2
		4500+ => 1
		7500 +=> 3
		8500+ => 2
		10500+ => 1
	*/
	err = chPartitioner.DoPartition()
	shouldBeNil(t, err)
	assignedPartitionsMustMatch(t, partitionerId, resources, []string{"3", "3", "2", "2", "2", "1", "1", "1", "3", "2", "1"})
	assignCallsMustMatch(t, partitionerId, resources, []int{2, 2, 2, 2, 2, 1, 1, 1, 2, 2, 1})

	chPartitioner.RemovePartition("1")

	/*
		0+ => 3
		1500+ => 2
		7500 +=> 3
		8500+ => 2
	*/
	err = chPartitioner.DoPartition()
	shouldBeNil(t, err)
	assignedPartitionsMustMatch(t, partitionerId, resources, []string{"3", "3", "2", "2", "2", "2", "2", "2", "3", "2", "2"})
	assignCallsMustMatch(t, partitionerId, resources, []int{2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2})

	chPartitioner.RemovePartition("3")

	/*
		1500+ => 2
		8500+ => 2
	*/
	err = chPartitioner.DoPartition()
	shouldBeNil(t, err)
	assignedPartitionsMustMatch(t, partitionerId, resources, []string{"2", "2", "2", "2", "2", "2", "2", "2", "2", "2", "2"})
	assignCallsMustMatch(t, partitionerId, resources, []int{3, 3, 2, 2, 2, 2, 2, 2, 3, 2, 2})
}

func assignCallsMustMatch(t *testing.T, partitionerId string, resources []models.Resource, assignedTimes []int) {
	for i, resource := range resources {
		memoryResource := resource.(*inMemoryResource)
		shouldBeTrue(t, memoryResource.setOwnerCount == assignedTimes[i])
	}
}

func assignedPartitionsMustMatch(t *testing.T, partitionerId string, resources []models.Resource, assignedPartitions []string) {
	for i, resource := range resources {
		str := resource.GetOwnerNodeForPartitionerId(partitionerId)
		shouldBeTrue(t, assignedPartitions[i] == str)
	}
}

func shouldBeNil(t *testing.T, obj interface{}) {
	if obj != nil {
		t.Error("Should be Nil")
	}
}

func shouldBeTrue(t *testing.T, cond bool) {
	if cond == false {
		t.Error("Should be True")
	}
}

func murmur3HashFunc(key string) uint64 {
	return murmur3.Sum64([]byte(key))
}
