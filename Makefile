.PHONY: time
time:
	go run tools/time/main.go

.PHONY: test
test:
	@hack/run-unit-tests.sh

.PHONY: verify
verify:
	@hack/verify-copyright.sh
