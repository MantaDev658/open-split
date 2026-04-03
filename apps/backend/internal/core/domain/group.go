package domain

import (
	"errors"
)

var (
	ErrGroupNotFound      = errors.New("group not found")
	ErrUserNotInGroup     = errors.New("user is not a member of this group")
	ErrUserAlreadyInGroup = errors.New("user is already in this group")
)

type GroupID string

type Group struct {
	ID      GroupID
	Name    string
	Members []UserID
}

func NewGroup(id GroupID, name string, creator UserID) (*Group, error) {
	if name == "" {
		return nil, errors.New("group name cannot be empty")
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
