TOOLS ?=$(BLOCKER_TOOLS) $(CRITICAL_TOOLS) $(TRIVIAL_TOOLS)

.PHONY: tools.install
tools.install: $(addprefix tools.install., $(TOOLS))

.PHONY: tools.install.%
tools.install.%:
	@echo "===========> Installing $*"
	@$(MAKE) install.$*

.PHONY: tools.verify.%
tools.verify.%:
	@powershell -Command "if (-not (Get-Command $* -ErrorAction SilentlyContinue)) { $(MAKE) tools.install.$* }" 2>nul || \
	if ! which $* &>/dev/null 2>&1; then $(MAKE) tools.install.$*; fi

.PHONY: install.swagger
install.swagger:
	@$(GO) install github.com/swaggo/swag/cmd/swag@latest