package generator

import (
	"fmt"
	"strings"

	"github.com/Jibaru/gormless/internal/parser"
)

func GenerateSQLiteDAO(model parser.Model) (string, error) {
	imports := []string{
		"context",
		"database/sql",
		"fmt",
		"strings",
		model.ImportPath,
	}

	var content strings.Builder

	content.WriteString(fmt.Sprintf("package %s\n\n", "sqlite"))
	content.WriteString("import (\n")
	for _, imp := range imports {
		content.WriteString(fmt.Sprintf("\t\"%s\"\n", imp))
	}
	content.WriteString(")\n\n")

	content.WriteString(fmt.Sprintf("type %s = %s.%s\n\n", model.Name, model.Package, model.Name))

	daoName := fmt.Sprintf("%sDAO", model.Name)

	content.WriteString(fmt.Sprintf("type %s struct {\n", daoName))
	content.WriteString("\tdb *sql.DB\n")
	content.WriteString("}\n\n")

	content.WriteString(fmt.Sprintf("func New%s(db *sql.DB) *%s {\n", daoName, daoName))
	content.WriteString(fmt.Sprintf("\treturn &%s{db: db}\n", daoName))
	content.WriteString("}\n\n")

	content.WriteString(generateSQLiteHelperMethods(daoName))
	content.WriteString(generateSQLiteCreateMethod(model, daoName))
	content.WriteString(generateSQLiteUpdateMethod(model, daoName))
	content.WriteString(generateSQLitePartialUpdateMethod(model, daoName))
	content.WriteString(generateSQLiteDeleteByIDMethod(model, daoName))
	content.WriteString(generateSQLiteFindByIDMethod(model, daoName))
	content.WriteString(generateSQLiteCreateManyMethod(model, daoName))
	content.WriteString(generateSQLiteUpdateManyMethod(model, daoName))
	content.WriteString(generateSQLiteDeleteManyByIDsMethod(model, daoName))
	content.WriteString(generateSQLiteFindOneMethod(model, daoName))
	content.WriteString(generateSQLiteFindAllMethod(model, daoName))
	content.WriteString(generateSQLiteFindPaginatedMethod(model, daoName))
	content.WriteString(generateSQLiteCountMethod(model, daoName))
	content.WriteString(generateSQLiteWithTransactionMethod(daoName))

	return content.String(), nil
}

func generateSQLiteHelperMethods(daoName string) string {
	var content strings.Builder

	content.WriteString(fmt.Sprintf("func (dao *%s) getTx(ctx context.Context) *sql.Tx {\n", daoName))
	content.WriteString("\tif tx, ok := ctx.Value(\"currentTx\").(*sql.Tx); ok {\n")
	content.WriteString("\t\treturn tx\n")
	content.WriteString("\t}\n")
	content.WriteString("\treturn nil\n")
	content.WriteString("}\n\n")

	content.WriteString(fmt.Sprintf("func (dao *%s) execContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {\n", daoName))
	content.WriteString("\tif tx := dao.getTx(ctx); tx != nil {\n")
	content.WriteString("\t\treturn tx.ExecContext(ctx, query, args...)\n")
	content.WriteString("\t}\n")
	content.WriteString("\treturn dao.db.ExecContext(ctx, query, args...)\n")
	content.WriteString("}\n\n")

	content.WriteString(fmt.Sprintf("func (dao *%s) queryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {\n", daoName))
	content.WriteString("\tif tx := dao.getTx(ctx); tx != nil {\n")
	content.WriteString("\t\treturn tx.QueryRowContext(ctx, query, args...)\n")
	content.WriteString("\t}\n")
	content.WriteString("\treturn dao.db.QueryRowContext(ctx, query, args...)\n")
	content.WriteString("}\n\n")

	content.WriteString(fmt.Sprintf("func (dao *%s) queryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {\n", daoName))
	content.WriteString("\tif tx := dao.getTx(ctx); tx != nil {\n")
	content.WriteString("\t\treturn tx.QueryContext(ctx, query, args...)\n")
	content.WriteString("\t}\n")
	content.WriteString("\treturn dao.db.QueryContext(ctx, query, args...)\n")
	content.WriteString("}\n\n")

	return content.String()
}

