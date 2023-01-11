.PHONY: 0
0:
	go build -o bin/smoketest -gcflags="all=-N -l" '0.smoke_test/main.go'
	./bin/smoketest

.PHONY: 1
1:
	go build -o bin/prime -gcflags="all=-N -l" '1.prime_time/main.go'
	./bin/prime

.PHONY: 2
2:
	go build -o bin/meanstoanend -gcflags="all=-N -l" '2.means_to_an_end/main.go'
	./bin/meanstoanend

.PHONY: 3
3:
	go build -o bin/budgetchat -gcflags="all=-N -l"  github.com/adityathebe/protohackers/3.budget_chat
	./bin/budgetchat

.PHONY: 4
4:
	go build -o bin/unusualdatagramprotocol -gcflags="all=-N -l" '4.unusual_database_program/main.go'
	./bin/unusualdatagramprotocol

.PHONY: 5
5:
	go build -o bin/mobinthemiddle -gcflags="all=-N -l" '5.mob_in_the_middle/main.go'
	./bin/mobinthemiddle

.PHONY: 6
6:
	go build -o bin/speeddaemon -gcflags="all=-N -l" 'github.com/adityathebe/protohackers/6.speed_daemon/...'
	./bin/speeddaemon

.PHONY: 7
7:
	go build -o bin/lrcp -gcflags="all=-N -l" 'github.com/adityathebe/protohackers/7.line_reversal/...'
	./bin/lrcp

.PHONY: 9
9:
	go build -o bin/jobcentre -gcflags="all=-N -l" github.com/adityathebe/protohackers/jobcentre/cmd/...
	./bin/jobcentre