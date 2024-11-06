package metadata

type Identifiable interface {
	ID() string
}

type Tenant interface {
	Identifiable
}

func TenantFromID(id string) Tenant {
	return TenantID{id: id}
}

var _ Tenant = (*TenantID)(nil)

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

var _ Team = (*TeamID)(nil)

func TeamFromID(id string) Team {
	return TeamID{id: id}
}

type TeamID struct {
	id string
}

func (t TeamID) ID() string {
	return t.id
}

var _ Team = (*UnimplementedTeam)(nil)

type UnimplementedTeam struct {
}

func (t UnimplementedTeam) ID() string {
	return ""
}

type Group interface {
	Identifiable
}

var _ Group = (*GroupID)(nil)

type GroupID struct {
	id string
}

func GroupFromID(id string) Group {
	return GroupID{id: id}
}

func (g GroupID) ID() string {
	return g.id
}

var _ Group = (*UnimplementedGroup)(nil)

type UnimplementedGroup struct {
}

func (g UnimplementedGroup) ID() string {
	return ""
}

type Metadata interface {
	Tenant() Tenant
	Team() Team
	Group() Group
}

var _ Metadata = (*UnimplementedMetadata)(nil)

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
