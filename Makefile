.PHONY: smoke-test
smoke-test:
	go run smoke_test/main.go

.PHONY: prime-test
prime-test:
	go build -o bin/prime -gcflags="-N -l" '1. prime_time/main.go'
	./bin/prime