func generateSQLiteCreateMethod(model parser.Model, daoName string) string {
	var content strings.Builder
	var columns []string
	var placeholders []string
	var args []string

	for _, field := range model.Fields {
		columns = append(columns, field.Column)
		placeholders = append(placeholders, "?")
		args = append(args, fmt.Sprintf("m.%s", field.Name))
	}

	content.WriteString(fmt.Sprintf("func (dao *%s) Create(ctx context.Context, m *%s) error {\n", daoName, model.Name))
	content.WriteString("\tquery := `\n")
	content.WriteString(fmt.Sprintf("\t\tINSERT INTO %s (%s)\n", model.TableName, strings.Join(columns, ", ")))
	content.WriteString(fmt.Sprintf("\t\tVALUES (%s)\n", strings.Join(placeholders, ", ")))
	content.WriteString("\t`\n\n")

	content.WriteString("\t_, err := dao.execContext(\n")
	content.WriteString("\t\tctx,\n")
	content.WriteString("\t\tquery,\n")
	for _, arg := range args {
		content.WriteString(fmt.Sprintf("\t\t%s,\n", arg))
	}
	content.WriteString("\t)\n\n")
	content.WriteString("\treturn err\n")
	content.WriteString("}\n\n")

	return content.String()
}

func generateSQLiteUpdateMethod(model parser.Model, daoName string) string {
	var content strings.Builder
	var setClauses []string
	var args []string
	var primaryKeyField string

	for _, field := range model.Fields {
		if field.IsPrimary {
			primaryKeyField = field.Name
			continue
		}
		setClauses = append(setClauses, fmt.Sprintf("%s = ?", field.Column))
		args = append(args, fmt.Sprintf("m.%s", field.Name))
	}

	args = append(args, fmt.Sprintf("m.%s", primaryKeyField))
	whereClause := fmt.Sprintf("WHERE %s = ?", getPrimaryColumn(model))

	content.WriteString(fmt.Sprintf("func (dao *%s) Update(ctx context.Context, m *%s) error {\n", daoName, model.Name))
	content.WriteString("\tquery := `\n")
	content.WriteString(fmt.Sprintf("\t\tUPDATE %s\n", model.TableName))
	content.WriteString(fmt.Sprintf("\t\tSET %s\n", strings.Join(setClauses, ",\n\t\t\t")))
	content.WriteString(fmt.Sprintf("\t\t%s\n", whereClause))
	content.WriteString("\t`\n\n")

	content.WriteString("\t_, err := dao.execContext(ctx, query,\n")
	for _, arg := range args {
		content.WriteString(fmt.Sprintf("\t\t%s,\n", arg))
	}
	content.WriteString("\t)\n")
	content.WriteString("\treturn err\n")
	content.WriteString("}\n\n")

	return content.String()
}

func generateSQLitePartialUpdateMethod(model parser.Model, daoName string) string {
	var content strings.Builder
	primaryColumn := getPrimaryColumn(model)
	primaryType := getPrimaryType(model)

	content.WriteString(fmt.Sprintf("func (dao *%s) PartialUpdate(ctx context.Context, pk %s, fields map[string]interface{}) error {\n", daoName, primaryType))
	content.WriteString("\tif len(fields) == 0 {\n")
	content.WriteString("\t\treturn nil\n")
	content.WriteString("\t}\n\n")

	content.WriteString("\tsetClauses := make([]string, 0, len(fields))\n")
	content.WriteString("\targs := make([]interface{}, 0, len(fields)+1)\n\n")

	content.WriteString("\tfor field, value := range fields {\n")
	content.WriteString("\t\tsetClauses = append(setClauses, field + \" = ?\")\n")
	content.WriteString("\t\targs = append(args, value)\n")
	content.WriteString("\t}\n\n")

	content.WriteString("\targs = append(args, pk)\n\n")

	content.WriteString(fmt.Sprintf("\tquery := fmt.Sprintf(\"UPDATE %s SET %%s WHERE %s = ?\", strings.Join(setClauses, \", \"))\n\n", model.TableName, primaryColumn))

	content.WriteString("\t_, err := dao.execContext(ctx, query, args...)\n")
	content.WriteString("\treturn err\n")
	content.WriteString("}\n\n")

	return content.String()
}

