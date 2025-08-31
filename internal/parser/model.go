package parser

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

type Model struct {
	Name       string
	Fields     []Field
	TableName  string
	PrimaryKey string
	Package    string
	ImportPath string
}

type Field struct {
	Name      string
	Type      string
	Column    string
	IsPrimary bool
}

func ParseModels(inputPath string) ([]Model, error) {
	var models []Model

	info, err := os.Stat(inputPath)
	if err != nil {
		return nil, err
	}

	if info.IsDir() {
		err = filepath.Walk(inputPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "_test.go") {
				fileModels, err := parseFileModels(path)
				if err != nil {
					return err
				}
				models = append(models, fileModels...)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	} else if strings.HasSuffix(inputPath, ".go") {
		models, err = parseFileModels(inputPath)
		if err != nil {
			return nil, err
		}
	}

	return models, nil
}

func parseFileModels(filePath string) ([]Model, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	var models []Model
	packageName := node.Name.Name

	importPath, err := determineImportPath(filePath, packageName)
	if err != nil {
		return nil, err
	}

	ast.Inspect(node, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.GenDecl:
			if x.Tok == token.TYPE {
				for _, spec := range x.Specs {
					if ts, ok := spec.(*ast.TypeSpec); ok {
						if st, ok := ts.Type.(*ast.StructType); ok {
							model, err := parseStruct(ts.Name.Name, st, node)
							if err == nil && len(model.Fields) > 0 {
								model.Package = packageName
								model.ImportPath = importPath
								models = append(models, model)
							}
						}
					}
				}
			}
		}
		return true
	})

	return models, nil
}

func determineImportPath(filePath, packageName string) (string, error) {
	goModPath := findGoMod(filepath.Dir(filePath))
	if goModPath == "" {
		return "", fmt.Errorf("could not find go.mod file")
	}

	modContent, err := os.ReadFile(goModPath)
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(modContent), "\n")
	var moduleName string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			moduleName = strings.TrimSpace(strings.TrimPrefix(line, "module"))
			break
		}
	}

	if moduleName == "" {
		return "", fmt.Errorf("could not find module name in go.mod")
	}

	relPath, err := filepath.Rel(filepath.Dir(goModPath), filepath.Dir(filePath))
	if err != nil {
		return "", err
	}

	if relPath == "." {
		return moduleName, nil
	}

	return moduleName + "/" + strings.ReplaceAll(relPath, "\\", "/"), nil
}

func findGoMod(dir string) string {
	current := dir
	for {
		goModPath := filepath.Join(current, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			return goModPath
		}

		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}
	return ""
}

func parseStruct(name string, st *ast.StructType, file *ast.File) (Model, error) {
	model := Model{
		Name:      name,
		TableName: name,
		Fields:    []Field{},
	}

	tableName := getTableNameFromMethods(name, file)
	if tableName != "" {
		model.TableName = tableName
	}

	var primaryKeyFound bool

	for _, field := range st.Fields.List {
		if len(field.Names) == 0 {
			continue
		}

		fieldName := field.Names[0].Name
		if fieldName[0:1] != strings.ToUpper(fieldName[0:1]) {
			continue
		}

		fieldType := getTypeString(field.Type)
		column := fieldName
		isPrimary := false

		if field.Tag != nil {
			tag := strings.Trim(field.Tag.Value, "`")
			sqlTag := extractTag(tag, "sql")
			if sqlTag != "" {
				parts := strings.Split(sqlTag, ",")
				if len(parts) > 0 && parts[0] != "" {
					column = parts[0]
				}
				for _, part := range parts {
					if strings.TrimSpace(part) == "primary" {
						isPrimary = true
						primaryKeyFound = true
						model.PrimaryKey = fieldName
					}
				}
			}
		}

		model.Fields = append(model.Fields, Field{
			Name:      fieldName,
			Type:      fieldType,
			Column:    column,
			IsPrimary: isPrimary,
		})
	}

	if len(model.Fields) == 0 {
		return Model{}, fmt.Errorf("there are no exposed fields in the %s model", name)
	}

	if !primaryKeyFound {
		return Model{}, fmt.Errorf("there is no primary tag in the %s model", name)
	}

	return model, nil
}

func getTableNameFromMethods(structName string, file *ast.File) string {
	for _, decl := range file.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok {
			if fn.Recv != nil && len(fn.Recv.List) > 0 {
				if receiverType := getReceiverType(fn.Recv.List[0].Type); receiverType == structName {
					if fn.Name.Name == "TableName" {
						if fn.Body != nil {
							for _, stmt := range fn.Body.List {
								if ret, ok := stmt.(*ast.ReturnStmt); ok {
									if len(ret.Results) > 0 {
										if lit, ok := ret.Results[0].(*ast.BasicLit); ok && lit.Kind == token.STRING {
											return strings.Trim(lit.Value, `"`)
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}
	return ""
}

func getReceiverType(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.StarExpr:
		if ident, ok := t.X.(*ast.Ident); ok {
			return ident.Name
		}
	case *ast.Ident:
		return t.Name
	}
	return ""
}

func getTypeString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + getTypeString(t.X)
	case *ast.SelectorExpr:
		return getTypeString(t.X) + "." + t.Sel.Name
	case *ast.ArrayType:
		return "[]" + getTypeString(t.Elt)
	case *ast.MapType:
		return "map[" + getTypeString(t.Key) + "]" + getTypeString(t.Value)
	}
	return "interface{}"
}

func extractTag(tag, key string) string {
	st := reflect.StructTag(tag)
	return st.Get(key)
}
