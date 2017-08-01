package sql

import (
	"database/sql"

	"github.com/pkg/errors"
	"github.com/urandom/readeef"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/sql/db"
)

type userRepo struct {
	db *db.DB

	log readeef.Logger
}

func (r userRepo) Get(login, content.Login) (content.User, error) {
	r.log.Infof("Getting user '%s'\n", login)

	user := r.db.Get(&data, r.db.SQL().Repo.GetUser, login)
	if err != nil {
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
	if err := r.db.Select(&users, r.db.SQL().Repo.GetUsers); err != nil {
		return users, errors.Wrap(err, "getting all users")
	}

	return users
}

// Update updates or creates the user data in the database.
func (r userRepo) Update(user content.User) error {
	if err := user.Validate(); err != nil {
		return errors.WithMessage(err, "validating user")
	}

	r.log.Infof("Updating user %s", user)

	tx, err := r.db.Beginx()
	if err != nil {
		return errors.Wrap(err, "creating transaction")
	}
	defer tx.Rollback()

	s := r.db.SQL()
	stmt, err := tx.Preparex(s.User.Update)
	if err != nil {
		return errors.Wrap(err, "preparing user update stmt")
	}
	defer stmt.Close()

	res, err := stmt.Exec(i.FirstName, i.LastName, i.Email, i.Admin, i.Active, i.ProfileJSON, i.HashType, i.Salt, i.Hash, i.MD5API, i.Login)
	if err != nil {
		return errors.Wrap(err, "executimg user update stmt")
	}

	if num, err := res.RowsAffected(); err == nil && num > 0 {
		if err := tx.Commit(); err != nil {
			return errors.Wrap(err, "committing transaction")
		}

		return nil
	}

	stmt, err = tx.Preparex(s.User.Create)
	if err != nil {
		return errors.Wrap(err, "preparing user create stmt")
	}
	defer stmt.Close()

	_, err = stmt.Exec(i.Login, i.FirstName, i.LastName, i.Email, i.Admin, i.Active, i.ProfileJSON, i.HashType, i.Salt, i.Hash, i.MD5API)
	if err != nil {
		return errors.Wrap(err, "executimg user create stmt")
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, "committing transaction")
	}

	return nil
}

// Delete deleted the user from the database.
func (r userRepo) Delete(user content.User) error {
	if err := user.Validate(); err != nil {
		return errors.WithMessage(err, "validating user")
	}

	r.log.Infof("Deleting user %s", r)

	tx, err := r.db.Beginx()
	if err != nil {
		return errors.Wrap(err, "creating transaction")
	}
	defer tx.Rollback()

	stmt, err := tx.Preparex(r.db.SQL().User.Delete)
	if err != nil {
		return errors.Wrap(err, "preparing user delete stmt")
	}
	defer stmt.Close()

	_, err = stmt.Exec(i.Login)
	if err != nil {
		return errors.Wrap(err, "executimg user delete stmt")
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, "committing transaction")
	}

	return nil
}
