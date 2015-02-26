package base

import "github.com/urandom/readeef/content"

type Repo struct {
	Error
}

type RepoRelated struct {
	repo content.Repo
}

func (rr *RepoRelated) Repo(re ...content.Repo) content.Repo {
	if len(re) > 0 {
		rr.repo = re[0]
	}

	return rr.repo
}
