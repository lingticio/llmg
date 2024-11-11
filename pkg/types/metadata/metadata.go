package metadata

import (
	"net/http"

	"github.com/samber/lo"
)

var _ Identifiable = (*Tenant)(nil)
var _ Identifiable = (*Team)(nil)
var _ Identifiable = (*Group)(nil)

type Identifiable interface {
	ID() string
}

type Tenant struct {
	Id string `json:"id" yaml:"id"`
}

func (t Tenant) ID() string {
	return t.Id
}

type Team struct {
	Id string `json:"id" yaml:"id"`
}

func (t Team) ID() string {
	return t.Id
}

type Group struct {
	Id string `json:"id" yaml:"id"`
}

func (g Group) ID() string {
	return g.Id
}

var _ Metadata = (*UnimplementedMetadata)(nil)

type Metadata interface {
	Tenant() Tenant
	Team() Team
	Group() Group
}

type UnimplementedMetadata struct {
}

func (m UnimplementedMetadata) Tenant() Tenant {
	return Tenant{}
}

func (m UnimplementedMetadata) Team() Team {
	return Team{}
}

func (m UnimplementedMetadata) Group() Group {
	return Group{}
}

type UpstreamOpenAICompatibleChat struct {
	Usage  bool `json:"usage" yaml:"usage"`
	Stream bool `json:"stream" yaml:"stream"`
}

type UpstreamOpenAICompatible struct {
	Chat       UpstreamOpenAICompatibleChat `json:"chat" yaml:"chat"`
	Models     bool                         `json:"models" yaml:"models"`
	Embeddings bool                         `json:"embeddings" yaml:"embeddings"`
	Images     bool                         `json:"images" yaml:"images"`
	Audio      bool                         `json:"audio" yaml:"audio"`
}

type UpstreamOpenAI struct {
	Weight *uint `json:"weight" yaml:"weight"`

	BaseURL      string                   `json:"base_url" yaml:"base_url"`
	APIKey       string                   `json:"api_key" yaml:"api_key"`
	ExtraHeaders http.Header              `json:"extra_headers" yaml:"extra_headers"`
	Compatible   UpstreamOpenAICompatible `json:"compatible" yaml:"compatible"`
}

var _ Upstreamable = (*Upstream)(nil)
var _ Upstreamable = (*Upstreams)(nil)
var _ Upstreamable = (*UpstreamSingleOrMultiple)(nil)

type Upstreamable interface {
	IsSingleUpstream() bool
	GetUpstream() *Upstream
	GetUpstreams() []*Upstream
}

type Upstream struct {
	OpenAI UpstreamOpenAI `json:"openai" yaml:"openai"`
}

func (*Upstream) IsSingleUpstream() bool {
	return true
}

func (u *Upstream) GetUpstream() *Upstream {
	return u
}

func (u *Upstream) GetUpstreams() []*Upstream {
	return []*Upstream{u}
}

type Upstreams []*Upstream

func (Upstreams) IsSingleUpstream() bool {
	return false
}

func (u Upstreams) GetUpstream() *Upstream {
	if len(u) == 0 {
		return nil
	}

	return u[0]
}

func (u Upstreams) GetUpstreams() []*Upstream {
	return lo.Map(u, func(item *Upstream, index int) *Upstream {
		return item
	})
}

type UpstreamSingleOrMultiple struct {
	*Upstream `yaml:",inline"`

	Group Upstreams `json:"group" yaml:"group"`
}

func (u *UpstreamSingleOrMultiple) IsSingleUpstream() bool {
	return len(u.Group) == 0
}

func (u *UpstreamSingleOrMultiple) GetUpstream() *Upstream {
	if u.IsSingleUpstream() {
		return u.Upstream
	}

	return u.Group.GetUpstream()
}

func (u *UpstreamSingleOrMultiple) GetUpstreams() []*Upstream {
	if u.IsSingleUpstream() {
		return []*Upstream{u.Upstream}
	}

	return u.Group.GetUpstreams()
}