func generateSQLiteDeleteByIDMethod(model parser.Model, daoName string) string {
	var content strings.Builder
	primaryColumn := getPrimaryColumn(model)
	primaryType := getPrimaryType(model)

	content.WriteString(fmt.Sprintf("func (dao *%s) DeleteByPk(ctx context.Context, pk %s) error {\n", daoName, primaryType))
	content.WriteString(fmt.Sprintf("\tquery := `DELETE FROM %s WHERE %s = ?`\n", model.TableName, primaryColumn))
	content.WriteString("\t_, err := dao.execContext(ctx, query, pk)\n")
	content.WriteString("\treturn err\n")
	content.WriteString("}\n\n")

	return content.String()
}

func generateSQLiteFindByIDMethod(model parser.Model, daoName string) string {
	var content strings.Builder
	var columns []string
	var scanArgs []string
	primaryColumn := getPrimaryColumn(model)
	primaryType := getPrimaryType(model)

	for _, field := range model.Fields {
		columns = append(columns, field.Column)
		scanArgs = append(scanArgs, fmt.Sprintf("&m.%s", field.Name))
	}

	content.WriteString(fmt.Sprintf("func (dao *%s) FindByPk(ctx context.Context, pk %s) (*%s, error) {\n", daoName, primaryType, model.Name))
	content.WriteString("\tquery := `\n")
	content.WriteString(fmt.Sprintf("\t\tSELECT %s\n", strings.Join(columns, ", ")))
	content.WriteString(fmt.Sprintf("\t\tFROM %s\n", model.TableName))
	content.WriteString(fmt.Sprintf("\t\tWHERE %s = ?\n", primaryColumn))
	content.WriteString("\t`\n")
	content.WriteString("\trow := dao.queryRowContext(ctx, query, pk)\n\n")

	content.WriteString(fmt.Sprintf("\tvar m %s\n", model.Name))
	content.WriteString("\terr := row.Scan(\n")
	for _, arg := range scanArgs {
		content.WriteString(fmt.Sprintf("\t\t%s,\n", arg))
	}
	content.WriteString("\t)\n\n")

	content.WriteString("\tif err != nil {\n")
	content.WriteString("\t\treturn nil, err\n")
	content.WriteString("\t}\n\n")

	content.WriteString("\treturn &m, nil\n")
	content.WriteString("}\n\n")

	return content.String()
}

func generateSQLiteCreateManyMethod(model parser.Model, daoName string) string {
	var content strings.Builder
	var columns []string

	for _, field := range model.Fields {
		columns = append(columns, field.Column)
	}

	fieldCount := len(model.Fields)
	placeholders := strings.Repeat("?,", fieldCount-1) + "?"

	content.WriteString(fmt.Sprintf("func (dao *%s) CreateMany(ctx context.Context, models []*%s) error {\n", daoName, model.Name))
	content.WriteString("\tif len(models) == 0 {\n")
	content.WriteString("\t\treturn nil\n")
	content.WriteString("\t}\n\n")

	content.WriteString("\tplaceholders := make([]string, len(models))\n")
	content.WriteString(fmt.Sprintf("\targs := make([]interface{}, 0, len(models)*%d)\n\n", fieldCount))

	content.WriteString("\tfor i, model := range models {\n")
	content.WriteString(fmt.Sprintf("\t\tplaceholders[i] = \"(%s)\"\n\n", placeholders))

	content.WriteString("\t\targs = append(args,\n")
	for _, field := range model.Fields {
		content.WriteString(fmt.Sprintf("\t\t\tmodel.%s,\n", field.Name))
	}
	content.WriteString("\t\t)\n")
	content.WriteString("\t}\n\n")

	content.WriteString("\tquery := fmt.Sprintf(`\n")
	content.WriteString(fmt.Sprintf("\t\tINSERT INTO %s (%s)\n", model.TableName, strings.Join(columns, ", ")))
	content.WriteString("\t\tVALUES %s\n")
	content.WriteString("\t`, strings.Join(placeholders, \", \"))\n\n")

	content.WriteString("\t_, err := dao.execContext(ctx, query, args...)\n")
	content.WriteString("\treturn err\n")
	content.WriteString("}\n\n")

	return content.String()
}

