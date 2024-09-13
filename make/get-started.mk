################################################################## Get Started
download:
	go mod download

tidy:
	go mod tidy

upgrade:
	go get -u ./...

kill:
	@[ "${word 2,$(MAKECMDGOALS)}" ] || { echo "Usage: make kill <port_number>"; exit 1; }
	@npx kill-port ${word 2,$(MAKECMDGOALS)}

workspace:
	@cd .. && rm -f go.work
	@cd .. && go work init
	@cd .. && go work use $(shell find .. -maxdepth 2 \( -name "*-service*" -o -name "*-ms-*" -name "*go-*" \) -type d -printf "%f ")


# Phony targets
.PHONY: download tidy upgrade kill workspace
