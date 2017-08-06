package fever

import (
	"encoding/hex"

	"github.com/pkg/errors"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/repo"
	"github.com/urandom/readeef/log"
)

func readeefUser(repo repo.User, md5hex string, log log.Log) (content.User, error) {
	md5, err := hex.DecodeString(md5hex)

	if err != nil {
		return content.User{}, errors.Wrap(err, "decoding hex api_key")
	}

	user, err := repo.FindByMD5(md5)
	if err != nil {
		return content.User{}, errors.WithMessage(err, "getting user by md5")
	}
	return user, nil
}
