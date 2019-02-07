TEST_HOST?=35.228.23.66

.PHONY: integration_test
integration_test:
	@GOOS=linux GOARCH=amd64 go test -tags integration -c .
	@scp -q wireguard-operator.test $(TEST_HOST):
	@ssh -q $(TEST_HOST) sudo ./wireguard-operator.test -test.run .
