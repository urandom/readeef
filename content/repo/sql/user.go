package sql

import (
	"database/sql"

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

	var user content.User
	err := r.db.Get(&user, r.db.SQL().User.Get, login)
	if err != nil {
		if err == sql.ErrNoRows {
			err = content.ErrNoContent
		}

		return content.User{}, errors.Wrapf(err, "getting user %s", login)
	}

	user.Login = login

	return user, nil
}

func (r userRepo) All() ([]content.User, error) {
	r.log.Infoln("Getting all users")

	var users []content.User
	if err := r.db.Select(&users, r.db.SQL().User.All); err != nil {
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

	tx, err := r.db.Beginx()
	if err != nil {
		return errors.Wrap(err, "creating transaction")
	}
	defer tx.Rollback()

	s := r.db.SQL()
	stmt, err := tx.PrepareNamed(s.User.Update)
	if err != nil {
		return errors.Wrap(err, "preparing user update stmt")
	}
	defer stmt.Close()

	res, err := stmt.Exec(user)
	if err != nil {
		return errors.Wrap(err, "executing user update stmt")
	}

	if num, err := res.RowsAffected(); err == nil && num > 0 {
		if err := tx.Commit(); err != nil {
			return errors.Wrap(err, "committing transaction")
		}

		return nil
	}

	stmt, err = tx.PrepareNamed(s.User.Create)
	if err != nil {
		return errors.Wrap(err, "preparing user create stmt")
	}
	defer stmt.Close()

	_, err = stmt.Exec(user)
	if err != nil {
		return errors.Wrap(err, "executing user create stmt")
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

	r.log.Infof("Deleting user %s", user)

	tx, err := r.db.Beginx()
	if err != nil {
		return errors.Wrap(err, "creating transaction")
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareNamed(r.db.SQL().User.Delete)
	if err != nil {
		return errors.Wrap(err, "preparing user delete stmt")
	}
	defer stmt.Close()

	_, err = stmt.Exec(user)
	if err != nil {
		return errors.Wrap(err, "executing user delete stmt")
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, "committing transaction")
	}

	return nil
}

func (r userRepo) FindByMD5(hash []byte) (content.User, error) {
	if len(hash) == 0 {
		return content.User{}, errors.New("no hash")
	}

	r.log.Infof("Getting user using md5 api field %v", hash)

	var user content.User
	if err := r.db.Get(&user, r.db.SQL().User.GetByMD5API, string(hash)); err != nil {
		if err == sql.ErrNoRows {
			err = content.ErrNoContent
		}

		return content.User{}, errors.Wrapf(err, "getting user by md5 %s", hash)
	}

	user.MD5API = hash

	return user, nil
}
