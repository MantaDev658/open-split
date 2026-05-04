package domain

// GroupID is the unique identifier for a Group.
type GroupID string

// Group is the aggregate root for a shared expense pool.
type Group struct {
	ID      GroupID
	Name    string
	Members []UserID
}

// NewGroup validates inputs and creates a Group with creator as its first member.
func NewGroup(id GroupID, name string, creator UserID) (*Group, error) {
	if name == "" {
		return nil, ErrEmptyGroupName
	}

	return &Group{
		ID:      id,
		Name:    name,
		Members: []UserID{creator},
	}, nil
}

// HasMember reports whether userID belongs to the group.
func (g *Group) HasMember(userID UserID) bool {
	for _, member := range g.Members {
		if member == userID {
			return true
		}
	}
	return false
}

// AddMember appends userID to the group, returning ErrUserAlreadyInGroup if already present.
func (g *Group) AddMember(userID UserID) error {
	if g.HasMember(userID) {
		return ErrUserAlreadyInGroup
	}
	g.Members = append(g.Members, userID)
	return nil
}
