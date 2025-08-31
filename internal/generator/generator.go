package generator

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/Jibaru/gormless/internal/parser"
)

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
