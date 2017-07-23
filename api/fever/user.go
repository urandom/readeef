package fever

import (
	"encoding/hex"

	"github.com/urandom/readeef"
	"github.com/urandom/readeef/content"
)

func readeefUser(repo content.Repo, md5hex string, log readeef.Logger) content.User {
	md5, err := hex.DecodeString(md5hex)

	if err != nil {
		log.Printf("Error decoding hex api_key")
		return nil
	}

	user := repo.UserByMD5Api(md5)
	if user.HasErr() {
		log.Printf("Error getting user by md5api field: %v\n", user.Err())
		return nil
	}
	return user
}
