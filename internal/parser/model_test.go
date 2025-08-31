package parser_test

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/Jibaru/gormless/internal/parser"
)

func TestParseModels(t *testing.T) {
	t.Run("success with single file", func(t *testing.T) {
		// Create temporary file with test model
		tmpDir := t.TempDir()

		// Create a temporary go.mod file for import path determination
		goModContent := `module github.com/test/models
go 1.21
`
		err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goModContent), 0644)
		if err != nil {
			t.Fatalf("failed to create go.mod file: %v", err)
		}

		testFile := filepath.Join(tmpDir, "models.go")

		testContent := `package models

import "time"

type User struct {
	ID        int       ` + "`sql:\"id,primary\"`" + `
	Name      string    ` + "`sql:\"name\"`" + `
	Email     *string   ` + "`sql:\"email\"`" + `
	Age       int       ` + "`sql:\"age\"`" + `
	CreatedAt time.Time ` + "`sql:\"created_at\"`" + `
}

type Product struct {
	ID          string    ` + "`sql:\"id,primary\"`" + `
	Title       string    ` + "`sql:\"title\"`" + `
	Description *string   ` + "`sql:\"description\"`" + `
	Price       float64   ` + "`sql:\"price\"`" + `
}

func (p *Product) TableName() string {
	return "products"
}
`

		err = os.WriteFile(testFile, []byte(testContent), 0644)
		if err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}

		// Call ParseModels
		models, err := parser.ParseModels(testFile)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify results
		if len(models) != 2 {
			t.Fatalf("expected 2 models, got %d", len(models))
		}

		// Check first model (User)
		userModel := findModel(models, "User")
		if userModel == nil {
			t.Fatal("User model not found")
		}

		expectedUserFields := []parser.Field{
			{Name: "ID", Type: "int", Column: "id", IsPrimary: true},
			{Name: "Name", Type: "string", Column: "name", IsPrimary: false},
			{Name: "Email", Type: "*string", Column: "email", IsPrimary: false},
			{Name: "Age", Type: "int", Column: "age", IsPrimary: false},
			{Name: "CreatedAt", Type: "time.Time", Column: "created_at", IsPrimary: false},
		}

		if userModel.Name != "User" {
			t.Errorf("expected User name, got %s", userModel.Name)
		}
		if userModel.TableName != "User" {
			t.Errorf("expected User tableName, got %s", userModel.TableName)
		}
		if userModel.PrimaryKey != "ID" {
			t.Errorf("expected ID primaryKey, got %s", userModel.PrimaryKey)
		}
		if userModel.Package != "models" {
			t.Errorf("expected models package, got %s", userModel.Package)
		}
		if userModel.ImportPath != "github.com/test/models" {
			t.Errorf("expected github.com/test/models import path, got %s", userModel.ImportPath)
		}

		if !reflect.DeepEqual(userModel.Fields, expectedUserFields) {
			t.Errorf("User fields mismatch.\nExpected: %+v\nGot: %+v", expectedUserFields, userModel.Fields)
		}

		// Check second model (Product with custom TableName)
		productModel := findModel(models, "Product")
		if productModel == nil {
			t.Fatal("Product model not found")
		}

		if productModel.TableName != "products" {
			t.Errorf("expected products tableName (from TableName() method), got %s", productModel.TableName)
		}

		expectedProductFields := []parser.Field{
			{Name: "ID", Type: "string", Column: "id", IsPrimary: true},
			{Name: "Title", Type: "string", Column: "title", IsPrimary: false},
			{Name: "Description", Type: "*string", Column: "description", IsPrimary: false},
			{Name: "Price", Type: "float64", Column: "price", IsPrimary: false},
		}

		if !reflect.DeepEqual(productModel.Fields, expectedProductFields) {
			t.Errorf("Product fields mismatch.\nExpected: %+v\nGot: %+v", expectedProductFields, productModel.Fields)
		}
	})

	t.Run("success with directory", func(t *testing.T) {
		// Create temporary directory with multiple files
		tmpDir := t.TempDir()

		// Create a temporary go.mod file for import path determination
		goModContent := `module github.com/test/models
go 1.21
`
		err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goModContent), 0644)
		if err != nil {
			t.Fatalf("failed to create go.mod file: %v", err)
		}

		// Create first file
		file1 := filepath.Join(tmpDir, "user.go")
		file1Content := `package models

type User struct {
	ID   int    ` + "`sql:\"id,primary\"`" + `
	Name string ` + "`sql:\"name\"`" + `
}
`

		// Create second file
		file2 := filepath.Join(tmpDir, "product.go")
		file2Content := `package models

type Product struct {
	ID    string  ` + "`sql:\"id,primary\"`" + `
	Price float64 ` + "`sql:\"price\"`" + `
}
`

		// Create test file (should be ignored)
		testFile := filepath.Join(tmpDir, "user_test.go")
		testFileContent := `package models

type TestModel struct {
	ID int ` + "`sql:\"id,primary\"`" + `
}
`

		err = os.WriteFile(file1, []byte(file1Content), 0644)
		if err != nil {
			t.Fatalf("failed to create test file 1: %v", err)
		}

		err = os.WriteFile(file2, []byte(file2Content), 0644)
		if err != nil {
			t.Fatalf("failed to create test file 2: %v", err)
		}

		err = os.WriteFile(testFile, []byte(testFileContent), 0644)
		if err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}

		// Call ParseModels with directory
		models, err := parser.ParseModels(tmpDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify results (should only have 2 models, test file should be ignored)
		if len(models) != 2 {
			t.Fatalf("expected 2 models, got %d", len(models))
		}

		// Check that we have both User and Product models
		userModel := findModel(models, "User")
		productModel := findModel(models, "Product")

		if userModel == nil {
			t.Fatal("User model not found")
		}
		if productModel == nil {
			t.Fatal("Product model not found")
		}

		// Verify basic properties
		if userModel.Name != "User" {
			t.Errorf("expected User name, got %s", userModel.Name)
		}
		if productModel.Name != "Product" {
			t.Errorf("expected Product name, got %s", productModel.Name)
		}
	})

	t.Run("struct with various field types", func(t *testing.T) {
		// Test various Go field types
		tmpDir := t.TempDir()

		// Create a temporary go.mod file for import path determination
		goModContent := `module github.com/test/models
go 1.21
`
		err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goModContent), 0644)
		if err != nil {
			t.Fatalf("failed to create go.mod file: %v", err)
		}

		testFile := filepath.Join(tmpDir, "complex.go")

		testContent := `package models

import "time"

type ComplexModel struct {
	ID          int                    ` + "`sql:\"id,primary\"`" + `
	Name        string                 ` + "`sql:\"name\"`" + `
	OptionalStr *string                ` + "`sql:\"opt_str\"`" + `
	Numbers     []int                  ` + "`sql:\"numbers\"`" + `
	Mapping     map[string]interface{} ` + "`sql:\"mapping\"`" + `
	CreatedAt   time.Time              ` + "`sql:\"created_at\"`" + `
	private     string                 // should be ignored
}
`

		err = os.WriteFile(testFile, []byte(testContent), 0644)
		if err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}

		models, err := parser.ParseModels(testFile)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(models) != 1 {
			t.Fatalf("expected 1 model, got %d", len(models))
		}

		model := models[0]
		expectedFields := []parser.Field{
			{Name: "ID", Type: "int", Column: "id", IsPrimary: true},
			{Name: "Name", Type: "string", Column: "name", IsPrimary: false},
			{Name: "OptionalStr", Type: "*string", Column: "opt_str", IsPrimary: false},
			{Name: "Numbers", Type: "[]int", Column: "numbers", IsPrimary: false},
			{Name: "Mapping", Type: "map[string]interface{}", Column: "mapping", IsPrimary: false},
			{Name: "CreatedAt", Type: "time.Time", Column: "created_at", IsPrimary: false},
		}

		// Should not include private field
		if len(model.Fields) != 6 {
			t.Fatalf("expected 6 fields (private field should be ignored), got %d", len(model.Fields))
		}

		if !reflect.DeepEqual(model.Fields, expectedFields) {
			t.Errorf("Fields mismatch.\nExpected: %+v\nGot: %+v", expectedFields, model.Fields)
		}
	})

	t.Run("struct without sql tags uses field names as columns", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create a temporary go.mod file for import path determination
		goModContent := `module github.com/test/models
go 1.21
`
		err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goModContent), 0644)
		if err != nil {
			t.Fatalf("failed to create go.mod file: %v", err)
		}

		testFile := filepath.Join(tmpDir, "notags.go")

		testContent := `package models

type NoTagsModel struct {
	ID   int    ` + "`sql:\",primary\"`" + ` // empty column name, should use field name
	Name string // no tag, should use field name
}
`

		err = os.WriteFile(testFile, []byte(testContent), 0644)
		if err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}

		models, err := parser.ParseModels(testFile)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(models) != 1 {
			t.Fatalf("expected 1 model, got %d", len(models))
		}

		model := models[0]
		expectedFields := []parser.Field{
			{Name: "ID", Type: "int", Column: "ID", IsPrimary: true},
			{Name: "Name", Type: "string", Column: "Name", IsPrimary: false},
		}

		if !reflect.DeepEqual(model.Fields, expectedFields) {
			t.Errorf("Fields mismatch.\nExpected: %+v\nGot: %+v", expectedFields, model.Fields)
		}
	})

	t.Run("error cases", func(t *testing.T) {
		t.Run("non-existent file", func(t *testing.T) {
			_, err := parser.ParseModels("/non/existent/file.go")
			if err == nil {
				t.Fatal("expected error for non-existent file")
			}
		})

		t.Run("struct without primary key", func(t *testing.T) {
			tmpDir := t.TempDir()

			// Create a temporary go.mod file for import path determination
			goModContent := `module github.com/test/models
go 1.21
`
			err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goModContent), 0644)
			if err != nil {
				t.Fatalf("failed to create go.mod file: %v", err)
			}

			testFile := filepath.Join(tmpDir, "nopk.go")

			testContent := `package models

type NoPrimaryKey struct {
	Name string ` + "`sql:\"name\"`" + `
	Age  int    ` + "`sql:\"age\"`" + `
}
`

			err = os.WriteFile(testFile, []byte(testContent), 0644)
			if err != nil {
				t.Fatalf("failed to create test file: %v", err)
			}

			models, err := parser.ParseModels(testFile)
			// Should return error or empty models since there's no primary key
			if err == nil && len(models) > 0 {
				t.Fatal("expected error or no models for struct without primary key")
			}
		})

		t.Run("struct without fields", func(t *testing.T) {
			tmpDir := t.TempDir()

			// Create a temporary go.mod file for import path determination
			goModContent := `module github.com/test/models
go 1.21
`
			err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goModContent), 0644)
			if err != nil {
				t.Fatalf("failed to create go.mod file: %v", err)
			}

			testFile := filepath.Join(tmpDir, "empty.go")

			testContent := `package models

type EmptyStruct struct {
}

type OnlyPrivateFields struct {
	private string
}
`

			err = os.WriteFile(testFile, []byte(testContent), 0644)
			if err != nil {
				t.Fatalf("failed to create test file: %v", err)
			}

			models, err := parser.ParseModels(testFile)
			// Should return no models since structs have no valid fields
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(models) != 0 {
				t.Fatalf("expected 0 models for structs without valid fields, got %d", len(models))
			}
		})
	})

	t.Run("custom table name via method", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create a temporary go.mod file for import path determination
		goModContent := `module github.com/test/models
go 1.21
`
		err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goModContent), 0644)
		if err != nil {
			t.Fatalf("failed to create go.mod file: %v", err)
		}

		testFile := filepath.Join(tmpDir, "custom.go")

		testContent := `package models

type CustomTableModel struct {
	ID   int    ` + "`sql:\"id,primary\"`" + `
	Name string ` + "`sql:\"name\"`" + `
}

func (c *CustomTableModel) TableName() string {
	return "custom_table_name"
}
`

		err = os.WriteFile(testFile, []byte(testContent), 0644)
		if err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}

		models, err := parser.ParseModels(testFile)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(models) != 1 {
			t.Fatalf("expected 1 model, got %d", len(models))
		}

		model := models[0]
		if model.TableName != "custom_table_name" {
			t.Errorf("expected custom_table_name, got %s", model.TableName)
		}
	})
}

// Helper function to find a model by name
func findModel(models []parser.Model, name string) *parser.Model {
	for i := range models {
		if models[i].Name == name {
			return &models[i]
		}
	}
	return nil
}
