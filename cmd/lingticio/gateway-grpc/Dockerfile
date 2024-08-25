# --- builder ---
FROM golang:1.22 as builder

ARG BUILD_VERSION
ARG BUILD_LAST_COMMIT

RUN mkdir /app
RUN mkdir /app/gateway

WORKDIR /app/gateway

COPY go.mod /app/gateway/go.mod
COPY go.sum /app/gateway/go.sum

RUN go env
RUN go env -w CGO_ENABLED=0
RUN go mod download

COPY . /app/gateway

RUN go build \
  -a \
  -o "release/lingticio/gateway-grpc" \
  -ldflags " -X './internal/meta.Version=$BUILD_VERSION' -X './internal/meta.LastCommit=$BUILD_LAST_COMMIT'" \
  "./cmd/lingticio/gateway-grpc"

# --- runner ---
FROM debian as runner

RUN apt update && apt upgrade -y && apt install -y ca-certificates curl && update-ca-certificates

COPY --from=builder /app/gateway/release/lingticio/gateway-grpc /app/gateway/release/lingticio/gateway-grpc

RUN mkdir -p /usr/local/bin/lingticio
RUN ln -s /app/gateway/release/lingticio/gateway-grpc /usr/local/bin/lingticio/gateway-grpc

ENV LOG_FILE_PATH /var/log/gateway-services/lingticio/gateway-grpc.log

CMD [ "/usr/local/bin/lingticio/gateway-grpc" ]
