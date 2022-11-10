mockgen:
	mockgen -source core/models/interfaces.go -destination mocks/core/models/interfaces_mocks.go

test:
	go test ./...