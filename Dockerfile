#FROM kong:2.0.2-centos
FROM kong:2.1.3-centos as builder


USER root

RUN mkdir -p /etc/kong/plugins/

RUN  yum -y install wget gcc && \
    wget https://studygolang.com/dl/golang/go1.15.1.linux-amd64.tar.gz && \
    tar -zxvf go1.15.1.linux-amd64.tar.gz -C /usr/local && \
    mkdir /go && \
    rm go1.15.1.linux-amd64.tar.gz && \
    yum clean all


ENV GOROOT /usr/local/go
ENV PATH $PATH:$HOME/bin:$GOROOT/bin:$GOPATH/bin
ENV GOPATH /go
ENV GOPROXY https://goproxy.cn,direct
ENV GO111MODULE on

COPY . /go/src/rate-limiting
RUN mkdir -p /etc/kong/plugins/ && \
    cd /go/src/rate-limiting && \
    go build -buildmode=plugin -o /etc/kong/plugins/nick-rate-limiting.so && \
    cd /go/src/rate-limiting/go-pluginserver && \
    go build github.com/Kong/go-pluginserver && \
    cp /go/src/rate-limiting/go-pluginserver/go-pluginserver /usr/local/bin

# for debug
#RUN /usr/local/bin/go-pluginserver -version && \
#    cd /etc/kong/plugins && \
#    /usr/local/bin/go-pluginserver -dump-plugin-info nick-rate-limiting


FROM kong:2.1.3-centos 

ENV KONG_DATABASE off
ENV KONG_GO_PLUGINS_DIR /etc/kong/plugins/
#ENV KONG_DECLARATIVE_CONFIG /etc/kong/kong.conf
ENV KONG_PLUGINS bundled,nick-rate-limiting
ENV KONG_PROXY_ACCESS_LOG=/dev/stdout
ENV KONG_ADMIN_ACCESS_LOG=/dev/stdout
ENV KONG_PROXY_ERROR_LOG=/dev/stderr
ENV KONG_ADMIN_ERROR_LOG=/dev/stderr
ENV KONG_ADMIN_LISTEN="0.0.0.0:8001, 0.0.0.0:8444 ssl"
ENV KONG_NGINX_USER="root root"
ENV KONG_PROXY_LISTEN 0.0.0.0:8000
ENV KONG_LOG_LEVEL debug
USER root
RUN  mkdir -p /etc/kong/plugins
COPY --from=builder  /go/src/rate-limiting/go-pluginserver/go-pluginserver /usr/local/bin
COPY --from=builder  /etc/kong/plugins/nick-rate-limiting.so /etc/kong/plugins
