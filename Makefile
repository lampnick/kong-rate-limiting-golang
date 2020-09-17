.PHONY: build-centos

IMAGE_NAME ?= "lampnick/kong-rate-limiting-plugin-golang"
CENTOS_IMAGE_TAG ?= "kong-2.1.3-centos-v1.1.0"
ALPINE_IMAGE_TAG ?= "kong-2.1.3-alpine-v1.1.0"

build-centos:
	@docker build -t $(IMAGE_NAME):$(CENTOS_IMAGE_TAG) .

build-alpine:
	@docker build -f Dockerfile.alpine -t $(IMAGE_NAME):$(ALPINE_IMAGE_TAG) .

test-run-centos:
	@docker run --rm --name $(IMAGE_NAME) \
    -e "KONG_LOG_LEVEL=info" \
    -e "KONG_NGINX_USER=root root" \
    -p 8000:8000 \
    -p 8443:8443 \
    -p 8001:8001 \
    -p 8444:8444 \
     $(IMAGE_NAME):$(CENTOS_IMAGE_TAG)

test-run-alpine:
	@docker run --rm --name $(IMAGE_NAME) \
    -e "KONG_LOG_LEVEL=info" \
    -e "KONG_NGINX_USER=root root" \
    -p 8000:8000 \
    -p 8443:8443 \
    -p 8001:8001 \
    -p 8444:8444 \
    $(IMAGE_NAME):$(ALPINE_IMAGE_TAG)

test-plugins:
	@curl -s http://localhost:8001/ |grep --color nick-rate-limit
