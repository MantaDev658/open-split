package money

import (
	"errors"
)

// Money represents a monetary amount in the lowest unit (e.g., cents).
// This prevents floating-point math errors.
type Money int64

var (
	ErrNegativeAmount = errors.New("monetary amount cannot be negative")
)

// New creates a new Money instance from cents.
func New(cents int64) (Money, error) {
	if cents < 0 {
		return 0, ErrNegativeAmount
	}
	return Money(cents), nil
}

// Add sums two Money amounts.
func (m Money) Add(other Money) Money {
	return m + other
}

// Sub subtracts another Money amount.
func (m Money) Sub(other Money) Money {
	return m - other
}

// Distribute splits the Money into N parts.
// It handles the "remainder penny" by distributing it to the first few parts.
// Example: $0.10 split 3 ways results in [4, 3, 3] cents.
func (m Money) Distribute(n int) []Money {
	if n <= 0 {
		return nil
	}

	lowResult := m / Money(n)
	remainder := int(m % Money(n))

	results := make([]Money, n)
	for i := 0; i < n; i++ {
		results[i] = lowResult
		if i < remainder {
			results[i]++
		}
	}

	return results
}

// Int64 returns the raw value (cents).
func (m Money) Int64() int64 {
	return int64(m)
}
