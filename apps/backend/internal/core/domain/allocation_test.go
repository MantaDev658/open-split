package domain

import (
	"testing"
)

func TestAllocate_Equal(t *testing.T) {
	// $10.00 split 3 ways. Should be $3.34, $3.33, $3.33
	inputs := []AllocationInput{
		{UserID: "Alice"}, {UserID: "Bob"}, {UserID: "Charlie"},
	}

	splits, err := Allocate(AllocationTypeEqual, 1000, inputs)
	if err != nil {
		t.Fatal(err)
	}

	if splits[0].Amount.Int64() != 334 || splits[1].Amount.Int64() != 333 || splits[2].Amount.Int64() != 333 {
		t.Errorf("Equal penny rounding failed, got: %v", splits)
	}
}

func TestAllocate_Percentage(t *testing.T) {
	inputs := []AllocationInput{
		{UserID: "Alice", Value: 60.00},
		{UserID: "Bob", Value: 40.00},
	}
	splits, err := Allocate(AllocationTypePercentage, 1000, inputs)
	if err != nil {
		t.Fatal(err)
	}
	if splits[0].Amount.Int64() != 600 || splits[1].Amount.Int64() != 400 {
		t.Errorf("Percentage split failed")
	}
}
