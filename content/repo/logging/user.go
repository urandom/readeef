package logging

import (
	"time"

	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/repo"
	"github.com/urandom/readeef/log"
)

type userRepo struct {
	repo.User

	log log.Log
}

func (r userRepo) Get(login content.Login) (content.User, error) {
	start := time.Now()

	user, err := r.User.Get(login)

	r.log.Infof("repo.User.Get took %s", time.Now().Sub(start))

	return user, err
}

func (r userRepo) All() ([]content.User, error) {
	start := time.Now()

	users, err := r.User.All()

	r.log.Infof("repo.User.All took %s", time.Now().Sub(start))

	return users, err
}

func (r userRepo) Update(user content.User) error {
	start := time.Now()

	err := r.User.Update(user)

	r.log.Infof("repo.User.Update took %s", time.Now().Sub(start))

	return err
}

func (r userRepo) Delete(user content.User) error {
	start := time.Now()

	err := r.User.Delete(user)

	r.log.Infof("repo.User.Delete took %s", time.Now().Sub(start))

	return err
}

func (r userRepo) FindByMD5(data []byte) (content.User, error) {
	start := time.Now()

	user, err := r.User.FindByMD5(data)

	r.log.Infof("repo.User.FindByMD5 took %s", time.Now().Sub(start))

	return user, err
}
