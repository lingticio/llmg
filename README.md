# llmg

🧘 Extensive LLM endpoints, expended capabilities through your favorite protocols, 🕸️ GraphQL, ↔️ gRPC, ♾️ WebSocket. Extended SOTA support for structured data, function calling, instruction mapping, load balancing, grouping, intelli-routing. Advanced tracing and inference tracking.

## Features

- [ ] Structured data
- [ ] Function calling
- [ ] Instruction mapping
- [ ] Load balancing
- [ ] Grouping
- [ ] Intelli-routing
- [ ] Generative streaming
- [ ] Advanced tracing
- [ ] Inference tracking

## Providers

- [x] OpenAI

## Protocols

- [x] GraphQL
- [ ] gRPC
- [ ] WebSockets
- [ ] RESTful

## Project structure

```
.
├── apis                # Protobuf files
│   ├── jsonapi         # Shared definitions
│   └── gatewayapi      # Business APIs of Gateway
├── cmd                 # Entry for microservices
├── config              # Configuration files
├── graph               # GraphQL Schemas, gqlgen configurations
├── hack                # Scripts for both development, testing, deployment
├── internal            # Actual implementation
│   ├── configs         # Configuration
│   ├── constants       # Constants
│   ├── graph           # GraphQL server & model & resolvers
│   ├── grpc            # gRPC server and client
│   ├── libs            # Libraries
│   └── meta            # Metadata
├── logs                # Logs, excluded from git
├── pkg                 # Public APIs
├── production          # Production related deployment, configurations and scripts
├── release             # Release bundles, excluded from git
├── tools               # Tools
│   └── tools.go        # Pinned tools
├── .dockerignore       # Docker ignore file
├── .editorconfig       # Editor configuration
├── .gitignore          # Git ignore file
├── .golangci.yml       # GolangCI configuration
├── buf.gen.yaml        # How .proto under apis/ are generated
├── buf.yaml            # How buf is configured
├── cspell.config.yaml  # CSpell configuration
└── docker-compose.yml  # Docker compose file, for bootstrapping the needed external services like db, redis, etc.
```

## Stacks involved

- [Go](https://golang.org/)
- [gqlgen](https://gqlgen.com/)
- [gRPC](https://grpc.io/)
- [uber/zap](https://go.uber.org/zap)
- [uber/fx](https://go.uber.org/fx)
- [Docker](https://docker.io/)
- [Grafana Promtail](https://grafana.com/docs/loki/latest/send-data/promtail/)
- [Buf](https://buf.build/)

## Configuration

Copy the `config.example.yaml` to `config.yaml` and update the values as needed.

```shell
cp config.example.yaml config.yaml
```

> [!NOTE]
> When developing locally, you can use the `config.local.yaml` file to override both testing and production configurations without need to worry
> about accidentally committing sensitive information since it is ignored by git.

Besides configurations, you can always use environment variables to override the configurations as well.

## Build

Every microservice and its entry should have similar build steps and usage as follows.

### Build `llmg-grpc`

```shell
go build \
  -a \
  -o "release/lingticio/llmg-grpc" \
  -ldflags " -X './internal/meta.Version=1.0.0' -X './internal/meta.LastCommit=abcdefg'" \
  "./cmd/lingticio/llmg-grpc"
```

### Build `llmg-grpc` with Docker

```shell
docker build \
  --build-arg="BUILD_VERSION=1.0.0" \
  --build-arg="BUILD_LAST_COMMIT=abcdefg" \
  -f cmd/lingticio/llmg-grpc/Dockerfile \
  .
```

## Start the server

### Start `llmg-grpc`

With `config/config.yaml`:

```shell
go run cmd/lingticio/llmg-grpc
```

With `config.local.yaml`:

```shell
go run cmd/lingticio/llmg-grpc -c $(pwd)/config/config.local.yaml
```

## Development

### Adding new queries, mutations, or subscriptions for GraphQL

We use [`gqlgen`](https://gqlgen.com/) to generate the GraphQL server and client codes based on the schema defined in the `graph/${category}/*.graphqls` file.

#### Generate the GraphQL server and client codes

```shell
go generate ./...
```

Once generated, you can start the server and test the queries, mutations, and subscriptions from `internal/graph/${category}/*.resolvers.go`.

### Prepare buf.build Protobuf dependencies

```shell
buf dep update
chmod +x ./hack/proto-export
./hack/proto-export
```

### Adding new services or endpoints

We use [`buf`](https://buf.build/) to manage and generate the APIs based on the Protobuf files.

#### Install `buf`

Follow the instructions here: [Buf - Install the Buf CLI](https://buf.build/docs/installation)

#### Prepare `buf`

```shell
buf dep update
```

#### Create new Protobuf files

Create new Protobuf files under the `apis` directory as following guidelines:

```
.
apis
├── jsonapi             # <shared defs, such as jsonapi>
│   └── jsonapi.proto
└── coreapi             # <api group, such as api, adminapi, you can categorize them by business>
    └── v1              # <version, such as v1>
        └── v1.proto
```

#### Generate the APIs

##### Install `grpc-ecosystem/grpc-gateway-ts` plugin

```shell
go install github.com/grpc-ecosystem/protoc-gen-grpc-gateway-ts
```

Run the following command to generate the needed files:

```shell
buf generate
```

The generated files includes:

1. `*.gw.go` files for gRPC-Gateway
2. `*.pb.go` files for gRPC server and client
3. `*.swagger.json` files for Swagger UI

Then you are off to go.

## [Adding new Test Doubles (a.k.a. Mocks)](https://github.com/maxbrunsfeld/counterfeiter)

To test the gRPC clients and all sorts of objects like this, as well as meet the [SOLID](https://en.wikipedia.org/wiki/SOLID) principle, we use a library called [counterfeiter](https://github.com/maxbrunsfeld/counterfeiter) to generate test doubles and quickly mock out the dependencies for both local defined and third party interfaces.

Generally all the generated test doubles are generated under the `fake` directory that located in the same package as the original interface.

#### Update the existing test doubles

After you have updated the interface, you can run the following command to update and generate the test doubles again freshly:

```bash
go generate ./...
```

#### Generate new test doubles for new interfaces

First you need to ensure the following comment annotation has been added to the package where you hold all the initial references to the interfaces in order to make sure the `go generate` command can find the interfaces:

```go
//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate
```

If the above comment annotation hasn't yet been added, add one please.

Then you can add the following comment annotation to tell counterfeiter to generate the test doubles for the interface:

##### Generate new test doubles for local defined interfaces

```go
//counterfeiter:generate -o <relative path to store the generated test doubles> --fake-name <the name of the generated fake test double, usually starts with Fake> <where the counterfeiter can find the interface> <the interface name>
```

For example:

```go
//counterfeiter:generate -o fake/some_client.go --fake-name FakeClient . Client
type Client struct {
    Method() string
}
```

##### Generate new test doubles for third party interfaces

```go
//counterfeiter:generate -o <relative path to store the generated test doubles> --fake-name <the name of the generated fake test double, usually starts with Fake> <the import path of the interface>
```

For example:

```go
import (
    "github.com/some/package"
)

//counterfeiter:generate -o fake/some_client.go --fake-name FakeClient github.com/some/package.Client

var (
    client package.Client
)
```
