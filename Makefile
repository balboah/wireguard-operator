TEST_HOST?=192.168.64.80
TEST_FLAGS?=-test.v
APP?=wireguard-operator
PROJECT?=$(shell gcloud -q config get-value project)
VERSION?=$(shell git describe --abbrev=4 --always --tags --dirty)

.PHONY: docker
docker: clean
	gcloud builds submit --config cloudbuild.yaml --substitutions=TAG_NAME=$(VERSION) --project $(PROJECT) .

.PHONY: local_docker
local_docker:
	docker buildx build --platform linux/arm64/v8,linux/amd64 -t $(APP) .

.PHONY: integration_test
integration_test:
	@GOOS=linux GOARCH=arm64 go test -tags integration -c .
	@scp -q $(APP).test $(TEST_HOST):
	@ssh -q $(TEST_HOST) sudo ./$(APP).test -test.run . $(TEST_FLAGS)

.PHONY: clean
clean:
	-@rm $(APP).test wgo

.PHONY: init
init: manifests/.secrets.yaml
	kustomize build | kubectl create --save-config -f -

.PHONY: deploy
deploy: manifests/.secrets.yaml
	kustomize build | kubectl apply -f -
