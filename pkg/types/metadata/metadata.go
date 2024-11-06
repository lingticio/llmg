package metadata

import (
	"net/http"

	"github.com/samber/lo"
)

var _ Identifiable = (*TenantID)(nil)
var _ Identifiable = (*TeamID)(nil)
var _ Identifiable = (*GroupID)(nil)

var _ Identifiable = (*UnimplementedTenant)(nil)
var _ Identifiable = (*UnimplementedTeam)(nil)
var _ Identifiable = (*UnimplementedGroup)(nil)

type Identifiable interface {
	ID() string
}

type Tenant interface {
	Identifiable
}

func TenantFromID(id string) Tenant {
	return TenantID{id: id}
}

type TenantID struct {
	id string
}

func (t TenantID) ID() string {
	return t.id
}

var _ Tenant = (*UnimplementedTenant)(nil)

type UnimplementedTenant struct {
}

func (t UnimplementedTenant) ID() string {
	return ""
}

type Team interface {
	Identifiable
}

func TeamFromID(id string) Team {
	return TeamID{id: id}
}

type TeamID struct {
	id string
}

func (t TeamID) ID() string {
	return t.id
}

type UnimplementedTeam struct {
}

func (t UnimplementedTeam) ID() string {
	return ""
}

type Group interface {
	Identifiable
}

type GroupID struct {
	id string
}

func GroupFromID(id string) Group {
	return GroupID{id: id}
}

func (g GroupID) ID() string {
	return g.id
}

type UnimplementedGroup struct {
}

func (g UnimplementedGroup) ID() string {
	return ""
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
	return UnimplementedTenant{}
}

func (m UnimplementedMetadata) Team() Team {
	return UnimplementedTeam{}
}

func (m UnimplementedMetadata) Group() Group {
	return UnimplementedGroup{}
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
