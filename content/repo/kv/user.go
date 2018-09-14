package kv

import (
	"github.com/asdine/storm"
	"github.com/pkg/errors"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/log"
)

type userRepo struct {
	db  *storm.DB
	log log.Log
}

const (
	usersBucket = "users"
)

func (r userRepo) Get(login content.Login) (content.User, error) {
	r.log.Infof("Getting user %s", login)

	var user content.User
	if err := r.db.From(usersBucket).One("Login", login, &user); err != nil {
		if err == storm.ErrNotFound {
			err = content.ErrNoContent
		}

		return content.User{}, errors.Wrapf(err, "getting user %s", login)
	}

	return user, nil
}

func (r userRepo) All() ([]content.User, error) {
	r.log.Infoln("Getting all users")

	var users []content.User
	if err := r.db.From(usersBucket).All(&users); err != nil {
		return nil, errors.Wrap(err, "getting all users")
	}

	return users, nil
}

func (r userRepo) Update(user content.User) error {
	if err := user.Validate(); err != nil {
		return errors.WithMessage(err, "validating user")
	}

	r.log.Infof("Updating user %s", user)

	if err := r.db.From(usersBucket).Save(&user); err != nil {
		return errors.Wrapf(err, "updating user %s", user.Login)
	}

	return nil
}

func (r userRepo) Delete(user content.User) error {
	if err := user.Validate(); err != nil {
		return errors.WithMessage(err, "validating user")
	}

	r.log.Infof("Deleting user %s", user)

	tx, err := r.db.Begin(true)
	if err != nil {
		return errors.Wrap(err, "starting transaction")
	}
	defer tx.Rollback()

	if err := tx.From(usersBucket).DeleteStruct(&user); err != nil {
		return errors.Wrapf(err, "deleting user %s", user.Login)
	}

	if err := deleteFeedUserConnections(tx, user); err != nil {
		return err
	}

	return tx.Commit()
}

func (r userRepo) FindByMD5(hash []byte) (content.User, error) {
	if len(hash) == 0 {
		return content.User{}, errors.New("no hash")
	}

	r.log.Infof("Getting user using md5 api field %v", hash)

	var user content.User
	if err := r.db.From(usersBucket).One("MD5API", hash, &user); err != nil {
		if err == storm.ErrNotFound {
			err = content.ErrNoContent
		}

		return content.User{}, errors.Wrapf(err, "getting user by md5 %s", hash)
	}

	return user, nil
}

func newUserRepo(db *storm.DB, log log.Log) (userRepo, error) {
	if err := db.From(usersBucket).Init(&content.User{}); err != nil {
		return userRepo{}, errors.Wrap(err, "initializing users indices")
	}

	return userRepo{db, log}, nil
}
