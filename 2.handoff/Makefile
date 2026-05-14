run:
	go run .

env:
	HANDOFF_CONNECT_STRING=mongodb://127.0.0.1:27017/?directConnection=true  go run .

test:
	go test -coverprofile=cover.out && go tool cover -html=cover.out
	
test-race:
	go test -race -coverprofile=cover.out && go tool cover -html=cover.out