# --- builder ---
FROM golang:1.24 as builder

ARG BUILD_VERSION
ARG BUILD_LAST_COMMIT

RUN mkdir /app
RUN mkdir /app/llmg

WORKDIR /app/llmg

COPY go.mod /app/llmg/go.mod
COPY go.sum /app/llmg/go.sum

RUN go env
RUN go env -w CGO_ENABLED=0
RUN go mod download

COPY . /app/llmg

RUN go build \
  -a \
  -o "release/lingticio/llmg-grpc" \
  -ldflags " -X './internal/meta.Version=$BUILD_VERSION' -X './internal/meta.LastCommit=$BUILD_LAST_COMMIT'" \
  "./cmd/lingticio/llmg-grpc"

# --- runner ---
FROM debian as runner

RUN apt update && apt upgrade -y && apt install -y ca-certificates curl && update-ca-certificates

COPY --from=builder /app/llmg/release/lingticio/llmg-grpc /app/llmg/release/lingticio/llmg-grpc

RUN mkdir -p /usr/local/bin/lingticio
RUN ln -s /app/llmg/release/lingticio/llmg-grpc /usr/local/bin/lingticio/llmg-grpc

ENV LOG_FILE_PATH /var/log/llmg-services/lingticio/llmg-grpc.log

CMD [ "/usr/local/bin/lingticio/llmg-grpc" ]
