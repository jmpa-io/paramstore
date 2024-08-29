PROJECT = paramstore

example: dist/example ## Builds & runs example service.
	@dist/example

---: ## ---

# include common targets.
include Makefile.common.mk
