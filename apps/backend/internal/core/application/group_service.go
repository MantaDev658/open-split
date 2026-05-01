package application

import (
	"context"
	"fmt"

	"opensplit/apps/backend/internal/core/domain"

	"github.com/google/uuid"
)

type GroupService struct {
	groupRepo   domain.GroupRepository
	expenseRepo domain.ExpenseRepository
	auditRepo   domain.AuditRepository
	transactor  domain.Transactor
}

func NewGroupService(groupRepo domain.GroupRepository, expenseRepo domain.ExpenseRepository, auditRepo domain.AuditRepository, tx domain.Transactor) *GroupService {
	return &GroupService{
		groupRepo:   groupRepo,
		expenseRepo: expenseRepo,
		auditRepo:   auditRepo,
		transactor:  tx,
	}
}

type CreateGroupCommand struct {
	Name      string `json:"name"`
	CreatorID string `json:"-"` // set by the handler from JWT; never read from client input
}

func (c CreateGroupCommand) Validate() error {
	if c.Name == "" {
		return domain.ErrEmptyGroupName
	}
	return nil
}

func (s *GroupService) CreateGroup(ctx context.Context, cmd CreateGroupCommand) (string, error) {
	id := domain.GroupID(uuid.NewString())
	group, err := domain.NewGroup(id, cmd.Name, domain.UserID(cmd.CreatorID))
	if err != nil {
		return "", err
	}

	err = s.transactor.RunInTx(ctx, func(txCtx context.Context) error {
		if saveErr := s.groupRepo.Save(txCtx, group); saveErr != nil {
			return fmt.Errorf("failed to save group: %w", saveErr)
		}
		return s.auditRepo.Save(txCtx, domain.AuditLog{
			ID:      uuid.NewString(),
			GroupID: string(group.ID),
			UserID:  cmd.CreatorID,
			Action:  "CREATED_GROUP",
			Details: "Created group: " + group.Name,
		})
	})
	if err != nil {
		return "", err
	}
	return string(id), nil
}

func (s *GroupService) ListGroupsForUser(ctx context.Context, userID string) ([]*domain.Group, error) {
	return s.groupRepo.ListForUser(ctx, domain.UserID(userID))
}

func (s *GroupService) AddMemberToGroup(ctx context.Context, groupID string, userID string, actorID string) error {
	gID := domain.GroupID(groupID)
	uID := domain.UserID(userID)

	group, err := s.groupRepo.GetByID(ctx, gID)
	if err != nil {
		return fmt.Errorf("failed to fetch group: %w", err)
	}

	if err := group.AddMember(uID); err != nil {
		return err
	}

	return s.transactor.RunInTx(ctx, func(txCtx context.Context) error {
		if err := s.groupRepo.Save(txCtx, group); err != nil {
			return fmt.Errorf("failed to save group member: %w", err)
		}
		return s.auditRepo.Save(txCtx, domain.AuditLog{
			ID:       uuid.NewString(),
			GroupID:  groupID,
			UserID:   actorID,
			Action:   "ADDED_MEMBER",
			TargetID: userID,
		})
	})
}

func (s *GroupService) UpdateGroup(ctx context.Context, groupID string, name string, actorID string) error {
	if name == "" {
		return domain.ErrEmptyGroupName
	}

	return s.transactor.RunInTx(ctx, func(txCtx context.Context) error {
		if err := s.groupRepo.UpdateName(txCtx, domain.GroupID(groupID), name); err != nil {
			return fmt.Errorf("failed to update group name: %w", err)
		}
		return s.auditRepo.Save(txCtx, domain.AuditLog{
			ID:      uuid.NewString(),
			GroupID: groupID,
			UserID:  actorID,
			Action:  "RENAMED_GROUP",
			Details: "Renamed to " + name,
		})
	})
}

func (s *GroupService) DeleteGroup(ctx context.Context, groupID string, userID string) error {
	return s.transactor.RunInTx(ctx, func(txCtx context.Context) error {
		if err := s.groupRepo.Delete(txCtx, domain.GroupID(groupID)); err != nil {
			return fmt.Errorf("failed to delete group: %w", err)
		}
		return s.auditRepo.Save(txCtx, domain.AuditLog{
			ID:       uuid.NewString(),
			GroupID:  groupID,
			UserID:   userID,
			Action:   "DELETED_GROUP",
			TargetID: groupID,
		})
	})
}

func (s *GroupService) RemoveMember(ctx context.Context, groupID string, userID string, actorID string) error {
	gID := domain.GroupID(groupID)
	uID := domain.UserID(userID)

	expenses, err := s.expenseRepo.ListByGroup(ctx, gID, domain.Page{})
	if err != nil {
		return fmt.Errorf("failed to fetch group expenses for validation: %w", err)
	}

	balances := domain.CalculateNetBalances(expenses)
	if balance, exists := balances[uID]; exists && balance != 0 {
		dollars := float64(balance) / 100.0
		return fmt.Errorf("%w: $%.2f", domain.ErrOutstandingBalance, dollars)
	}

	return s.transactor.RunInTx(ctx, func(txCtx context.Context) error {
		if err := s.groupRepo.RemoveMember(txCtx, gID, uID); err != nil {
			return fmt.Errorf("failed to remove group member: %w", err)
		}
		return s.auditRepo.Save(txCtx, domain.AuditLog{
			ID:       uuid.NewString(),
			GroupID:  groupID,
			UserID:   actorID,
			Action:   "REMOVED_GROUP_MEMBER",
			TargetID: userID,
		})
	})
}
