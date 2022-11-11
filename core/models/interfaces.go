package models

type PartitionNode interface {
	GetId() string
	GetSeeds() []uint64
}

type Resource interface {
	GetResourcePartitionKeyForPartitionerId(string) (error, string)
	SetOwnerNodeForPartitionerId(string, string)
	GetOwnerNodeForPartitionerId(string) (error, string)
}

type ResourcesClient interface {
	QueryResources() ([]Resource, error)
}

type Partitioner interface {
	AddPartitionNode(PartitionNode) error
	RemovePartition(string) error
	DoPartition() error
}
