package generator

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Jibaru/gormless/internal/parser"
)

func GenerateDAOInterfaces(models []parser.Model, outputPath string) error {
	if err := os.MkdirAll(outputPath, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	for _, model := range models {
		fileName := fmt.Sprintf("%s_dao.go", toSnakeCase(model.Name))
		filePath := filepath.Join(outputPath, fileName)

		if _, err := os.Stat(filePath); err == nil {
			return fmt.Errorf("file with name %s already exists", filePath)
		}

		content, err := generateDAOInterface(model)
		if err != nil {
			return fmt.Errorf("failed to generate DAO interface for model %s: %v", model.Name, err)
		}

		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write DAO interface file for model %s: %v", model.Name, err)
		}

		if err := formatGoFile(filePath); err != nil {
			return fmt.Errorf("failed to format DAO interface file for model %s: %v", model.Name, err)
		}
	}

	return nil
}

func GenerateDAOs(models []parser.Model, outputPath, driver string) error {
	driverPath := filepath.Join(outputPath, driver)

	if err := os.MkdirAll(driverPath, 0755); err != nil {
		return fmt.Errorf("failed to create driver directory: %v", err)
	}

	for _, model := range models {
		fileName := fmt.Sprintf("%s_dao.go", toSnakeCase(model.Name))
		filePath := filepath.Join(driverPath, fileName)

		if _, err := os.Stat(filePath); err == nil {
			return fmt.Errorf("file with name %s already exists", filePath)
		}

		var content string
		var err error

		switch driver {
		case "postgres":
			content, err = GeneratePostgresDAO(model)
		case "mysql":
			content, err = GenerateMySQLDAO(model)
		case "sqlserver":
			content, err = GenerateSQLServerDAO(model)
		case "oracle":
			content, err = GenerateOracleDAO(model)
		case "sqlite":
			content, err = GenerateSQLiteDAO(model)
		default:
			return fmt.Errorf("unsupported driver: %s", driver)
		}

		if err != nil {
			return fmt.Errorf("failed to generate DAO for model %s: %v", model.Name, err)
		}

		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write DAO file for model %s: %v", model.Name, err)
		}

		if err := formatGoFile(filePath); err != nil {
			return fmt.Errorf("failed to format DAO file for model %s: %v", model.Name, err)
		}
	}

	return nil
}

func formatGoFile(filePath string) error {
	cmd := exec.Command("goimports", "-w", filePath)
	if err := cmd.Run(); err != nil {
		cmd = exec.Command("gofmt", "-w", filePath)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to format file with both goimports and gofmt: %v", err)
		}
	}
	return nil
}

func generateDAOInterface(model parser.Model) (string, error) {
	imports := []string{
		"context",
		model.ImportPath,
	}

	var content strings.Builder

	content.WriteString("package dao\n\n")
	content.WriteString("import (\n")
	for _, imp := range imports {
		content.WriteString(fmt.Sprintf("\t\"%s\"\n", imp))
	}
	content.WriteString(")\n\n")

	content.WriteString(fmt.Sprintf("type %s = %s.%s\n\n", model.Name, model.Package, model.Name))

	daoInterfaceName := fmt.Sprintf("%sDAO", model.Name)
	primaryType := getPrimaryType(model)

	content.WriteString(fmt.Sprintf("type %s interface {\n", daoInterfaceName))
	
	// CRUD operations
	content.WriteString(fmt.Sprintf("\t// Create creates a new %s\n", model.Name))
	content.WriteString(fmt.Sprintf("\tCreate(ctx context.Context, m *%s) error\n\n", model.Name))

	content.WriteString(fmt.Sprintf("\t// Update updates an existing %s\n", model.Name))
	content.WriteString(fmt.Sprintf("\tUpdate(ctx context.Context, m *%s) error\n\n", model.Name))

	content.WriteString(fmt.Sprintf("\t// PartialUpdate updates specific fields of a %s\n", model.Name))
	content.WriteString(fmt.Sprintf("\tPartialUpdate(ctx context.Context, pk %s, fields map[string]interface{}) error\n\n", primaryType))

	content.WriteString(fmt.Sprintf("\t// DeleteByPk deletes a %s by primary key\n", model.Name))
	content.WriteString(fmt.Sprintf("\tDeleteByPk(ctx context.Context, pk %s) error\n\n", primaryType))

	content.WriteString(fmt.Sprintf("\t// FindByPk finds a %s by primary key\n", model.Name))
	content.WriteString(fmt.Sprintf("\tFindByPk(ctx context.Context, pk %s) (*%s, error)\n\n", primaryType, model.Name))

	// Batch operations
	content.WriteString(fmt.Sprintf("\t// CreateMany creates multiple %s records\n", model.Name))
	content.WriteString(fmt.Sprintf("\tCreateMany(ctx context.Context, models []*%s) error\n\n", model.Name))

	content.WriteString(fmt.Sprintf("\t// UpdateMany updates multiple %s records\n", model.Name))
	content.WriteString(fmt.Sprintf("\tUpdateMany(ctx context.Context, models []*%s) error\n\n", model.Name))

	content.WriteString(fmt.Sprintf("\t// DeleteManyByPks deletes multiple %s records by primary keys\n", model.Name))
	content.WriteString(fmt.Sprintf("\tDeleteManyByPks(ctx context.Context, pks []%s) error\n\n", primaryType))

	// Query operations
	content.WriteString(fmt.Sprintf("\t// FindOne finds a single %s with optional where clause and sort expression\n", model.Name))
	content.WriteString(fmt.Sprintf("\tFindOne(ctx context.Context, where string, sort string, args ...interface{}) (*%s, error)\n\n", model.Name))

	content.WriteString(fmt.Sprintf("\t// FindAll finds all %s records with optional where clause and sort expression\n", model.Name))
	content.WriteString(fmt.Sprintf("\tFindAll(ctx context.Context, where string, sort string, args ...interface{}) ([]*%s, error)\n\n", model.Name))

	content.WriteString(fmt.Sprintf("\t// FindPaginated finds %s records with pagination, optional where clause and sort expression\n", model.Name))
	content.WriteString(fmt.Sprintf("\tFindPaginated(ctx context.Context, limit, offset int, where string, sort string, args ...interface{}) ([]*%s, error)\n\n", model.Name))

	content.WriteString(fmt.Sprintf("\t// Count counts %s records with optional where clause\n", model.Name))
	content.WriteString("\tCount(ctx context.Context, where string, args ...interface{}) (int64, error)\n\n")

	// Transaction support
	content.WriteString("\t// WithTransaction executes a function within a database transaction\n")
	content.WriteString("\tWithTransaction(ctx context.Context, fn func(ctx context.Context) error) error\n")

	content.WriteString("}\n")

	return content.String(), nil
}

func getPrimaryType(model parser.Model) string {
	for _, field := range model.Fields {
		if field.IsPrimary {
			return field.Type
		}
	}
	return "string"
}

func toSnakeCase(str string) string {
	var result []rune
	for i, r := range str {
		if i > 0 && 'A' <= r && r <= 'Z' {
			result = append(result, '_')
		}
		if 'A' <= r && r <= 'Z' {
			result = append(result, r+32)
		} else {
			result = append(result, r)
		}
	}
	return string(result)
}
