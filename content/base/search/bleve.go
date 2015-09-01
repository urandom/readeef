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
	"github.com/blevesearch/bleve/search"
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

func NewBleve(path string, size int64, logger webfw.Logger) (*Bleve, error) {
	var err error
	var index bleve.Index

	b := &Bleve{}

	_, err = os.Stat(path)
	if err == nil {
		logger.Infoln("Opening search index " + path)
		index, err = bleve.Open(path)

		if err != nil {
			return b, errors.New(fmt.Sprintf("Error opening search index: %v\n", err))
		}
	} else if os.IsNotExist(err) {
		mapping := bleve.NewIndexMapping()
		docMapping := bleve.NewDocumentMapping()

		idfieldmapping := bleve.NewTextFieldMapping()
		idfieldmapping.IncludeInAll = false
		docMapping.AddFieldMappingsAt("FeedId", idfieldmapping)
		docMapping.AddFieldMappingsAt("ArticleId", idfieldmapping)

		mapping.AddDocumentMapping(mapping.DefaultType, docMapping)

		logger.Infoln("Creating search index " + path)
		index, err = bleve.NewUsing(path, mapping, "goleveldb", nil)

		if err != nil {
			return b, errors.New(fmt.Sprintf("Error creating search index: %v\n", err))
		}

		b.newIndex = true
	} else {
		return b, errors.New(
			fmt.Sprintf("Error getting stat of '%s': %v\n", path, err))
	}

	b.logger = logger
	b.index = index
	b.batchSize = size

	return b, nil
}

func (b Bleve) IsNewIndex() bool {
	return b.newIndex
}

func (b Bleve) IndexAllArticles(repo content.Repo) error {
	b.logger.Infoln("Indexing all articles")

	for _, f := range repo.AllFeeds() {
		articles := f.AllArticles()
		if f.HasErr() {
			return f.Err()
		}

		b.batchIndex(articles)
	}

	return repo.Err()
}

func (b Bleve) UpdateFeed(feed content.Feed) {
	b.logger.Infof("Updating article search index for feed '%s'\n", feed)

	newArticleLinks := map[string]bool{}
	for _, a := range feed.NewArticles() {
		newArticleLinks[a.Data().Link] = true
	}

	var articles []content.Article
	for _, a := range feed.NewArticles() {
		if newArticleLinks[a.Data().Link] {
			articles = append(articles, a)
		}
	}

	b.batchIndex(articles)
}

func (b Bleve) DeleteFeed(feed content.Feed) error {
	articles := feed.AllArticles()

	if !feed.HasErr() {
		b.logger.Infof("Removing all articles from the search index for feed '%s'\n", feed)

		b.batchDelete(articles)
	} else {
		return feed.Err()
	}
	return nil
}

func (b Bleve) Search(term string, u content.User, feedIds []data.FeedId, limit, offset int) (ua []content.UserArticle, err error) {
	var query bleve.Query

	query = bleve.NewQueryStringQuery(term)

	if len(feedIds) > 0 {
		queries := make([]bleve.Query, len(feedIds))
		conjunct := make([]bleve.Query, 2)

		for i, id := range feedIds {
			q := bleve.NewTermQuery(strconv.FormatInt(int64(id), 10))
			q.SetField("FeedId")

			queries[i] = q
		}

		disjunct := bleve.NewDisjunctionQuery(queries)

		conjunct[0] = query
		conjunct[1] = disjunct

		query = bleve.NewConjunctionQuery(conjunct)
	}

	searchRequest := bleve.NewSearchRequest(query)

	searchRequest.Highlight = bleve.NewHighlightWithStyle("html")

	searchRequest.Size = limit
	searchRequest.From = offset

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

func (b Bleve) batchIndex(articles []content.Article) {
	if len(articles) == 0 {
		return
	}

	batch := b.index.NewBatch()
	count := int64(0)

	for i := range articles {
		data := articles[i].Data()

		b.logger.Debugf("Indexing article '%d' from feed id '%d'\n", data.Id, data.FeedId)

		batch.Index(prepareArticle(data))
		count++

		if count >= b.batchSize {
			if err := b.index.Batch(batch); err != nil {
				b.logger.Printf("Error indexing article batch: %v\n", err)
			}
			batch = b.index.NewBatch()
			count = 0
		}
	}

	if count > 0 {
		if err := b.index.Batch(batch); err != nil {
			b.logger.Printf("Error indexing article batch: %v\n", err)
		}
	}
}

func (b Bleve) batchDelete(articles []content.Article) {
	if len(articles) == 0 {
		return
	}

	batch := b.index.NewBatch()
	count := int64(0)

	for i := range articles {
		data := articles[i].Data()

		b.logger.Debugf("Indexing article '%d' from feed id '%d'\n", data.Id, data.FeedId)

		batch.Delete(strconv.FormatInt(int64(data.Id), 10))
		count++

		if count >= b.batchSize {
			if err := b.index.Batch(batch); err != nil {
				b.logger.Printf("Error indexing article batch: %v\n", err)
			}
			batch = b.index.NewBatch()
			count = 0
		}
	}

	if count > 0 {
		if err := b.index.Batch(batch); err != nil {
			b.logger.Printf("Error indexing article batch: %v\n", err)
		}
	}
}

func prepareArticle(data data.Article) (string, indexArticle) {
	id := strconv.FormatInt(int64(data.FeedId), 10)
	ia := indexArticle{FeedId: id,
		ArticleId:   strconv.FormatInt(int64(data.Id), 10),
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
