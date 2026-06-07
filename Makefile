run:
	JWT_SECRET=Dap7IWYkO2Dq0yRxlJnHUBTi7OM7D3liOkhswob6poY= go run .

env:
	go run .

test:
	go test -coverprofile=cover.out && go tool cover -html=cover.out
	
test-race:
	go test -race -coverprofile=cover.out && go tool cover -html=cover.out