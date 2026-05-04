package application

import (
	"context"
	"fmt"

	"opensplit/apps/backend/internal/core/domain"

	"github.com/google/uuid"
)

// GroupService implements group management and member lifecycle use cases.
type GroupService struct {
	groupRepo   domain.GroupRepository
	expenseRepo domain.ExpenseRepository
	auditRepo   domain.AuditRepository
	transactor  domain.Transactor
}

// NewGroupService wires the repositories and transactor into a GroupService.
func NewGroupService(groupRepo domain.GroupRepository, expenseRepo domain.ExpenseRepository, auditRepo domain.AuditRepository, tx domain.Transactor) *GroupService {
	return &GroupService{
		groupRepo:   groupRepo,
		expenseRepo: expenseRepo,
		auditRepo:   auditRepo,
		transactor:  tx,
	}
}

// CreateGroupCommand carries the input for creating a new group.
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

// CreateGroup validates the command, persists the group, and writes an audit log entry.
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
			Action:  domain.AuditActionCreatedGroup,
			Details: "Created group: " + group.Name,
		})
	})
	if err != nil {
		return "", err
	}
	return string(id), nil
}

// ListGroupsForUser returns all groups the given user belongs to.
func (s *GroupService) ListGroupsForUser(ctx context.Context, userID string) ([]*domain.Group, error) {
	return s.groupRepo.ListForUser(ctx, domain.UserID(userID))
}

// AddMemberToGroup adds userID to the group and records the change in the audit log.
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
			Action:   domain.AuditActionAddedMember,
			TargetID: userID,
		})
	})
}

// UpdateGroup renames a group and records the change in the audit log.
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
			Action:  domain.AuditActionRenamedGroup,
			Details: "Renamed to " + name,
		})
	})
}

// DeleteGroup removes the group and records the deletion in the audit log.
func (s *GroupService) DeleteGroup(ctx context.Context, groupID string, userID string) error {
	return s.transactor.RunInTx(ctx, func(txCtx context.Context) error {
		if err := s.groupRepo.Delete(txCtx, domain.GroupID(groupID)); err != nil {
			return fmt.Errorf("failed to delete group: %w", err)
		}
		return s.auditRepo.Save(txCtx, domain.AuditLog{
			ID:       uuid.NewString(),
			GroupID:  groupID,
			UserID:   userID,
			Action:   domain.AuditActionDeletedGroup,
			TargetID: groupID,
		})
	})
}

// RemoveMember removes userID from the group, returning ErrOutstandingBalance if they still owe or are owed money.
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
			Action:   domain.AuditActionRemovedMember,
			TargetID: userID,
		})
	})
}
