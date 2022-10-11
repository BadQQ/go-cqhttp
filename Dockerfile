#FROM golang:1.17-alpine AS builder
#RUN  sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories
#
#RUN apk add --no-cache clang binutils clang-static musl-dev \
#  && go env -w CC=clang \
#  && go env -w CXX=clang++ \
#  && go env -w GO111MODULE=auto \
#  && go env -w CGO_ENABLED=1 \
#  && go env -w GOPROXY=https://goproxy.cn,direct
FROM golang:1.18-bullseye AS builder
COPY ./sources.list /etc/apt/sources.list
RUN apt-get update && \
    apt-get install -fy  build-essential clang git \
    && go env -w CC=clang \
    && go env -w CXX=clang++ \
    && go env -w GO111MODULE=auto \
    && go env -w CGO_ENABLED=1 \
    && go env -w GOPROXY=https://goproxy.cn,direct

WORKDIR /build

COPY ./ .

RUN set -ex \
    && BUILD=`date +%FT%T%z` \
    && COMMIT_SHA1=`git rev-parse HEAD` \
    && cd /build \
    && go build -ldflags "-s -w -extldflags '-static' -X github.com/Mrs4s/go-cqhttp/internal/base.Version=${COMMIT_SHA1}_._${BUILD}" -v -o cqhttp

FROM alpine:latest
RUN  sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories
RUN apk add --no-cache ffmpeg
COPY ./init.sh /
COPY --from=builder /build/cqhttp /usr/bin/cqhttp
RUN chmod +x /usr/bin/cqhttp && chmod +x /init.sh

ENV BOT_ADAPTER_ENABLE "false"
ENV BOT_ADAPTER_POST_URL "http://bot-adapter:5800/msginput"
ENV BOT_ADAPTER_POST_SECRET "secret"
ENV BOT_ADAPTER_POST_INTERVAL "1500"
ENV BOT_ADAPTER_POST_RETRIES "3"
ENV BOT_ADAPTER_APPID "go-cqhttp"
ENV BOT_ADAPTER_APPSECRET "HGJKLHSADJKLG"
ENV BOT_ADAPTER_GRPC_ADDR "bot-adapter:8001"


WORKDIR /data

ENTRYPOINT [ "/init.sh" ]
