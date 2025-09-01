package sqlserver

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/Jibaru/gormless/internal/generator/data/models"
	"strings"
)

type User = models.User

type UserDAO struct {
	db *sql.DB
}

func NewUserDAO(db *sql.DB) *UserDAO {
	return &UserDAO{db: db}
}

func (dao *UserDAO) getTx(ctx context.Context) *sql.Tx {
	if tx, ok := ctx.Value("currentTx").(*sql.Tx); ok {
		return tx
	}
	return nil
}

func (dao *UserDAO) execContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	if tx := dao.getTx(ctx); tx != nil {
		return tx.ExecContext(ctx, query, args...)
	}
	return dao.db.ExecContext(ctx, query, args...)
}

func (dao *UserDAO) queryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	if tx := dao.getTx(ctx); tx != nil {
		return tx.QueryRowContext(ctx, query, args...)
	}
	return dao.db.QueryRowContext(ctx, query, args...)
}

func (dao *UserDAO) queryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	if tx := dao.getTx(ctx); tx != nil {
		return tx.QueryContext(ctx, query, args...)
	}
	return dao.db.QueryContext(ctx, query, args...)
}

func (dao *UserDAO) Create(ctx context.Context, m *User) error {
	query := `
		INSERT INTO users (id, name, email, password, age, deleted_at)
		VALUES (@p1, @p2, @p3, @p4, @p5, @p6)
	`

	_, err := dao.execContext(
		ctx,
		query,
		m.ID,
		m.Name,
		m.Email,
		m.Password,
		m.Age,
		m.DeletedAt,
	)

	return err
}

func (dao *UserDAO) Update(ctx context.Context, m *User) error {
	query := `
		UPDATE users
		SET name = @p1,
			email = @p2,
			password = @p3,
			age = @p4,
			deleted_at = @p5
		WHERE id = @p6
	`

	_, err := dao.execContext(ctx, query,
		m.Name,
		m.Email,
		m.Password,
		m.Age,
		m.DeletedAt,
		m.ID,
	)
	return err
}

func (dao *UserDAO) PartialUpdate(ctx context.Context, pk int, fields map[string]interface{}) error {
	if len(fields) == 0 {
		return nil
	}

	setClauses := make([]string, 0, len(fields))
	args := make([]interface{}, 0, len(fields)+1)
	i := 1

	for field, value := range fields {
		setClauses = append(setClauses, fmt.Sprintf("%s = @p%d", field, i))
		args = append(args, value)
		i++
	}

	args = append(args, pk)

	query := fmt.Sprintf(`UPDATE users SET %s WHERE id = @p%d`, strings.Join(setClauses, ", "), i)

	_, err := dao.execContext(ctx, query, args...)
	return err
}

func (dao *UserDAO) DeleteByPk(ctx context.Context, pk int) error {
	query := `DELETE FROM users WHERE id = @p1`
	_, err := dao.execContext(ctx, query, pk)
	return err
}

func (dao *UserDAO) FindByPk(ctx context.Context, pk int) (*User, error) {
	query := `
		SELECT id, name, email, password, age, deleted_at
		FROM users
		WHERE id = @p1
	`
	row := dao.queryRowContext(ctx, query, pk)

	var m User
	err := row.Scan(
		&m.ID,
		&m.Name,
		&m.Email,
		&m.Password,
		&m.Age,
		&m.DeletedAt,
	)

	if err != nil {
		return nil, err
	}

	return &m, nil
}

func (dao *UserDAO) CreateMany(ctx context.Context, models []*User) error {
	if len(models) == 0 {
		return nil
	}

	placeholders := make([]string, len(models))
	args := make([]interface{}, 0, len(models)*6)

	for i, model := range models {
		placeholders[i] = fmt.Sprintf("(@p%d, @p%d, @p%d, @p%d, @p%d, @p%d)",
			i*6+1, i*6+2, i*6+3, i*6+4, i*6+5, i*6+6)

		args = append(args,
			model.ID,
			model.Name,
			model.Email,
			model.Password,
			model.Age,
			model.DeletedAt,
		)
	}

	query := fmt.Sprintf(`
		INSERT INTO users (id, name, email, password, age, deleted_at)
		VALUES %s
	`, strings.Join(placeholders, ", "))

	_, err := dao.execContext(ctx, query, args...)
	return err
}

