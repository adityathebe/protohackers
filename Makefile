.PHONY: smoke-test
smoke-test:
	go build -o bin/smoketest -gcflags="-N -l" '0. smoke_test/main.go'
	./bin/smoketest

.PHONY: prime-test
prime-test:
	go build -o bin/prime -gcflags="-N -l" '1. prime_time/main.go'
	./bin/prime

.PHONY: 3
3:
	go build -o bin/meanstoanend -gcflags="-N -l" '2. means_to_an_end/main.go'
	./bin/meanstoanend