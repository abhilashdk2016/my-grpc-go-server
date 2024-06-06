GO_MODULE := github.com/abhilashdk2016/my-grpc-go-server

ifeq ($(OS), Windows_NT)
	BIN_FILENAME  := my-grpc-server.exe
else
	BIN_FILENAME  := my-grpc-server
endif

.PHONY: clean
clean:
ifeq ($(OS), Windows_NT)
	if exist "protogen" rd /s /q protogen
	mkdir protogen\go
else
	rm -fR ./protogen 
	mkdir -p ./protogen/go
endif

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: protoc-go
protoc-go:
	protoc --go_opt=module=${GO_MODULE} --go_out=. \
	--go-grpc_opt=module=${GO_MODULE} --go-grpc_out=. \
	./proto/hello/*.proto ./proto/payment/*.proto ./proto/transaction/*.proto \

.PHONY: build
build: clean tidy
	go build -o ./bin/${BIN_FILENAME} ./cmd

.PHONY: execute
execute: build
	./bin/${BIN_FILENAME}