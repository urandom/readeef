package base

import "github.com/urandom/readeef/content"

type Repo struct {
	Error

	processors []content.ArticleProcessor
}

type RepoRelated struct {
	repo content.Repo
}

func (r *Repo) ArticleProcessors(p ...[]content.ArticleProcessor) []content.ArticleProcessor {
	if r.HasErr() {
		return []content.ArticleProcessor{}
	}

	if len(p) > 0 {
		r.processors = p[0]
	}

	return r.processors
}

func (rr *RepoRelated) Repo(re ...content.Repo) content.Repo {
	if len(re) > 0 {
		rr.repo = re[0]
	}

	return rr.repo
}
