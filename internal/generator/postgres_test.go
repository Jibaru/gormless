package generator_test

import (
	"testing"

	"github.com/Jibaru/gormless/internal/generator"
)

func TestGeneratePostgresDAO(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		evaluateGenerateDAO(t, "postgres", generator.GeneratePostgresDAO)
	})
}
