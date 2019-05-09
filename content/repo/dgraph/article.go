package dgraph

import (
	"strings"
	"text/template"

	"github.com/Masterminds/sprig"
	"github.com/dgraph-io/dgo"
	"github.com/pkg/errors"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/log"
)

type articleRepo struct {
	dg *dgo.Dgraph

	log log.Log
}

func (r articleRepo) ForUser(user content.User, opts ...content.QueryOpt) ([]content.Article, error) {
	panic("not implemented")
}

func (r articleRepo) All(opts ...content.QueryOpt) ([]content.Article, error) {
	panic("not implemented")
}

func (r articleRepo) Count(user content.User, opts ...content.QueryOpt) (int64, error) {
	panic("not implemented")
}

func (r articleRepo) IDs(user content.User, opts ...content.QueryOpt) ([]content.ArticleID, error) {
	panic("not implemented")
}

func (r articleRepo) Read(state bool, user content.User, opts ...content.QueryOpt) error {
	panic("not implemented")
}

func (r articleRepo) Favor(state bool, user content.User, opts ...content.QueryOpt) error {
	panic("not implemented")
}

func (r articleRepo) RemoveStaleUnreadRecords() error {
	panic("not implemented")
}

type queryParams struct {
	Params     []string
	Limit      int
	Offset     int
	Facets     []string
	Sorting    []string
	Filter     []string
	Predicates string
}

var (
	base = template.Must(template.New("base").Funcs(sprig.TxtFuncMap()).Parse(`
query Articles({{ .Params | join ", " }}) {
	{{ template "articles" }}

	articles(func: uid(articles)) {
		{{ .Predicates }}
	}
}
	`))
	userless = template.Must(template.Must(base.Clone()).Parse(`
{{ define "article" }}

articles as var(func: has(article.link), first: {{ .Limit }}
{{- if .Offset }}, offset: {{ .Offset }} {{ end -}}
{{- if .Sorting }}, {{ .Sorting | join ", " }} {{ end -}}
) {{- if .Filter }} @filter({{ .Filter | join " AND " }}) {{ end -}} @cascade {
	article.link
}

{{ end }}
	`))
	untaggedOnly = template.Must(template.Must(base.Clone()).Parse(`
{{ define "article" }}

var(func: uid($id)) {
	tag {
		taggedFeeds as feed {
			uid
		}
	}
}

var(func: has(feed.link)) @filter(NOT uid(taggedFeeds)) @cascade {
	articles as article (first: {{ .Limit }} {{- if .Offset }}, offset: {{ .Offset }} {{ end -}})
	{{- if .Sorting -}} ( {{- .Sorting | join ", " -}} ) {{- end -}}
    {{- if .Filter }} @filter({{ .Filter | join " AND " }}) {{ end -}}
    {{- if .Facets }} @facets({{ .Facets | join " AND " }}) {{ end -}}
	{
		article.link
	}
}

{{ end }}
	`))

	unreadFacet   = `eq(unread, true)`
	readFacet     = `NOT eq(unread, true)`
	favoriteFacet = `eq(favorite, true)`

	faceted = template.Must(template.Must(base.Clone()).Parse(`
{{ define "article" }}

var(func: uid($id)) @cascade {
	feed {
		articles as article {{- if .Facets }} @facets({{ .Facets | join " AND " }}) {{ end -}}
		(first: {{ .Limit }} {{- if .Offset }}, offset: {{ .Offset }} {{ end -}})
		{{- if .Sorting -}} ( {{- .Sorting | join ", " -}} ) {{- end -}}
        {{- if .Filter }} @filter({{ .Filter | join " AND " }}) {{ end -}}
		{
			article.link
		}
	}
}

{{ end }}
	`))
)

func constructQuery(
	userUID UID,
	opts content.QueryOptions,
) (string, map[string]string, error) {
	var b strings.Builder

	q := queryParams{Params: make([]string, 0, 4), Facets: make([]string, 0, 2), Filter: make([]string, 0, 4)}

	args := map[string]string{}

	var tmpl *template.Template

	if userUID.Valid() {
		args["$id"] = userUID.Value
		q.Params = append(q.Params, "$id: string")

		if opts.UntaggedOnly {
			tmpl = untaggedOnly
		} else {
			tmpl = faceted

			if opts.UnreadOnly {
				q.Facets = append(q.Facets, unreadFacet)
			} else if opts.ReadOnly {
				q.Facets = append(q.Facets, readFacet)
			}

			if opts.FavoriteOnly {
				q.Facets = append(q.Facets, favoriteFacet)
			}
		}
	} else {
		tmpl = userless
	}

	if opts.BeforeID > 0 {
		q.Filter = append(q.Filter)
	}

	if err := tmpl.Execute(&b, q); err != nil {
		return "", nil, errors.Wrap(err, "executing template")
	}

	return b.String(), args, nil
}
