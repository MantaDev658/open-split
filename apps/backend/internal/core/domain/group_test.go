package domain

import (
	"errors"
	"testing"
)

func TestNewGroup(t *testing.T) {
	t.Run("Successfully creates group with creator as first member", func(t *testing.T) {
		group, err := NewGroup("g1", "Ski Trip", "Alice")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if group.Name != "Ski Trip" {
			t.Errorf("expected name 'Ski Trip', got %s", group.Name)
		}
		if len(group.Members) != 1 || group.Members[0] != "Alice" {
			t.Errorf("expected creator 'Alice' to be the only member")
		}
	})

	t.Run("Fails with empty name", func(t *testing.T) {
		_, err := NewGroup("g2", "", "Alice")
		if err == nil {
			t.Errorf("expected error for empty name, got nil")
		}
	})
}

func TestGroup_Membership(t *testing.T) {
	group, _ := NewGroup("g1", "Apartment", "Alice")

	t.Run("HasMember works correctly", func(t *testing.T) {
		if !group.HasMember("Alice") {
			t.Errorf("expected group to have member 'Alice'")
		}
		if group.HasMember("Bob") {
			t.Errorf("expected group to NOT have member 'Bob'")
		}
	})

	t.Run("AddMember prevents duplicates", func(t *testing.T) {
		err := group.AddMember("Bob")
		if err != nil {
			t.Fatalf("expected to add Bob successfully, got %v", err)
		}

		err = group.AddMember("Bob")
		if !errors.Is(err, ErrUserAlreadyInGroup) {
			t.Errorf("expected ErrUserAlreadyInGroup, got %v", err)
		}

		if len(group.Members) != 2 {
			t.Errorf("expected exactly 2 members, got %d", len(group.Members))
		}
	})
}
