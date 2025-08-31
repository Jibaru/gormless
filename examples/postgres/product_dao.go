package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/Jibaru/gormless/examples"
)

type Product = examples.Product

type ProductDAO struct {
	db *sql.DB
}

func NewProductDAO(db *sql.DB) *ProductDAO {
	return &ProductDAO{db: db}
}

func (dao *ProductDAO) getTx(ctx context.Context) *sql.Tx {
	if tx, ok := ctx.Value("currentTx").(*sql.Tx); ok {
		return tx
	}
	return nil
}

func (dao *ProductDAO) execContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	if tx := dao.getTx(ctx); tx != nil {
		return tx.ExecContext(ctx, query, args...)
	}
	return dao.db.ExecContext(ctx, query, args...)
}

func (dao *ProductDAO) queryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	if tx := dao.getTx(ctx); tx != nil {
		return tx.QueryRowContext(ctx, query, args...)
	}
	return dao.db.QueryRowContext(ctx, query, args...)
}

func (dao *ProductDAO) queryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	if tx := dao.getTx(ctx); tx != nil {
		return tx.QueryContext(ctx, query, args...)
	}
	return dao.db.QueryContext(ctx, query, args...)
}

func (dao *ProductDAO) Create(ctx context.Context, p *Product) error {
	query := `
		INSERT INTO products (id, name, description, category, price, stock, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := dao.execContext(
		ctx,
		query,
		p.ID,
		p.Name,
		p.Description,
		p.Category,
		p.Price,
		p.Stock,
		p.CreatedAt,
	)

	return err
}

func (dao *ProductDAO) Update(ctx context.Context, p *Product) error {
	query := `
		UPDATE products
		SET name = $1,
			description = $2,
			category = $3,
			price = $4,
			stock = $5,
			created_at = $6
		WHERE id = $7
	`

	_, err := dao.execContext(ctx, query,
		p.Name,
		p.Description,
		p.Category,
		p.Price,
		p.Stock,
		p.CreatedAt,
		p.ID,
	)
	return err
}

func (dao *ProductDAO) PartialUpdate(ctx context.Context, id string, fields map[string]interface{}) error {
	if len(fields) == 0 {
		return nil
	}

	setClauses := make([]string, 0, len(fields))
	args := make([]interface{}, 0, len(fields)+1)
	i := 1

	for field, value := range fields {
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", field, i))
		args = append(args, value)
		i++
	}

	args = append(args, id)

	query := fmt.Sprintf(`UPDATE products SET %s WHERE id = $%d`, strings.Join(setClauses, ", "), i)

	_, err := dao.execContext(ctx, query, args...)
	return err
}

func (dao *ProductDAO) DeleteByPk(ctx context.Context, pk string) error {
	query := `DELETE FROM products WHERE id = $1`
	_, err := dao.execContext(ctx, query, pk)
	return err
}

func (dao *ProductDAO) FindByPk(ctx context.Context, pk string) (*Product, error) {
	query := `
		SELECT id, name, description, category, price, stock, created_at
		FROM products
		WHERE id = $1
	`
	row := dao.queryRowContext(ctx, query, pk)

	var p Product
	err := row.Scan(
		&p.ID,
		&p.Name,
		&p.Description,
		&p.Category,
		&p.Price,
		&p.Stock,
		&p.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &p, nil
}

func (dao *ProductDAO) CreateMany(ctx context.Context, products []*Product) error {
	if len(products) == 0 {
		return nil
	}

	// Build placeholders dynamically
	placeholders := make([]string, len(products))
	args := make([]interface{}, 0, len(products)*7)

	for i, product := range products {
		placeholders[i] = fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d)",
			i*7+1, i*7+2, i*7+3, i*7+4, i*7+5, i*7+6, i*7+7)

		args = append(args,
			product.ID,
			product.Name,
			product.Description,
			product.Category,
			product.Price,
			product.Stock,
			product.CreatedAt,
		)
	}

	query := fmt.Sprintf(`
		INSERT INTO products (id, name, description, category, price, stock, created_at)
		VALUES %s
	`, strings.Join(placeholders, ", "))

	_, err := dao.execContext(ctx, query, args...)
	return err
}

func (dao *ProductDAO) UpdateMany(ctx context.Context, products []*Product) error {
	if len(products) == 0 {
		return nil
	}

	query := `
		UPDATE products
		SET name = $1,
			description = $2,
			category = $3,
			price = $4,
			stock = $5,
			created_at = $6
		WHERE id = $7
	`

	for _, product := range products {
		_, err := dao.execContext(ctx, query,
			product.Name,
			product.Description,
			product.Category,
			product.Price,
			product.Stock,
			product.CreatedAt,
			product.ID,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func (dao *ProductDAO) DeleteManyByPks(ctx context.Context, pks []string) error {
	if len(pks) == 0 {
		return nil
	}

	// Build placeholders dynamically
	placeholders := make([]string, len(pks))
	args := make([]interface{}, len(pks))
	for i, pk := range pks {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = pk
	}

	query := fmt.Sprintf(`DELETE FROM products WHERE id IN (%s)`, strings.Join(placeholders, ","))
	_, err := dao.execContext(ctx, query, args...)
	return err
}

func (dao *ProductDAO) FindAll(ctx context.Context, where string, args ...interface{}) ([]*Product, error) {
	query := `
		SELECT id, name, description, category, price, stock, created_at
		FROM products
	`

	if where != "" {
		query += " WHERE " + where
	}

	rows, err := dao.queryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []*Product
	for rows.Next() {
		var p Product
		err := rows.Scan(
			&p.ID,
			&p.Name,
			&p.Description,
			&p.Category,
			&p.Price,
			&p.Stock,
			&p.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		products = append(products, &p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return products, nil
}

func (dao *ProductDAO) FindPaginated(ctx context.Context, limit, offset int, where string, args ...interface{}) ([]*Product, error) {
	query := `
		SELECT id, name, description, category, price, stock, created_at
		FROM products
	`

	if where != "" {
		query += " WHERE " + where
	}

	query += fmt.Sprintf(" LIMIT %d OFFSET %d", limit, offset)

	rows, err := dao.queryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []*Product
	for rows.Next() {
		var p Product
		err := rows.Scan(
			&p.ID,
			&p.Name,
			&p.Description,
			&p.Category,
			&p.Price,
			&p.Stock,
			&p.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		products = append(products, &p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return products, nil
}

func (dao *ProductDAO) Count(ctx context.Context, where string, args ...interface{}) (int64, error) {
	query := "SELECT COUNT(*) FROM products"

	if where != "" {
		query += " WHERE " + where
	}

	row := dao.queryRowContext(ctx, query, args...)

	var count int64
	err := row.Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (dao *ProductDAO) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := dao.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	ctxWithTx := context.WithValue(ctx, "currentTx", tx)

	err = fn(ctxWithTx)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx err: %v, rb err: %v", err, rbErr)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}
