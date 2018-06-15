.IPHONY: install build test-unit test-coverage doc benchmark smoke-test docker-build pre-deploy docker-launch docker-stop

install:
	go get -u golang.org/x/lint/golint

build: 
	CGO_ENABLED=0 GOOS=linux go build -o ./limitri -a -ldflags '-s' -installsuffix cgo main.go bombardier.go parser.go template.go

test-unit:
	go test -race -v `go list ./... | grep -v -e /vendor/ -e /mock/`
	go list ./... | grep -v /vendor/ | xargs -L1 golint -set_exit_status
	go vet `go list ./... | grep -v /vendor/`
	
test-coverage:
	go test -cover `go list ./... | grep -v -e /vendor/ -e /mock/`
	go test `go list ./... | grep -v -e /vendor/ -e /mock/` -coverprofile=cover.out
	#go tool cover -html=cover.out
	go tool cover -func=cover.out

benchmark:
	go test -bench=. `go list ./... -race | grep -v -e /vendor/ -e /mock/`

doc:
	godoc -http=:3030

smoke-test: 
#	tput setaf 3; echo "\t\t===== SMOKE / INTEGRATION TEST ====="; tput sgr0;	
	smoke -f smoke-test.yaml -u http://localhost:20001 -v
