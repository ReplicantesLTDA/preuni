package domain_test

import (
	"testing"

	"github.com/preuni/app/internal/auth/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidatePassword(t *testing.T) {
	t.Run("valid passwords", func(t *testing.T) {
		valid := []string{
			"Abc123",          // exactly 6 chars — minimum
			"Password1",       // typical strong password
			"MyP@ssw0rd",     // special chars + digits
		}
		for _, p := range valid {
			err := domain.ValidatePassword(p)
			assert.NoError(t, err, "expected %q to be valid", p)
		}
	})

	t.Run("too short — 5 characters", func(t *testing.T) {
		err := domain.ValidatePassword("Ab1cd")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "6 characters")
	})

	t.Run("too short — empty string", func(t *testing.T) {
		err := domain.ValidatePassword("")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "6 characters")
	})

	t.Run("too long — 256 characters", func(t *testing.T) {
		long := "Aa1" + repeat("x", 253) // 256 total
		err := domain.ValidatePassword(long)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "255 characters")
	})

	t.Run("exactly 255 characters — valid", func(t *testing.T) {
		p := "Aa1" + repeat("x", 252) // 255 total
		err := domain.ValidatePassword(p)
		assert.NoError(t, err)
	})

	t.Run("no uppercase letter", func(t *testing.T) {
		err := domain.ValidatePassword("abc123def")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "uppercase")
	})

	t.Run("no lowercase letter", func(t *testing.T) {
		err := domain.ValidatePassword("ABC123DEF")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "lowercase")
	})

	t.Run("no digit", func(t *testing.T) {
		err := domain.ValidatePassword("AbcDefGhi")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "digit")
	})
}

func repeat(s string, n int) string {
	result := make([]byte, n*len(s))
	for i := range n {
		copy(result[i*len(s):], s)
	}
	return string(result)
}
