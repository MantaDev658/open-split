package application

import (
	"context"
	"fmt"

	"opensplit/apps/backend/internal/core/domain"

	"github.com/google/uuid"
)

type GroupService struct {
	repo domain.GroupRepository
}

func NewGroupService(repo domain.GroupRepository) *GroupService {
	return &GroupService{repo: repo}
}

type CreateGroupCommand struct {
	Name    string `json:"name"`
	Creator string `json:"creator"`
}

func (s *GroupService) CreateGroup(ctx context.Context, cmd CreateGroupCommand) (string, error) {
	id := domain.GroupID(uuid.NewString())
	group, err := domain.NewGroup(id, cmd.Name, domain.UserID(cmd.Creator))
	if err != nil {
		return "", err
	}

	if err := s.repo.Save(ctx, group); err != nil {
		return "", fmt.Errorf("failed to save group: %w", err)
	}
	return string(id), nil
}

func (s *GroupService) ListGroupsForUser(ctx context.Context, userID string) ([]*domain.Group, error) {
	return s.repo.ListForUser(ctx, domain.UserID(userID))
}

func (s *GroupService) AddMemberToGroup(ctx context.Context, groupID string, userID string) error {
	gID := domain.GroupID(groupID)
	uID := domain.UserID(userID)

	group, err := s.repo.GetByID(ctx, gID)
	if err != nil {
		return fmt.Errorf("failed to fetch group: %w", err)
	}

	if err := group.AddMember(uID); err != nil {
		return err
	}

	if err := s.repo.Save(ctx, group); err != nil {
		return fmt.Errorf("failed to save group member: %w", err)
	}

	return nil
}
