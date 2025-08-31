package generator_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/Jibaru/gormless/internal/parser"
)

func evaluateGenerateDAO(t *testing.T, driver string, genFunc func(parser.Model) (string, error)) {
	t.Helper()

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

	content, err := genFunc(model)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	expectedContent, err := os.ReadFile(fmt.Sprintf("data/%s_stub.txt", driver))
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if string(expectedContent) != content {
		t.Errorf("expected %s, got %s", string(expectedContent), content)
	}
}
