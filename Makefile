NAME=terraform-provider-k8s

.PHONY: build
build:
	go build -o $(NAME) main.go
