.PHONY: cli.install
cli.install:
	@which gf > /dev/null 2>&1 || go install github.com/gogf/gf/cmd/gf/v2@latest
