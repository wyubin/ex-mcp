wasm:
	@go mod init main.go && go mod tidy \
    && env GOOS=wasip1 GOARCH=wasm go build -buildmode=c-shared -o main.wasm main.go

up:
	docker-compose up --build
