.PHONY: build

build:
	docker build -t kong-rate-limiting-golang .

test-run:
	docker run --rm --name kong-rate-limiting-golang \
    -e "KONG_LOG_LEVEL=info" \
    -e "KONG_NGINX_USER=root root" \
    -p 8000:8000 \
    -p 8443:8443 \
    -p 8001:8001 \
    -p 8444:8444 \
    kong-rate-limiting-golang

test-plugins:
	curl http://localhost:8001/ |grep --color nick-rate-limit
