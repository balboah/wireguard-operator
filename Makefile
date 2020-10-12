TEST_HOST?=35.228.23.66
TEST_FLAGS?=-test.v
APP?=wireguard-operator
PROJECT?=$(shell gcloud -q config get-value project)

.PHONY: docker
docker: clean
	gcloud builds submit --tag gcr.io/$(PROJECT)/$(APP):latest --project $(PROJECT) .

.PHONY: local_docker
local_docker:
	docker build -t $(APP) .

.PHONY: integration_test
integration_test:
	@GOOS=linux GOARCH=amd64 go test -tags integration -c .
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
