package sql

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/repo/sql/db"
	"github.com/urandom/readeef/log"
)

type userRepo struct {
	db *db.DB

	log log.Log
}

func (r userRepo) Get(login content.Login) (content.User, error) {
	r.log.Infof("Getting user %s", login)

	user := content.User{Login: login}
	if err := r.db.WithNamedStmt(r.db.SQL().User.Get, nil, func(stmt *sqlx.NamedStmt) error {
		return stmt.Get(&user, user)
	}); err != nil {
		if err == sql.ErrNoRows {
			err = content.ErrNoContent
		}

		return content.User{}, errors.Wrapf(err, "getting user %s", login)
	}

	return user, nil
}

func (r userRepo) All() ([]content.User, error) {
	r.log.Infoln("Getting all users")

	var users []content.User
	if err := r.db.WithStmt(r.db.SQL().User.All, nil, func(stmt *sqlx.Stmt) error {
		return stmt.Select(&users)
	}); err != nil {
		return users, errors.Wrap(err, "getting all users")
	}

	return users, nil
}

// Update updates or creates the user data in the database.
func (r userRepo) Update(user content.User) error {
	if err := user.Validate(); err != nil {
		return errors.WithMessage(err, "validating user")
	}

	r.log.Infof("Updating user %s", user)

	return r.db.WithTx(func(tx *sqlx.Tx) error {
		s := r.db.SQL()
		return r.db.WithNamedStmt(s.User.Update, tx, func(stmt *sqlx.NamedStmt) error {
			res, err := stmt.Exec(user)
			if err != nil {
				return errors.Wrap(err, "executing user update stmt")
			}

			if num, err := res.RowsAffected(); err == nil && num > 0 {
				return nil
			}

			return r.db.WithNamedStmt(s.User.Create, tx, func(stmt *sqlx.NamedStmt) error {
				if _, err := stmt.Exec(user); err != nil {
					return errors.Wrap(err, "executing user create stmt")
				}

				return nil
			})
		})
	})
}

// Delete deleted the user from the database.
func (r userRepo) Delete(user content.User) error {
	if err := user.Validate(); err != nil {
		return errors.WithMessage(err, "validating user")
	}

	r.log.Infof("Deleting user %s", user)

	return r.db.WithNamedTx(r.db.SQL().User.Delete, func(stmt *sqlx.NamedStmt) error {
		if _, err := stmt.Exec(user); err != nil {
			return errors.Wrap(err, "executing user delete stmt")
		}
		return nil
	})
}

func (r userRepo) FindByMD5(hash []byte) (content.User, error) {
	if len(hash) == 0 {
		return content.User{}, errors.New("no hash")
	}

	r.log.Infof("Getting user using md5 api field %v", hash)

	user := content.User{MD5API: hash}
	if err := r.db.WithNamedStmt(r.db.SQL().User.GetByMD5API, nil, func(stmt *sqlx.NamedStmt) error {
		if err := stmt.Get(&user, user); err != nil {
			if err == sql.ErrNoRows {
				err = content.ErrNoContent
			}

			return errors.Wrapf(err, "getting user by md5 %s", hash)
		}

		return nil
	}); err != nil {
		return content.User{}, err
	}

	return user, nil
}