func generateSQLiteUpdateManyMethod(model parser.Model, daoName string) string {
	var content strings.Builder
	var setClauses []string
	var args []string
	var primaryKeyField string

	for _, field := range model.Fields {
		if field.IsPrimary {
			primaryKeyField = field.Name
			continue
		}
		setClauses = append(setClauses, fmt.Sprintf("%s = ?", field.Column))
		args = append(args, fmt.Sprintf("model.%s", field.Name))
	}

	args = append(args, fmt.Sprintf("model.%s", primaryKeyField))
	whereClause := fmt.Sprintf("WHERE %s = ?", getPrimaryColumn(model))

	content.WriteString(fmt.Sprintf("func (dao *%s) UpdateMany(ctx context.Context, models []*%s) error {\n", daoName, model.Name))
	content.WriteString("\tif len(models) == 0 {\n")
	content.WriteString("\t\treturn nil\n")
	content.WriteString("\t}\n\n")

	content.WriteString("\tquery := `\n")
	content.WriteString(fmt.Sprintf("\t\tUPDATE %s\n", model.TableName))
	content.WriteString(fmt.Sprintf("\t\tSET %s\n", strings.Join(setClauses, ",\n\t\t\t")))
	content.WriteString(fmt.Sprintf("\t\t%s\n", whereClause))
	content.WriteString("\t`\n\n")

	content.WriteString("\tfor _, model := range models {\n")
	content.WriteString("\t\t_, err := dao.execContext(ctx, query,\n")
	for _, arg := range args {
		content.WriteString(fmt.Sprintf("\t\t\t%s,\n", arg))
	}
	content.WriteString("\t\t)\n")
	content.WriteString("\t\tif err != nil {\n")
	content.WriteString("\t\t\treturn err\n")
	content.WriteString("\t\t}\n")
	content.WriteString("\t}\n\n")

	content.WriteString("\treturn nil\n")
	content.WriteString("}\n\n")

	return content.String()
}

func generateSQLiteDeleteManyByIDsMethod(model parser.Model, daoName string) string {
	var content strings.Builder
	primaryColumn := getPrimaryColumn(model)
	primaryType := getPrimaryType(model)

	content.WriteString(fmt.Sprintf("func (dao *%s) DeleteManyByPks(ctx context.Context, pks []%s) error {\n", daoName, primaryType))
	content.WriteString("\tif len(pks) == 0 {\n")
	content.WriteString("\t\treturn nil\n")
	content.WriteString("\t}\n\n")

	content.WriteString("\tplaceholders := strings.Repeat(\"?,\", len(pks)-1) + \"?\"\n")
	content.WriteString("\targs := make([]interface{}, len(pks))\n")
	content.WriteString("\tfor i, pk := range pks {\n")
	content.WriteString("\t\targs[i] = pk\n")
	content.WriteString("\t}\n\n")

	content.WriteString(fmt.Sprintf("\tquery := fmt.Sprintf(\"DELETE FROM %s WHERE %s IN (%%s)\", placeholders)\n", model.TableName, primaryColumn))
	content.WriteString("\t_, err := dao.execContext(ctx, query, args...)\n")
	content.WriteString("\treturn err\n")
	content.WriteString("}\n\n")

	return content.String()
}

