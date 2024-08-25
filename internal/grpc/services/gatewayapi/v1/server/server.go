package serverv1

import (
	"context"

	gatewayapiv1 "github.com/lingticio/llmg/apis/gatewayapi/v1"
	"github.com/lingticio/llmg/internal/meta"
)

type ServerService struct {
	gatewayapiv1.UnimplementedServerServiceServer
}

func NewServerService() func() *ServerService {
	return func() *ServerService {
		return &ServerService{}
	}
}

func (s *ServerService) Release(ctx context.Context, req *gatewayapiv1.ReleaseRequest) (*gatewayapiv1.ReleaseResponse, error) {
	return &gatewayapiv1.ReleaseResponse{
		Version:    meta.Version,
		LastCommit: meta.LastCommit,
	}, nil
}
