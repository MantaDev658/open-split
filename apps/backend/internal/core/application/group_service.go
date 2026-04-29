package application

import (
	"context"
	"errors"
	"fmt"

	"opensplit/apps/backend/internal/core/domain"

	"github.com/google/uuid"
)

type GroupService struct {
	groupRepo   domain.GroupRepository
	expenseRepo domain.ExpenseRepository
}

func NewGroupService(groupRepo domain.GroupRepository, expenseRepo domain.ExpenseRepository) *GroupService {
	return &GroupService{
		groupRepo:   groupRepo,
		expenseRepo: expenseRepo,
	}
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

	if err := s.groupRepo.Save(ctx, group); err != nil {
		return "", fmt.Errorf("failed to save group: %w", err)
	}
	return string(id), nil
}

func (s *GroupService) ListGroupsForUser(ctx context.Context, userID string) ([]*domain.Group, error) {
	return s.groupRepo.ListForUser(ctx, domain.UserID(userID))
}

func (s *GroupService) AddMemberToGroup(ctx context.Context, groupID string, userID string) error {
	gID := domain.GroupID(groupID)
	uID := domain.UserID(userID)

	group, err := s.groupRepo.GetByID(ctx, gID)
	if err != nil {
		return fmt.Errorf("failed to fetch group: %w", err)
	}

	if err := group.AddMember(uID); err != nil {
		return err
	}

	if err := s.groupRepo.Save(ctx, group); err != nil {
		return fmt.Errorf("failed to save group member: %w", err)
	}

	return nil
}

func (s *GroupService) UpdateGroup(ctx context.Context, id string, name string) error {
	if name == "" {
		return errors.New("group name cannot be empty")
	}
	return s.groupRepo.UpdateName(ctx, domain.GroupID(id), name)
}

func (s *GroupService) DeleteGroup(ctx context.Context, id string) error {
	return s.groupRepo.Delete(ctx, domain.GroupID(id))
}

func (s *GroupService) RemoveMember(ctx context.Context, groupID string, userID string) error {
	gID := domain.GroupID(groupID)
	uID := domain.UserID(userID)

	expenses, err := s.expenseRepo.ListByGroup(ctx, gID)
	if err != nil {
		return fmt.Errorf("failed to fetch group expenses for validation: %w", err)
	}

	balances := domain.CalculateNetBalances(expenses)

	if balance, exists := balances[uID]; exists && balance != 0 {
		dollars := float64(balance) / 100.0
		return fmt.Errorf("cannot remove user: outstanding balance of $%.2f must be settled first", dollars)
	}

	return s.groupRepo.RemoveMember(ctx, gID, uID)
}