func generateSQLiteFindOneMethod(model parser.Model, daoName string) string {
	var content strings.Builder
	var columns []string
	var scanArgs []string

	for _, field := range model.Fields {
		columns = append(columns, field.Column)
		scanArgs = append(scanArgs, fmt.Sprintf("&m.%s", field.Name))
	}

	content.WriteString(fmt.Sprintf("func (dao *%s) FindOne(ctx context.Context, where string, args ...interface{}) (*%s, error) {\n", daoName, model.Name))
	content.WriteString("\tquery := `\n")
	content.WriteString(fmt.Sprintf("\t\tSELECT %s\n", strings.Join(columns, ", ")))
	content.WriteString(fmt.Sprintf("\t\tFROM %s\n", model.TableName))
	content.WriteString("\t`\n\n")

	content.WriteString("\tif where != \"\" {\n")
	content.WriteString("\t\tquery += \" WHERE \" + where\n")
	content.WriteString("\t}\n\n")

	content.WriteString("\trow := dao.queryRowContext(ctx, query, args...)\n\n")

	content.WriteString(fmt.Sprintf("\tvar m %s\n", model.Name))
	content.WriteString("\terr := row.Scan(\n")
	for _, arg := range scanArgs {
		content.WriteString(fmt.Sprintf("\t\t%s,\n", arg))
	}
	content.WriteString("\t)\n\n")

	content.WriteString("\tif err != nil {\n")
	content.WriteString("\t\treturn nil, err\n")
	content.WriteString("\t}\n\n")

	content.WriteString("\treturn &m, nil\n")
	content.WriteString("}\n\n")

	return content.String()
}

func generateSQLiteFindAllMethod(model parser.Model, daoName string) string {
	var content strings.Builder
	var columns []string
	var scanArgs []string

	for _, field := range model.Fields {
		columns = append(columns, field.Column)
		scanArgs = append(scanArgs, fmt.Sprintf("&m.%s", field.Name))
	}

	content.WriteString(fmt.Sprintf("func (dao *%s) FindAll(ctx context.Context, where string, args ...interface{}) ([]*%s, error) {\n", daoName, model.Name))
	content.WriteString("\tquery := `\n")
	content.WriteString(fmt.Sprintf("\t\tSELECT %s\n", strings.Join(columns, ", ")))
	content.WriteString(fmt.Sprintf("\t\tFROM %s\n", model.TableName))
	content.WriteString("\t`\n\n")

	content.WriteString("\tif where != \"\" {\n")
	content.WriteString("\t\tquery += \" WHERE \" + where\n")
	content.WriteString("\t}\n\n")

	content.WriteString("\trows, err := dao.queryContext(ctx, query, args...)\n")
	content.WriteString("\tif err != nil {\n")
	content.WriteString("\t\treturn nil, err\n")
	content.WriteString("\t}\n")
	content.WriteString("\tdefer rows.Close()\n\n")

	content.WriteString(fmt.Sprintf("\tvar models []*%s\n", model.Name))
	content.WriteString("\tfor rows.Next() {\n")
	content.WriteString(fmt.Sprintf("\t\tvar m %s\n", model.Name))
	content.WriteString("\t\terr := rows.Scan(\n")
	for _, arg := range scanArgs {
		content.WriteString(fmt.Sprintf("\t\t\t%s,\n", arg))
	}
	content.WriteString("\t\t)\n")
	content.WriteString("\t\tif err != nil {\n")
	content.WriteString("\t\t\treturn nil, err\n")
	content.WriteString("\t\t}\n")
	content.WriteString("\t\tmodels = append(models, &m)\n")
	content.WriteString("\t}\n\n")

	content.WriteString("\tif err := rows.Err(); err != nil {\n")
	content.WriteString("\t\treturn nil, err\n")
	content.WriteString("\t}\n\n")

	content.WriteString("\treturn models, nil\n")
	content.WriteString("}\n\n")

	return content.String()
}

