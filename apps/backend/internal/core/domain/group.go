package domain

type GroupID string

type Group struct {
	ID      GroupID
	Name    string
	Members []UserID
}

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

func (g *Group) HasMember(userID UserID) bool {
	for _, member := range g.Members {
		if member == userID {
			return true
		}
	}
	return false
}

func (g *Group) AddMember(userID UserID) error {
	if g.HasMember(userID) {
		return ErrUserAlreadyInGroup
	}
	g.Members = append(g.Members, userID)
	return nil
}
