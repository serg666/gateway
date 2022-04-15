run:
	./build/bin/gateway -config ./config/config.yml
all:
	go build -o ./build/bin/ ./cmd/gateway/
