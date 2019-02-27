TEST_HOST?=35.228.23.66
TEST_FLAGS?=-test.v
APP?=wireguard-operator
KUBECONFIG?=$(CURDIR)/../kube/kubeadm/admin.conf
export KUBECONFIG

.PHONY: docker
docker: clean
	gcloud builds submit --config cloudbuild.yaml .

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
	kubectl create --save-config -f manifests

.PHONY: deploy
deploy: manifests/.secrets.yaml
	kubectl apply -f manifests

manifests/.secrets.yaml: manifests/.secrets.yaml.kms
	gcloud kms decrypt --location global \
		--keyring blokada --key kubernetes-secrets \
		--plaintext-file $@ \
		--ciphertext-file $@.kms

manifests/.secrets.yaml.kms: manifests/.secrets.yaml
	gcloud kms encrypt --location global \
		--keyring blokada --key kubernetes-secrets \
		--plaintext-file manifests/.secrets.yaml \
		--ciphertext-file $@
