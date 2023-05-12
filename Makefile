.PHONY: time
time:
	go run tools/time/main.go

.PHONY: verify
verify:
	@hack/verify-copyright.sh
