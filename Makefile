build:
	@go build -o bin/fs

run: build
	@./bin/fs

test:
	@go test ./... -v

t:
	@telnet localhost 3000