func (dao *UserDAO) UpdateMany(ctx context.Context, models []*User) error {
	if len(models) == 0 {
		return nil
	}

	query := `
		UPDATE users
		SET name = @p1,
			email = @p2,
			password = @p3,
			age = @p4,
			deleted_at = @p5
		WHERE id = @p6
	`

	for _, model := range models {
		_, err := dao.execContext(ctx, query,
			model.Name,
			model.Email,
			model.Password,
			model.Age,
			model.DeletedAt,
			model.ID,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func (dao *UserDAO) DeleteManyByPks(ctx context.Context, pks []int) error {
	if len(pks) == 0 {
		return nil
	}

	placeholders := make([]string, len(pks))
	args := make([]interface{}, len(pks))
	for i, pk := range pks {
		placeholders[i] = fmt.Sprintf("@p%d", i+1)
		args[i] = pk
	}

	query := fmt.Sprintf(`DELETE FROM users WHERE id IN (%s)`, strings.Join(placeholders, ","))
	_, err := dao.execContext(ctx, query, args...)
	return err
}

func (dao *UserDAO) FindOne(ctx context.Context, where string, sort string, args ...interface{}) (*User, error) {
	query := `
		SELECT id, name, email, password, age, deleted_at
		FROM users
	`

	if where != "" {
		query += " WHERE " + where
	}

	if sort != "" {
		query += " ORDER BY " + sort
	}

	row := dao.queryRowContext(ctx, query, args...)

	var m User
	err := row.Scan(
		&m.ID,
		&m.Name,
		&m.Email,
		&m.Password,
		&m.Age,
		&m.DeletedAt,
	)

	if err != nil {
		return nil, err
	}

	return &m, nil
}

func (dao *UserDAO) FindAll(ctx context.Context, where string, sort string, args ...interface{}) ([]*User, error) {
	query := `
		SELECT id, name, email, password, age, deleted_at
		FROM users
	`

	if where != "" {
		query += " WHERE " + where
	}

	if sort != "" {
		query += " ORDER BY " + sort
	}

	rows, err := dao.queryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var models []*User
	for rows.Next() {
		var m User
		err := rows.Scan(
			&m.ID,
			&m.Name,
			&m.Email,
			&m.Password,
			&m.Age,
			&m.DeletedAt,
		)
		if err != nil {
			return nil, err
		}
		models = append(models, &m)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return models, nil
}

func (dao *UserDAO) FindPaginated(ctx context.Context, limit, offset int, where string, sort string, args ...interface{}) ([]*User, error) {
	query := `
		SELECT id, name, email, password, age, deleted_at
		FROM users
	`

	if where != "" {
		query += " WHERE " + where
	}

	if sort != "" {
		query += " ORDER BY " + sort
	} else {
		query += " ORDER BY (SELECT NULL)"
	}

	query += fmt.Sprintf(" OFFSET %d ROWS FETCH NEXT %d ROWS ONLY", offset, limit)

	rows, err := dao.queryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var models []*User
	for rows.Next() {
		var m User
		err := rows.Scan(
			&m.ID,
			&m.Name,
			&m.Email,
			&m.Password,
			&m.Age,
			&m.DeletedAt,
		)
		if err != nil {
			return nil, err
		}
		models = append(models, &m)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return models, nil
}

func (dao *UserDAO) Count(ctx context.Context, where string, args ...interface{}) (int64, error) {
	query := "SELECT COUNT(*) FROM users"

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

func (dao *UserDAO) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
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
