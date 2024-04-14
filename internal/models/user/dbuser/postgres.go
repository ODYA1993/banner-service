package dbuser

import (
	"banner-service/internal/models/user"
	"banner-service/pkg/db/postgresql"
	"banner-service/pkg/logging"
	"context"
	"fmt"
)

type userRepository struct {
	db     postgresql.Client
	logger *logging.Logger
}

func NewUserRepository(db postgresql.Client, logger *logging.Logger) user.Storage {
	return &userRepository{
		db:     db,
		logger: logger,
	}
}

func (r *userRepository) Create(ctx context.Context, user *user.User) error {
	q := `INSERT INTO "user" (name, email, password, is_admin) VALUES ($1, $2, $3, $4) RETURNING id`

	r.logger.Trace(fmt.Sprintf("SQL Query: %s", q))

	err := r.db.QueryRow(ctx, q, user.Name, user.Email, user.Password, user.IsAdmin).Scan(&user.ID)
	if err != nil {
		return err
	}

	return nil
}

func (r *userRepository) FindAll(ctx context.Context) ([]user.User, error) {
	q := `SELECT id, name, email, password, is_admin FROM "user"`

	r.logger.Trace(fmt.Sprintf("SQL Query: %s", q))

	rows, err := r.db.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]user.User, 0)

	for rows.Next() {
		var us user.User

		err = rows.Scan(&us.ID, &us.Name, &us.Email, &us.Password, &us.IsAdmin)
		if err != nil {
			return nil, err
		}

		users = append(users, us)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func (r *userRepository) FindOne(ctx context.Context, id string) (user.User, error) {
	query := `SELECT id, name, email, password, is_admin FROM "user" WHERE id = $1`

	r.logger.Trace(fmt.Sprintf("SQL Query: %s", query))

	var us user.User
	err := r.db.QueryRow(ctx, query, id).Scan(&us.ID, &us.Name, &us.Email, &us.Password, &us.IsAdmin)
	if err != nil {
		return user.User{}, err
	}

	return us, nil
}

func (r *userRepository) FindOneByEmail(ctx context.Context, email string) (*user.User, error) {
	query := `SELECT id, name, email, password, is_admin FROM "user" WHERE email = $1`

	r.logger.Trace(fmt.Sprintf("SQL Query: %s", query))

	var us user.User
	err := r.db.QueryRow(ctx, query, email).Scan(&us.ID, &us.Name, &us.Email, &us.Password, &us.IsAdmin)
	if err != nil {
		return nil, err
	}

	return &us, nil
}

func (r *userRepository) Delete(ctx context.Context, id string) (int64, error) {
	query := `DELETE FROM "user" WHERE id=$1`

	r.logger.Trace(fmt.Sprintf("SQL Query: %s", query))

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return 0, err
	}

	rowsAffected := result.RowsAffected()

	return rowsAffected, nil
}
