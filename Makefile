identity:
	@echo "Principal: $(BLUE)$(shell dfx identity get-principal)$(RESET)"
	@echo "Account ID: $(BLUE)$(shell dfx ledger account-id --of-principal $(shell dfx identity get-principal))$(RESET)"
	@echo "Balance: $(BLUE)$(shell dfx ledger balance --network ic)$(RESET)"

build:
	GOOS=linux GOARCH=amd64 go build -o bin/intervention ./cmd/main.go

deploy.all:
	gcloud compute scp ~/dev/ic/intervention/bin/intervention ~/dev/ic/intervention/identity.pem niccologcp1@intervention:~/

.PHONY: identity build deploy.all


