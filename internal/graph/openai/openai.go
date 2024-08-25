package openai

import (
	"net/http"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/lingticio/llmg/internal/graph/openai/generated"
	"github.com/lingticio/llmg/internal/graph/openai/resolvers"
	"github.com/nekomeowww/xo/logger"
	"go.uber.org/fx"
)

//go:generate go run github.com/99designs/gqlgen generate --config ../../../graph/openai/gqlgen.yml

type NewGraphQLHandlerParams struct {
	fx.In

	Logger *logger.Logger
}

type GraphQLHandler struct {
	logger *logger.Logger
}

func NewGraphQLHandler() func(params NewGraphQLHandlerParams) *GraphQLHandler {
	return func(params NewGraphQLHandlerParams) *GraphQLHandler {
		return &GraphQLHandler{
			logger: params.Logger,
		}
	}
}

func (h *GraphQLHandler) InstallForEcho(endpoint string, e *echo.Echo) {
	graphqlHandler := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &resolvers.Resolver{Logger: h.logger}}))

	// As documentation of Subscriptions — gqlgen https://gqlgen.com/recipes/subscriptions/
	// has stated, websocket transport is needed for subscriptions.
	//
	// But it is possible for future implementation to use other transport types. (e.g. SSE)
	graphqlHandler.AddTransport(&transport.Websocket{
		Upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
	})

	// As documentation of Subscriptions — gqlgen https://gqlgen.com/recipes/subscriptions/
	// has stated, POST only handler will not gonna work for subscriptions.
	// Therefore e.Any is used to handle all request methods.
	e.Any(endpoint, func(c echo.Context) error {
		graphqlHandler.ServeHTTP(c.Response(), c.Request())
		return nil
	})
}

func (h *GraphQLHandler) InstallPlaygroundForEcho(endpoint string, playgroundEndpoint string, e *echo.Echo) {
	playgroundHandler := playground.Handler("GraphQL", endpoint)

	e.GET(playgroundEndpoint, func(c echo.Context) error {
		playgroundHandler.ServeHTTP(c.Response(), c.Request())
		return nil
	})
}
