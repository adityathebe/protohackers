.PHONY: 0
0:
	go build -o bin/smoketest -gcflags="-N -l" '0.smoke_test/main.go'
	./bin/smoketest

.PHONY: 1
1:
	go build -o bin/prime -gcflags="-N -l" '1.prime_time/main.go'
	./bin/prime

.PHONY: 2
2:
	go build -o bin/meanstoanend -gcflags="-N -l" '2.means_to_an_end/main.go'
	./bin/meanstoanend

.PHONY: 3
3:
	go build -o bin/budgetchat -gcflags="-N -l"  github.com/adityathebe/protohackers/3.budget_chat
	./bin/budgetchat

.PHONY: 4
4:
	go build -o bin/unusualdatagramprotocol -gcflags="-N -l" '4.unusual_database_program/main.go'
	./bin/unusualdatagramprotocol

.PHONY: 9
9:
	go build -o bin/jobcentre -gcflags="-N -l" github.com/adityathebe/protohackers/jobcentre/cmd/...
	./bin/jobcentre