func generateSQLiteFindPaginatedMethod(model parser.Model, daoName string) string {
	var content strings.Builder
	var columns []string
	var scanArgs []string

	for _, field := range model.Fields {
		columns = append(columns, field.Column)
		scanArgs = append(scanArgs, fmt.Sprintf("&m.%s", field.Name))
	}

	content.WriteString(fmt.Sprintf("func (dao *%s) FindPaginated(ctx context.Context, limit, offset int, where string, args ...interface{}) ([]*%s, error) {\n", daoName, model.Name))
	content.WriteString("\tquery := `\n")
	content.WriteString(fmt.Sprintf("\t\tSELECT %s\n", strings.Join(columns, ", ")))
	content.WriteString(fmt.Sprintf("\t\tFROM %s\n", model.TableName))
	content.WriteString("\t`\n\n")

	content.WriteString("\tif where != \"\" {\n")
	content.WriteString("\t\tquery += \" WHERE \" + where\n")
	content.WriteString("\t}\n\n")

	content.WriteString("\tquery += fmt.Sprintf(\" LIMIT %d OFFSET %d\", limit, offset)\n\n")

	content.WriteString("\trows, err := dao.queryContext(ctx, query, args...)\n")
	content.WriteString("\tif err != nil {\n")
	content.WriteString("\t\treturn nil, err\n")
	content.WriteString("\t}\n")
	content.WriteString("\tdefer rows.Close()\n\n")

	content.WriteString(fmt.Sprintf("\tvar models []*%s\n", model.Name))
	content.WriteString("\tfor rows.Next() {\n")
	content.WriteString(fmt.Sprintf("\t\tvar m %s\n", model.Name))
	content.WriteString("\t\terr := rows.Scan(\n")
	for _, arg := range scanArgs {
		content.WriteString(fmt.Sprintf("\t\t\t%s,\n", arg))
	}
	content.WriteString("\t\t)\n")
	content.WriteString("\t\tif err != nil {\n")
	content.WriteString("\t\t\treturn nil, err\n")
	content.WriteString("\t\t}\n")
	content.WriteString("\t\tmodels = append(models, &m)\n")
	content.WriteString("\t}\n\n")

	content.WriteString("\tif err := rows.Err(); err != nil {\n")
	content.WriteString("\t\treturn nil, err\n")
	content.WriteString("\t}\n\n")

	content.WriteString("\treturn models, nil\n")
	content.WriteString("}\n\n")

	return content.String()
}

func generateSQLiteCountMethod(model parser.Model, daoName string) string {
	var content strings.Builder

	content.WriteString(fmt.Sprintf("func (dao *%s) Count(ctx context.Context, where string, args ...interface{}) (int64, error) {\n", daoName))
	content.WriteString(fmt.Sprintf("\tquery := \"SELECT COUNT(*) FROM %s\"\n\n", model.TableName))

	content.WriteString("\tif where != \"\" {\n")
	content.WriteString("\t\tquery += \" WHERE \" + where\n")
	content.WriteString("\t}\n\n")

	content.WriteString("\trow := dao.queryRowContext(ctx, query, args...)\n\n")

	content.WriteString("\tvar count int64\n")
	content.WriteString("\terr := row.Scan(&count)\n")
	content.WriteString("\tif err != nil {\n")
	content.WriteString("\t\treturn 0, err\n")
	content.WriteString("\t}\n\n")

	content.WriteString("\treturn count, nil\n")
	content.WriteString("}\n\n")

	return content.String()
}

func generateSQLiteWithTransactionMethod(daoName string) string {
	var content strings.Builder

	content.WriteString(fmt.Sprintf("func (dao *%s) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {\n", daoName))
	content.WriteString("\ttx, err := dao.db.BeginTx(ctx, nil)\n")
	content.WriteString("\tif err != nil {\n")
	content.WriteString("\t\treturn err\n")
	content.WriteString("\t}\n\n")

	content.WriteString("\tctxWithTx := context.WithValue(ctx, \"currentTx\", tx)\n\n")

	content.WriteString("\terr = fn(ctxWithTx)\n")
	content.WriteString("\tif err != nil {\n")
	content.WriteString("\t\tif rbErr := tx.Rollback(); rbErr != nil {\n")
	content.WriteString("\t\t\treturn fmt.Errorf(\"tx err: %v, rb err: %v\", err, rbErr)\n")
	content.WriteString("\t\t}\n")
	content.WriteString("\t\treturn err\n")
	content.WriteString("\t}\n\n")

	content.WriteString("\tif err := tx.Commit(); err != nil {\n")
	content.WriteString("\t\treturn err\n")
	content.WriteString("\t}\n\n")

	content.WriteString("\treturn nil\n")
	content.WriteString("}\n")

	return content.String()
}
