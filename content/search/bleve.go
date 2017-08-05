package search

import (
	"bytes"
	"html"
	"os"
	"strconv"
	"time"

	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/index/store/goleveldb"
	"github.com/blevesearch/bleve/index/upsidedown"
	"github.com/blevesearch/bleve/search"
	"github.com/blevesearch/bleve/search/query"
	"github.com/pkg/errors"
	"github.com/urandom/readeef"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/repo"
)

type bleveSearch struct {
	index     bleve.Index
	log       readeef.Logger
	newIndex  bool
	batchSize int64
	service   repo.Service
}

type indexArticle struct {
	FeedID      string    `json:"feed_id"`
	ArticleID   string    `json:"article_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Link        string    `json:"link"`
	Date        time.Time `json:"date"`
}

func NewBleve(path string, size int64, service repo.Service, log readeef.Logger) (bleveSearch, error) {
	var err error
	var exists bool
	var index bleve.Index

	_, err = os.Stat(path)
	if err == nil {
		log.Infoln("Opening search index " + path)
		index, err = bleve.Open(path)

		if err != nil {
			return nil, errors.Wrap(err, "opening bleve search index")
		}

		exists = true
	} else if os.IsNotExist(err) {
		mapping := bleve.NewIndexMapping()
		docMapping := bleve.NewDocumentMapping()

		idfieldmapping := bleve.NewTextFieldMapping()
		idfieldmapping.IncludeInAll = false
		docMapping.AddFieldMappingsAt("FeedID", idfieldmapping)
		docMapping.AddFieldMappingsAt("ArticleID", idfieldmapping)

		mapping.AddDocumentMapping(mapping.DefaultType, docMapping)

		log.Infoln("Creating search index " + path)
		index, err = bleve.NewUsing(path, mapping, upsidedown.Name, goleveldb.Name, nil)

		if err != nil {
			return nil, errors.Wrap(err, "creating search index")
		}
	} else {
		return nil, errors.Wrapf(err, "getting file '%s' stat", path)
	}

	return &bleveSearch{log: log, index: index, batchSize: size, service: service, newIndex: !exists}, nil
}

func (b bleveSearch) IsNewIndex() bool {
	return b.newIndex
}

func (b bleveSearch) Search(
	term string,
	u content.User,
	opts ...content.QueryOpt,
) ([]content.Article, error) {

	o := content.QueryOptions{}
	o.Apply(opts)

	var q query.Query

	q = bleve.NewQueryStringQuery(term)

	feedIDs := o.FeedIDs

	if len(feedIDs) == 0 {
		var err error
		if feedIDs, err = b.service.FeedRepo().IDs(); err != nil {
			return []content.Article{}, errors.WithMessage(err, "getting feed ids")
		} else if len(feedIDs) == 0 {
			return []content.Article{}, nil
		}
	}

	queries := make([]query.Query, len(feedIDs))
	conjunct := make([]query.Query, 2)

	for i, id := range feedIDs {
		q := bleve.NewTermQuery(strconv.FormatInt(int64(id), 10))
		q.SetField("FeedID")

		queries[i] = q
	}

	disjunct := bleve.NewDisjunctionQuery(queries...)

	conjunct[0] = q
	conjunct[1] = disjunct

	q = bleve.NewConjunctionQuery(conjunct...)

	searchRequest := bleve.NewSearchRequest(q)

	searchRequest.Highlight = bleve.NewHighlightWithStyle("html")

	searchRequest.Size = o.Limit
	searchRequest.From = o.Offset

	order := ""
	if o.SortOrder == content.DescendingOrder {
		order = "-"
	}
	switch o.SortField {
	case content.SortByDate:
		searchRequest.SortBy([]string{order + "Date"})
	case content.SortByID:
		searchRequest.SortBy([]string{order + "ArticleID"})
	case content.DefaultSort:
		searchRequest.SortBy([]string{order + "_score"})
	}

	searchResult, err := b.index.Search(searchRequest)

	if err != nil {
		return []content.Article{}, errors.Wrap(err, "searching")
	}

	if len(searchResult.Hits) == 0 {
		return []content.Article{}, nil
	}

	articleIDs := []content.ArticleID{}
	hitMap := map[content.ArticleID]*search.DocumentMatch{}

	for _, hit := range searchResult.Hits {
		if articleID, err := strconv.ParseInt(hit.ID, 10, 64); err == nil {
			id := content.ArticleID(articleID)
			articleIDs = append(articleIDs, id)
			hitMap[id] = hit
		}
	}

	articles, err := b.service.ArticleRepo().All(content.IDs(articleIDs))
	if err != nil {
		return []content.Article{}, errors.WithMessage(err, "getting articles by ids")
	}

	for i := range articles {
		hit := hitMap[article.ID]

		if len(hit.Fragments) > 0 {
			articles[i].Hit.Fragments = hit.Fragments
		}
	}

	return articles, nil
}

func (b bleveSearch) BatchIndex(articles []content.Article, op indexOperation) error {
	if len(articles) == 0 {
		return nil
	}

	batch := b.index.NewBatch()
	count := int64(0)

	for i := range articles {
		a := articles[i]

		switch op {
		case BatchAdd:
			b.log.Debugf("Indexing article '%d' of feed id '%d'\n", a.ID, a.FeedID)

			batch.Index(prepareArticle(a))
		case BatchDelete:
			b.log.Debugf("Removing article '%d' of feed id '%d' from index\n", a.ID, a.FeedID)

			batch.Delete(strconv.FormatInt(int64(a.ID), 10))
		default:
			return errors.Errorf("unknown operation type %v", op)
		}

		count++

		if count >= b.batchSize {
			if err := b.index.Batch(batch); err != nil {
				return errors.Wrap(err, "indexing article batch")
			}
			batch = b.index.NewBatch()
			count = 0
		}
	}

	if count > 0 {
		if err := b.index.Batch(batch); err != nil {
			return errors.Wrap(err, "indexing article batch")
		}
	}

	return nil
}

func prepareArticle(article content.Article) (string, indexArticle) {
	id := strconv.FormatInt(int64(article.ID), 10)
	ia := indexArticle{
		FeedID:      strconv.FormatInt(int64(article.FeedID), 10),
		ArticleID:   article.ID,
		Title:       html.UnescapeString(StripTags(article.Title)),
		Description: html.UnescapeString(StripTags(article.Description)),
		Link:        article.Link, Date: article.Date,
	}

	return id, ia
}

func StripTags(text string) string {
	b := bytes.NewBufferString("")
	inTag := 0

	for _, r := range text {
		switch r {
		case '<':
			inTag++
		case '>':
			if inTag > 0 {
				inTag--
			}
		default:
			if inTag < 1 {
				b.WriteRune(r)
			}
		}
	}

	return b.String()
}
