include ./hack/hack-cli.mk

.PHONY: up
up: cli.install
	@gf up -a

.PHONY: build
build: cli.install
	@gf build -ew

.PHONY: ctrl
ctrl: cli.install
	@gf gen ctrl

.PHONY: dao
dao: cli.install
	@gf gen dao

.PHONY: service
service: cli.install
	@gf gen service

.PHONY: image
image: cli.install
	$(eval _TAG  = $(shell git describe --dirty --always --tags --abbrev=8 --match 'v*' | sed 's/-/./2' | sed 's/-/./2'))
ifneq (, $(shell git status --porcelain 2>/dev/null))
	$(eval _TAG  = $(_TAG).dirty)
endif
	$(eval _TAG  = $(if ${TAG},  ${TAG}, $(_TAG)))
	$(eval _PUSH = $(if ${PUSH}, ${PUSH}, ))
	@gf docker ${_PUSH} -tn $(DOCKER_NAME):${_TAG};

.PHONY: image.push
image.push:
	@make image PUSH=-p;

.PHONY: deploy
deploy:
	$(eval _TAG = $(if ${TAG},  ${TAG}, develop))
	@set -e; \
	mkdir -p $(ROOT_DIR)/temp/kustomize;\
	cd $(ROOT_DIR)/manifest/deploy/kustomize/overlays/${_ENV};\
	kustomize build > $(ROOT_DIR)/temp/kustomize.yaml;\
	kubectl   apply -f $(ROOT_DIR)/temp/kustomize.yaml; \
	if [ $(DEPLOY_NAME) != "" ]; then \
		kubectl   patch -n $(NAMESPACE) deployment/$(DEPLOY_NAME) -p "{\"spec\":{\"template\":{\"metadata\":{\"labels\":{\"date\":\"$(shell date +%s)\"}}}}}"; \
	fi;
