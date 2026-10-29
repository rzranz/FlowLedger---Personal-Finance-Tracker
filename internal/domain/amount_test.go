package domain_test

import (
	"testing"

	"flowledger/internal/domain"
)

func TestAmountIDR_Operations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		amount     domain.AmountIDR
		other      domain.AmountIDR
		wantAdd    domain.AmountIDR
		wantSub    domain.AmountIDR
		wantPos    bool
		wantZero   bool
		wantInt64  int64
		wantString string
	}{
		{
			name:       "zero amount",
			amount:     0,
			other:      50000,
			wantAdd:    50000,
			wantSub:    -50000,
			wantPos:    false,
			wantZero:   true,
			wantInt64:  0,
			wantString: "Rp 0",
		},
		{
			name:       "small amount under one thousand",
			amount:     500,
			other:      200,
			wantAdd:    700,
			wantSub:    300,
			wantPos:    true,
			wantZero:   false,
			wantInt64:  500,
			wantString: "Rp 500",
		},
		{
			name:       "standard transaction amount",
			amount:     50000,
			other:      25000,
			wantAdd:    75000,
			wantSub:    25000,
			wantPos:    true,
			wantZero:   false,
			wantInt64:  50000,
			wantString: "Rp 50.000",
		},
		{
			name:       "large amount with millions",
			amount:     1250000,
			other:      250000,
			wantAdd:    1500000,
			wantSub:    1000000,
			wantPos:    true,
			wantZero:   false,
			wantInt64:  1250000,
			wantString: "Rp 1.250.000",
		},
		{
			name:       "negative amount",
			amount:     -25000,
			other:      10000,
			wantAdd:    -15000,
			wantSub:    -35000,
			wantPos:    false,
			wantZero:   false,
			wantInt64:  -25000,
			wantString: "-Rp 25.000",
		},
		{
			name:       "small negative amount under one thousand",
			amount:     -500,
			other:      100,
			wantAdd:    -400,
			wantSub:    -600,
			wantPos:    false,
			wantZero:   false,
			wantInt64:  -500,
			wantString: "-Rp 500",
		},
		{
			name:       "negative millions",
			amount:     -10000000,
			other:      5000000,
			wantAdd:    -5000000,
			wantSub:    -15000000,
			wantPos:    false,
			wantZero:   false,
			wantInt64:  -10000000,
			wantString: "-Rp 10.000.000",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.amount.Int64(); got != tt.wantInt64 {
				t.Errorf("Int64() = %d, want %d", got, tt.wantInt64)
			}
			if got := tt.amount.IsPositive(); got != tt.wantPos {
				t.Errorf("IsPositive() = %v, want %v", got, tt.wantPos)
			}
			if got := tt.amount.IsZero(); got != tt.wantZero {
				t.Errorf("IsZero() = %v, want %v", got, tt.wantZero)
			}
			if got := tt.amount.Add(tt.other); got != tt.wantAdd {
				t.Errorf("Add(%d) = %d, want %d", tt.other, got, tt.wantAdd)
			}
			if got := tt.amount.Sub(tt.other); got != tt.wantSub {
				t.Errorf("Sub(%d) = %d, want %d", tt.other, got, tt.wantSub)
			}
			if got := tt.amount.String(); got != tt.wantString {
				t.Errorf("String() = %q, want %q", got, tt.wantString)
			}
		})
	}
}
