package generator_test

import (
	"bufio"
	"fmt"
	"os"
	"testing"

	"github.com/Jibaru/gormless/internal/generator"
	"github.com/Jibaru/gormless/internal/parser"
)

func TestGenerateDAOs(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		drivers := []string{
			"mysql",
			"postgres",
			"sqlserver",
			"oracle",
		}

		for _, driver := range drivers {
			t.Run("driver: "+driver, func(t *testing.T) {
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

				outputPath, err := os.MkdirTemp("", "test")
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				defer os.RemoveAll(outputPath)

				err = generator.GenerateDAOs([]parser.Model{model}, outputPath, driver)
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}

				compareFilesLineByLine(t, fmt.Sprintf("data/%s_formatted_stub.txt", driver), fmt.Sprintf("%s/%s/user_dao.go", outputPath, driver))
			})
		}
	})
}

func compareFilesLineByLine(t *testing.T, expectedPath, gotPath string) {
	t.Helper()

	expectedFile, err := os.Open(expectedPath)
	if err != nil {
		t.Fatalf("failed to open expected file %q: %v", expectedPath, err)
	}
	defer expectedFile.Close()

	gotFile, err := os.Open(gotPath)
	if err != nil {
		t.Fatalf("failed to open got file %q: %v", gotPath, err)
	}
	defer gotFile.Close()

	expectedScanner := bufio.NewScanner(expectedFile)
	gotScanner := bufio.NewScanner(gotFile)

	lineNum := 1
	for expectedScanner.Scan() {
		if !gotScanner.Scan() {
			t.Fatalf("file %q ended early at line %d, expected more lines", gotPath, lineNum)
		}

		expectedLine := expectedScanner.Text()
		gotLine := gotScanner.Text()

		if expectedLine != gotLine {
			t.Fatalf("mismatch at line %d:\nexpected: %q\ngot:      %q", lineNum, expectedLine, gotLine)
		}

		lineNum++
	}

	if err := expectedScanner.Err(); err != nil {
		t.Fatalf("error reading expected file: %v", err)
	}
	if err := gotScanner.Err(); err != nil {
		t.Fatalf("error reading got file: %v", err)
	}

	// If gotFile has more lines than expectedFile
	if gotScanner.Scan() {
		t.Fatalf("file %q has extra lines starting at line %d", gotPath, lineNum)
	}
}
