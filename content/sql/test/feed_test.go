package test

import (
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/data"
)

func createFeed(d data.Feed) (f content.Feed) {
	f = repo.Feed()
	f.Data(d)
	f.Update()

	return
}

func createUserFeed(u content.User, d data.Feed) (uf content.UserFeed) {
	uf = repo.UserFeed(u)
	uf.Data(d)
	uf.Update()

	u.AddFeed(uf)

	return
}

func createTaggedFeed(u content.User, d data.Feed) (tf content.TaggedFeed) {
	tf = repo.TaggedFeed(u)
	tf.Data(d)
	tf.Update()

	u.AddFeed(tf)

	return
}
