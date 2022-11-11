mockgen:
	mockgen -source core/models/interfaces.go -destination mocks/core/models/interfaces_mocks.go

test:
	go test github.com/kube-pangu/kube-pangu/core/partitioner

integration-test:
	go test github.com/kube-pangu/kube-pangu/integrationtests