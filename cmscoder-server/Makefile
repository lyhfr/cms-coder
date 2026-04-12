# cmscoder-server Makefile
# Top-level build orchestration for all services

SERVICES = web-server user-service

.PHONY: $(SERVICES) build test run docker tidy $(addsuffix -tidy,$(SERVICES))

# Build all services
build: $(SERVICES)

$(SERVICES):
	$(MAKE) -C $@ build

# Run a specific service
run:
ifeq ($(service),)
	@echo "Usage: make run service=web-server"
else
	$(MAKE) -C $(service) run
endif

# Test all services
test:
	@for svc in $(SERVICES); do \
		$(MAKE) -C $$svc test || exit 1; \
	done

# Docker build all services
docker:
	@for svc in $(SERVICES); do \
		$(MAKE) -C $$svc docker; \
	done

# Tidy all services
tidy: $(addsuffix -tidy,$(SERVICES))

%-tidy:
	$(MAKE) -C $(subst -tidy,,$@) tidy
