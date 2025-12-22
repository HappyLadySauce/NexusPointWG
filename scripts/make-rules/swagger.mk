
.PHONY: swagger.run
swagger.run: tools.verify.swagger
	@echo "===========> Generating swagger API docs"
	@swag init -g cmd/main.go -o api/swagger/docs
