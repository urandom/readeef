package readeef

import (
	"bytes"
	"errors"
	"fmt"
	"html"
	"os"
	"strconv"
	"time"

	"github.com/blevesearch/bleve"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/data"
	"github.com/urandom/webfw"
)

type SearchIndex struct {
	Index bleve.Index

	logger    webfw.Logger
	repo      content.Repo
	newIndex  bool
	batchSize int64
}

type indexArticle struct {
	FeedId      string
	ArticleId   string
	Title       string
	Description string
	Link        string
	Date        time.Time
}

var EmptySearchIndex = SearchIndex{}

func NewSearchIndex(repo content.Repo, config Config, logger webfw.Logger) (SearchIndex, error) {
	var err error
	var index bleve.Index

	si := SearchIndex{}

	_, err = os.Stat(config.SearchIndex.BlevePath)
	if err == nil {
		logger.Infoln("Opening search index " + config.SearchIndex.BlevePath)
		index, err = bleve.Open(config.SearchIndex.BlevePath)

		if err != nil {
			return EmptySearchIndex, errors.New(fmt.Sprintf("Error opening search index: %v\n", err))
		}
	} else if os.IsNotExist(err) {
		mapping := bleve.NewIndexMapping()
		docMapping := bleve.NewDocumentMapping()

		idfieldmapping := bleve.NewTextFieldMapping()
		idfieldmapping.IncludeInAll = false
		docMapping.AddFieldMappingsAt("FeedId", idfieldmapping)
		docMapping.AddFieldMappingsAt("ArticleId", idfieldmapping)

		mapping.AddDocumentMapping(mapping.DefaultType, docMapping)

		logger.Infoln("Creating search index " + config.SearchIndex.BlevePath)
		index, err = bleve.New(config.SearchIndex.BlevePath, mapping)

		if err != nil {
			return EmptySearchIndex, errors.New(fmt.Sprintf("Error creating search index: %v\n", err))
		}

		si.newIndex = true
	} else {
		return EmptySearchIndex, errors.New(
			fmt.Sprintf("Error getting stat of '%s': %v\n", config.SearchIndex.BlevePath, err))
	}

	si.logger = logger
	si.repo = repo
	si.Index = index
	si.batchSize = config.SearchIndex.BatchSize

	return si, nil
}

func (si SearchIndex) IndexAllArticles() error {
	si.logger.Infoln("Indexing all articles")

	for _, f := range si.repo.AllFeeds() {
		articles := f.AllArticles()
		if f.HasErr() {
			return f.Err()
		}

		si.batchIndex(articles)
	}

	return si.repo.Err()
}

func (si SearchIndex) IsNewIndex() bool {
	return si.newIndex
}

func (si SearchIndex) UpdateFeed(feed content.Feed) {
	si.logger.Infof("Updating article search index for feed '%s'\n", feed)

	newArticleLinks := map[string]bool{}
	for _, a := range feed.NewArticles() {
		newArticleLinks[a.Data().Link] = true
	}

	var articles []content.Article
	for _, a := range feed.ParsedArticles() {
		if newArticleLinks[a.Data().Link] {
			articles = append(articles, a)
		}
	}

	si.batchIndex(articles)
}

func (si SearchIndex) DeleteFeed(feed content.Feed) error {
	articles := feed.AllArticles()

	if !feed.HasErr() {
		si.logger.Infof("Removing all articles from the search index for feed '%s'\n", feed)

		si.batchDelete(articles)
	} else {
		return feed.Err()
	}
	return nil
}

func (si SearchIndex) batchIndex(articles []content.Article) {
	if len(articles) == 0 {
		return
	}

	batch := bleve.NewBatch()
	count := int64(0)

	for i := range articles {
		data := articles[i].Data()

		si.logger.Debugf("Indexing article '%d' from feed id '%d'\n", data.Id, data.FeedId)

		batch.Index(prepareArticle(data))
		count++

		if count >= si.batchSize {
			if err := si.Index.Batch(batch); err != nil {
				si.logger.Printf("Error indexing article batch: %v\n", err)
			}
			batch = bleve.NewBatch()
			count = 0
		}
	}

	if count > 0 {
		if err := si.Index.Batch(batch); err != nil {
			si.logger.Printf("Error indexing article batch: %v\n", err)
		}
	}
}

func (si SearchIndex) batchDelete(articles []content.Article) {
	if len(articles) == 0 {
		return
	}

	batch := bleve.NewBatch()
	count := int64(0)

	for i := range articles {
		data := articles[i].Data()

		si.logger.Debugf("Indexing article '%d' from feed id '%d'\n", data.Id, data.FeedId)

		batch.Delete(strconv.FormatInt(int64(data.Id), 10))
		count++

		if count >= si.batchSize {
			if err := si.Index.Batch(batch); err != nil {
				si.logger.Printf("Error indexing article batch: %v\n", err)
			}
			batch = bleve.NewBatch()
			count = 0
		}
	}

	if count > 0 {
		if err := si.Index.Batch(batch); err != nil {
			si.logger.Printf("Error indexing article batch: %v\n", err)
		}
	}
}

func prepareArticle(data data.Article) (string, indexArticle) {
	ia := indexArticle{FeedId: strconv.FormatInt(int64(data.FeedId), 10),
		ArticleId:   strconv.FormatInt(int64(data.Id), 10),
		Title:       html.UnescapeString(StripTags(data.Title)),
		Description: html.UnescapeString(StripTags(data.Description)),
		Link:        data.Link, Date: data.Date,
	}

	return strconv.FormatInt(int64(data.Id), 10), ia
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
