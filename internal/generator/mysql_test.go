package generator_test

import (
	"testing"

	"github.com/Jibaru/gormless/internal/generator"
)

func TestGenerateMySQLDAO(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		evaluateGenerateDAO(t, "mysql", generator.GenerateMySQLDAO)
	})
}
