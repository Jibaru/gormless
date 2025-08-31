package generator_test

import (
	"os"
	"testing"

	"github.com/Jibaru/gormless/internal/generator"
	"github.com/Jibaru/gormless/internal/parser"
)

func TestGenerateSQLServerDAO(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		model := parser.Model{
			Name: "User",
			Fields: []parser.Field{
				{Name: "ID", Type: "int", Column: "id", IsPrimary: true},
				{Name: "Name", Type: "string", Column: "name"},
				{Name: "Email", Type: "*string", Column: "email"},
				{Name: "Password", Type: "string", Column: "password"},
				{Name: "Age", Type: "int", Column: "age"},
				{Name: "DeletedAt", Type: "*time.Time", Column: "deleted_at"},
			},
			TableName:  "users",
			PrimaryKey: "ID",
			Package:    "models",
			ImportPath: "github.com/someone/models",
		}

		content, err := generator.GenerateSQLServerDAO(model)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		expectedContent, err := os.ReadFile("data/sqlserver_stub.txt")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if string(expectedContent) != content {
			t.Errorf("expected %s, got %s", string(expectedContent), content)
		}
	})
}
