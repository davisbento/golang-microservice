run-server:
	@echo "Running server..."
	@go run ./http-server/server.go

run-worker:
	@echo "Running worker..."
	@go run ./worker/worker.go