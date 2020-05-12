HUB_LOCATION := ./cmd/hub
SERVER_LOCATION := ./cmd/server
CLIENT_LOCATION := ./cmd/client
PB_LOCATION := ./internal/pb
BINARY_LOCATION := ./bin

KEY_OUT := $(BINARY_LOCATION)/test.key
CERT_OUT := $(BINARY_LOCATION)/test.crt

.PHONY: all pb dep build_hub build_server build_client clean gencert readcert

all: build_hub build_server build_client

fluentd/fluentd.pb.go: $(PB_LOCATION)/remote/fluentd/fluentd.proto
	@protoc -I $(PB_LOCATION)/remote/fluentd \
			-I ${GOPATH}/src \
			--go_out=plugins=grpc:$(PB_LOCATION)/remote/fluentd \
			--go_opt=paths=source_relative \
			$(PB_LOCATION)/remote/fluentd/fluentd.proto
systemd/systemd.pb.go: $(PB_LOCATION)/remote/systemd/systemd.proto
	@protoc -I $(PB_LOCATION)/remote/systemd \
			-I ${GOPATH}/src \
			--go_out=plugins=grpc:$(PB_LOCATION)/remote/systemd \
			--go_opt=paths=source_relative \
			$(PB_LOCATION)/remote/systemd/systemd.proto
hub/hub.pb.go: $(PB_LOCATION)/local/hub/hub.proto
	@protoc -I $(PB_LOCATION)/local/hub \
			-I ${GOPATH}/src \
			--go_out=plugins=grpc:$(PB_LOCATION)/local/hub \
			--go_opt=paths=source_relative \
			$(PB_LOCATION)/local/hub/hub.proto

pb: fluentd/fluentd.pb.go \
	systemd/systemd.pb.go \
	hub/hub.pb.go  ## compile protocol buffers

dep: ## Get dependencies
	@go get -v -d ./...

setup_build:
	@mkdir -p $(BINARY_LOCATION)

build_hub: gencert setup_build pb dep ## build hub binary
	@go build -i -v -o $(BINARY_LOCATION)/hub $(HUB_LOCATION)

build_server: setup_build pb dep ## build server binary
	@go build -i -v -o $(BINARY_LOCATION)/server $(SERVER_LOCATION)

build_client: setup_build pb dep ## build client binary
	@go build -i -v -o $(BINARY_LOCATION)/client $(CLIENT_LOCATION)

clean: ## delete binary folder
	@rm -rf $(BINARY_LOCATION)

## helpers
gencert: ## generate cert/key pair and output as test.crt/test.key
	@echo "generating self signed cert.."
	openssl req \
		-newkey rsa:2048 -nodes -keyout $(KEY_OUT) \
		-subj '/C=XX/ST=XX/L=XX/O=XX/CN=example.com' \
		-x509 -days 365 -out $(CERT_OUT)

readcert: ## print cert as text
	@echo "reading generated cert.."
	openssl x509 -text -noout -in test.crt
