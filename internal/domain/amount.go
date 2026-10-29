// Package domain defines the core business entities, value objects, and business invariants for FlowLedger.
package domain

import (
	"fmt"
	"strconv"
	"strings"
)

// AmountIDR represents a monetary value in whole Indonesian Rupiah (IDR).
// Floating-point types are strictly forbidden across FlowLedger to avoid rounding discrepancies.
type AmountIDR int64

// Int64 returns the raw integer value of the amount.
func (a AmountIDR) Int64() int64 {
	return int64(a)
}

// IsPositive returns true if the amount is strictly greater than zero.
func (a AmountIDR) IsPositive() bool {
	return a > 0
}

// IsZero returns true if the amount equals zero.
func (a AmountIDR) IsZero() bool {
	return a == 0
}

// Add returns the sum of this amount and another amount.
func (a AmountIDR) Add(other AmountIDR) AmountIDR {
	return a + other
}

// Sub returns the difference of this amount and another amount.
func (a AmountIDR) Sub(other AmountIDR) AmountIDR {
	return a - other
}

// String formats the AmountIDR as a standardized Indonesian Rupiah string (e.g., "Rp 50.000").
// Negative values are formatted with a leading minus sign (e.g., "-Rp 25.000").
func (a AmountIDR) String() string {
	raw := int64(a)
	isNegative := raw < 0
	if isNegative {
		raw = -raw
	}

	str := strconv.FormatInt(raw, 10)
	n := len(str)
	if n <= 3 {
		if isNegative {
			return fmt.Sprintf("-Rp %s", str)
		}
		return fmt.Sprintf("Rp %s", str)
	}

	var b strings.Builder
	remainder := n % 3
	if remainder > 0 {
		b.WriteString(str[:remainder])
		if n > remainder {
			b.WriteByte('.')
		}
	}

	for i := remainder; i < n; i += 3 {
		b.WriteString(str[i : i+3])
		if i+3 < n {
			b.WriteByte('.')
		}
	}

	if isNegative {
		return fmt.Sprintf("-Rp %s", b.String())
	}
	return fmt.Sprintf("Rp %s", b.String())
}

