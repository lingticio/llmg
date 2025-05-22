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
  -o "release/lingticio/llmg-graphql" \
  -ldflags " -X './internal/meta.Version=$BUILD_VERSION' -X './internal/meta.LastCommit=$BUILD_LAST_COMMIT'" \
  "./cmd/lingticio/llmg-graphql"

# --- runner ---
FROM debian as runner

RUN apt update && apt upgrade -y && apt install -y ca-certificates curl && update-ca-certificates

COPY --from=builder /app/llmg/release/lingticio/llmg-graphql /app/llmg/release/lingticio/llmg-graphql

RUN mkdir -p /usr/local/bin/lingticio
RUN ln -s /app/llmg/release/lingticio/llmg-graphql /usr/local/bin/lingticio/llmg-graphql

ENV LOG_FILE_PATH /var/log/llmg-services/lingticio/llmg-graphql.log

CMD [ "/usr/local/bin/lingticio/llmg-graphql" ]
