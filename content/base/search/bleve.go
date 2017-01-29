package search

import (
	"bytes"
	"errors"
	"fmt"
	"html"
	"os"
	"strconv"
	"time"

	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/index/store/goleveldb"
	"github.com/blevesearch/bleve/index/upsidedown"
	"github.com/blevesearch/bleve/search"
	"github.com/blevesearch/bleve/search/query"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/base"
	"github.com/urandom/readeef/content/data"
	"github.com/urandom/webfw"
)

type Bleve struct {
	base.ArticleSorting
	index     bleve.Index
	logger    webfw.Logger
	newIndex  bool
	batchSize int64
}

type indexArticle struct {
	FeedId      string    `json:"feed_id"`
	ArticleId   string    `json:"article_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Link        string    `json:"link"`
	Date        time.Time `json:"date"`
}

func NewBleve(path string, size int64, logger webfw.Logger) (content.SearchProvider, error) {
	var err error
	var exists bool
	var index bleve.Index

	_, err = os.Stat(path)
	if err == nil {
		logger.Infoln("Opening search index " + path)
		index, err = bleve.Open(path)

		if err != nil {
			return nil, errors.New(fmt.Sprintf("Error opening search index: %v\n", err))
		}

		exists = true
	} else if os.IsNotExist(err) {
		mapping := bleve.NewIndexMapping()
		docMapping := bleve.NewDocumentMapping()

		idfieldmapping := bleve.NewTextFieldMapping()
		idfieldmapping.IncludeInAll = false
		docMapping.AddFieldMappingsAt("FeedId", idfieldmapping)
		docMapping.AddFieldMappingsAt("ArticleId", idfieldmapping)

		mapping.AddDocumentMapping(mapping.DefaultType, docMapping)

		logger.Infoln("Creating search index " + path)
		index, err = bleve.NewUsing(path, mapping, upsidedown.Name, goleveldb.Name, nil)

		if err != nil {
			return nil, errors.New(fmt.Sprintf("Error creating search index: %v\n", err))
		}
	} else {
		return nil, errors.New(
			fmt.Sprintf("Error getting stat of '%s': %v\n", path, err))
	}

	return &Bleve{logger: logger, index: index, batchSize: size, newIndex: !exists}, nil
}

func (b Bleve) IsNewIndex() bool {
	return b.newIndex
}

func (b Bleve) IndexAllFeeds(repo content.Repo) error {
	b.logger.Infoln("Indexing all articles")

	for _, f := range repo.AllFeeds() {
		articles := f.AllArticles()
		if f.HasErr() {
			return f.Err()
		}

		if err := b.BatchIndex(articles, data.BatchAdd); err != nil {
			return err
		}
	}

	return repo.Err()
}

func (b Bleve) Search(term string, u content.User, feedIds []data.FeedId, limit, offset int) (ua []content.UserArticle, err error) {
	var q query.Query

	q = bleve.NewQueryStringQuery(term)

	if len(feedIds) > 0 {
		queries := make([]query.Query, len(feedIds))
		conjunct := make([]query.Query, 2)

		for i, id := range feedIds {
			q := bleve.NewTermQuery(strconv.FormatInt(int64(id), 10))
			q.SetField("FeedId")

			queries[i] = q
		}

		disjunct := bleve.NewDisjunctionQuery(queries...)

		conjunct[0] = q
		conjunct[1] = disjunct

		q = bleve.NewConjunctionQuery(conjunct...)
	}

	searchRequest := bleve.NewSearchRequest(q)

	searchRequest.Highlight = bleve.NewHighlightWithStyle("html")

	searchRequest.Size = limit
	searchRequest.From = offset

	order := ""
	if b.Order() == data.DescendingOrder {
		order = "-"
	}
	switch b.Field() {
	case data.SortByDate:
		searchRequest.SortBy([]string{order + "Date"})
	case data.SortById:
		searchRequest.SortBy([]string{order + "ArticleId"})
	case data.DefaultSort:
		searchRequest.SortBy([]string{order + "_score"})
	}

	searchResult, err := b.index.Search(searchRequest)

	if err != nil {
		return
	}

	if len(searchResult.Hits) == 0 {
		return
	}

	articleIds := []data.ArticleId{}
	hitMap := map[data.ArticleId]*search.DocumentMatch{}

	for _, hit := range searchResult.Hits {
		if articleId, err := strconv.ParseInt(hit.ID, 10, 64); err == nil {
			id := data.ArticleId(articleId)
			articleIds = append(articleIds, id)
			hitMap[id] = hit
		}
	}

	ua = u.ArticlesById(articleIds)
	if u.HasErr() {
		return ua, u.Err()
	}

	for i := range ua {
		data := ua[i].Data()

		hit := hitMap[data.Id]

		if len(hit.Fragments) > 0 {
			data.Hit.Fragments = hit.Fragments
			ua[i].Data(data)
		}
	}

	return
}

func (b Bleve) BatchIndex(articles []content.Article, op data.IndexOperation) error {
	if len(articles) == 0 {
		return nil
	}

	batch := b.index.NewBatch()
	count := int64(0)

	for i := range articles {
		d := articles[i].Data()

		switch op {
		case data.BatchAdd:
			b.logger.Debugf("Indexing article '%d' of feed id '%d'\n", d.Id, d.FeedId)

			batch.Index(prepareArticle(d))
		case data.BatchDelete:
			b.logger.Debugf("Removing article '%d' of feed id '%d' from index\n", d.Id, d.FeedId)

			batch.Delete(strconv.FormatInt(int64(d.Id), 10))
		default:
			return fmt.Errorf("Unknown operation type %v", op)
		}

		count++

		if count >= b.batchSize {
			if err := b.index.Batch(batch); err != nil {
				return fmt.Errorf("Error indexing article batch: %v\n", err)
			}
			batch = b.index.NewBatch()
			count = 0
		}
	}

	if count > 0 {
		if err := b.index.Batch(batch); err != nil {
			return fmt.Errorf("Error indexing article batch: %v\n", err)
		}
	}

	return nil
}

func prepareArticle(data data.Article) (string, indexArticle) {
	id := strconv.FormatInt(int64(data.Id), 10)
	ia := indexArticle{FeedId: strconv.FormatInt(int64(data.FeedId), 10),
		ArticleId:   id,
		Title:       html.UnescapeString(StripTags(data.Title)),
		Description: html.UnescapeString(StripTags(data.Description)),
		Link:        data.Link, Date: data.Date,
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
