package mocks

import (
	"context"
	"opensplit/apps/backend/internal/core/domain"
)

// Audit
type MockAuditRepo struct {
	SaveFunc        func(ctx context.Context, log domain.AuditLog) error
	ListByGroupFunc func(ctx context.Context, groupID domain.GroupID) ([]domain.AuditLog, error)
}

func (m *MockAuditRepo) Save(ctx context.Context, log domain.AuditLog) error {
	if m.SaveFunc != nil {
		return m.SaveFunc(ctx, log)
	}
	return nil
}

func (m *MockAuditRepo) ListByGroup(ctx context.Context, groupID domain.GroupID) ([]domain.AuditLog, error) {
	if m.ListByGroupFunc != nil {
		return m.ListByGroupFunc(ctx, groupID)
	}
	return nil, nil
}

// Expense
type MockExpenseRepo struct {
	SaveFunc                       func(ctx context.Context, expense *domain.Expense) error
	GetByIDFunc                    func(ctx context.Context, id domain.ExpenseID) (*domain.Expense, error)
	ListAllFunc                    func(ctx context.Context) ([]*domain.Expense, error)
	ListByGroupFunc                func(ctx context.Context, groupID domain.GroupID) ([]*domain.Expense, error)
	ListNonGroupExpensesByUserFunc func(ctx context.Context, userID domain.UserID) ([]*domain.Expense, error)
	UpdateFunc                     func(ctx context.Context, expense *domain.Expense) error
	DeleteFunc                     func(ctx context.Context, id domain.ExpenseID) error
}

func (m *MockExpenseRepo) Save(ctx context.Context, expense *domain.Expense) error {
	if m.SaveFunc != nil {
		return m.SaveFunc(ctx, expense)
	}
	return nil
}

func (m *MockExpenseRepo) GetByID(ctx context.Context, id domain.ExpenseID) (*domain.Expense, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return nil, domain.ErrExpenseNotFound
}

func (m *MockExpenseRepo) ListAll(ctx context.Context) ([]*domain.Expense, error) {
	if m.ListAllFunc != nil {
		return m.ListAllFunc(ctx)
	}
	return []*domain.Expense{}, nil
}

func (m *MockExpenseRepo) ListByGroup(ctx context.Context, groupID domain.GroupID) ([]*domain.Expense, error) {
	if m.ListByGroupFunc != nil {
		return m.ListByGroupFunc(ctx, groupID)
	}
	return nil, nil
}

func (m *MockExpenseRepo) ListNonGroupExpensesByUser(ctx context.Context, userID domain.UserID) ([]*domain.Expense, error) {
	if m.ListNonGroupExpensesByUserFunc != nil {
		return m.ListNonGroupExpensesByUserFunc(ctx, userID)
	}
	return []*domain.Expense{}, nil
}

func (m *MockExpenseRepo) Update(ctx context.Context, expense *domain.Expense) error {
	if m.ListByGroupFunc != nil {
		return m.UpdateFunc(ctx, expense)
	}
	return nil
}

func (m *MockExpenseRepo) Delete(ctx context.Context, id domain.ExpenseID) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

// User
type MockUserRepo struct {
	SaveFunc       func(ctx context.Context, user domain.User) error
	GetByIDFunc    func(ctx context.Context, id domain.UserID) (*domain.User, error)
	ListAllFunc    func(ctx context.Context) ([]domain.User, error)
	UpdateFunc     func(ctx context.Context, user domain.UserID, displayName string) error
	SoftDeleteFunc func(ctx context.Context, user domain.UserID) error
}

func (m *MockUserRepo) Save(ctx context.Context, user domain.User) error {
	if m.SaveFunc != nil {
		return m.SaveFunc(ctx, user)
	}
	return nil
}

func (m *MockUserRepo) GetByID(ctx context.Context, id domain.UserID) (*domain.User, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return &domain.User{ID: id, IsActive: true}, nil
}

func (m *MockUserRepo) ListAll(ctx context.Context) ([]domain.User, error) {
	if m.ListAllFunc != nil {
		return m.ListAllFunc(ctx)
	}
	return nil, nil
}

func (m *MockUserRepo) Update(ctx context.Context, user domain.UserID, displayName string) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, user, displayName)
	}
	return nil
}

func (m *MockUserRepo) SoftDelete(ctx context.Context, user domain.UserID) error {
	if m.SoftDeleteFunc != nil {
		return m.SoftDeleteFunc(ctx, user)
	}
	return nil
}

// Group
type MockGroupRepo struct {
	SaveFunc         func(ctx context.Context, group *domain.Group) error
	GetByIDFunc      func(ctx context.Context, id domain.GroupID) (*domain.Group, error)
	ListForUserFunc  func(ctx context.Context, userID domain.UserID) ([]*domain.Group, error)
	UpdateNameFunc   func(ctx context.Context, id domain.GroupID, name string) error
	DeleteFunc       func(ctx context.Context, id domain.GroupID) error
	RemoveMemberFunc func(ctx context.Context, id domain.GroupID, userID domain.UserID) error
}

func (m *MockGroupRepo) Save(ctx context.Context, group *domain.Group) error {
	if m.SaveFunc != nil {
		return m.SaveFunc(ctx, group)
	}
	return nil
}

func (m *MockGroupRepo) ListForUser(ctx context.Context, userID domain.UserID) ([]*domain.Group, error) {
	if m.ListForUserFunc != nil {
		return m.ListForUserFunc(ctx, userID)
	}
	return []*domain.Group{}, nil
}

func (m *MockGroupRepo) GetByID(ctx context.Context, id domain.GroupID) (*domain.Group, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return &domain.Group{}, nil
}

func (m *MockGroupRepo) UpdateName(ctx context.Context, id domain.GroupID, name string) error {
	if m.UpdateNameFunc != nil {
		return m.UpdateNameFunc(ctx, id, name)
	}
	return nil
}

func (m *MockGroupRepo) Delete(ctx context.Context, id domain.GroupID) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

func (m *MockGroupRepo) RemoveMember(ctx context.Context, id domain.GroupID, user domain.UserID) error {
	if m.RemoveMemberFunc != nil {
		return m.RemoveMemberFunc(ctx, id, user)
	}
	return nil
}
