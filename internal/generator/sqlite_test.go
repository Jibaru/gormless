package generator_test

import (
	"testing"

	"github.com/Jibaru/gormless/internal/generator"
)

func TestGenerateSQLiteDAO(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		evaluateGenerateDAO(t, "sqlite", generator.GenerateSQLiteDAO)
	})
}
