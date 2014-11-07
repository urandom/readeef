package readeef

import (
	"bytes"
	"html"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/search"
)

type SearchIndex struct {
	index     bleve.Index
	logger    *log.Logger
	db        DB
	newIndex  bool
	batchSize int64
	verbose   int
}

type SearchResult struct {
	Article
	Hit search.DocumentMatch
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

func NewSearchIndex(config Config, db DB, logger *log.Logger) (SearchIndex, error) {
	var err error
	var index bleve.Index

	si := SearchIndex{}

	_, err = os.Stat(config.SearchIndex.BlevePath)
	if err == nil {
		Debug.Println("Opening search index " + config.SearchIndex.BlevePath)
		index, err = bleve.Open(config.SearchIndex.BlevePath)

		if err != nil {
			return EmptySearchIndex, err
		}
	} else if os.IsNotExist(err) {
		mapping := bleve.NewIndexMapping()
		docMapping := bleve.NewDocumentMapping()

		idfieldmapping := bleve.NewTextFieldMapping()
		idfieldmapping.IncludeInAll = false
		docMapping.AddFieldMappingsAt("FeedId", idfieldmapping)
		docMapping.AddFieldMappingsAt("ArticleId", idfieldmapping)

		mapping.AddDocumentMapping(mapping.DefaultType, docMapping)

		Debug.Println("Creating search index " + config.SearchIndex.BlevePath)
		index, err = bleve.New(config.SearchIndex.BlevePath, mapping)

		if err != nil {
			return EmptySearchIndex, err
		}

		si.newIndex = true
	} else {
		return EmptySearchIndex, err
	}

	si.logger = logger
	si.db = db
	si.index = index
	si.batchSize = config.SearchIndex.BatchSize

	return si, nil
}

func (si *SearchIndex) SetVerbose(verbose int) {
	si.verbose = verbose
}

func (si SearchIndex) IndexAllArticles() error {
	Debug.Println("Indexing all articles")

	if articles, err := si.db.GetAllArticles(); err == nil {
		si.batchIndex(articles)
	} else {
		return err
	}

	return nil
}

func (si SearchIndex) IsNewIndex() bool {
	return si.newIndex
}

func (si SearchIndex) UpdateFeed(feed Feed) {
	Debug.Printf("Updating article search index for feed '%s'\n", feed.Link)

	var articles []Article
	for _, a := range feed.Articles {
		if feed.lastUpdatedArticleLinks[a.Link] {
			articles = append(articles, a)
		}
	}

	si.batchIndex(articles)
}

func (si SearchIndex) DeleteFeed(feed Feed) error {
	articles, err := si.db.GetAllFeedArticles(feed)

	if err == nil {
		Debug.Printf("Removing all articles from the search index for feed '%s'\n", feed.Link)

		si.batchDelete(articles)
	} else {
		return err
	}
	return nil
}

func (si SearchIndex) batchIndex(articles []Article) {
	if len(articles) == 0 {
		return
	}

	batch := bleve.NewBatch()
	count := int64(0)

	for i, l := 0, len(articles); i < l; i++ {
		a := articles[i]

		if si.verbose > 0 {
			Debug.Printf("Indexing article '%d' from feed id '%d'\n", a.Id, a.FeedId)
		}

		batch.Index(prepareArticle(a))
		count++

		if count >= si.batchSize {
			if err := si.index.Batch(batch); err != nil {
				si.logger.Printf("Error indexing article batch: %v\n", err)
			}
			batch = bleve.NewBatch()
			count = 0
		}
	}

	if count > 0 {
		if err := si.index.Batch(batch); err != nil {
			si.logger.Printf("Error indexing article batch: %v\n", err)
		}
	}
}

func (si SearchIndex) batchDelete(articles []Article) {
	if len(articles) == 0 {
		return
	}

	batch := bleve.NewBatch()
	count := int64(0)

	for i, l := 0, len(articles); i < l; i++ {
		a := articles[i]

		if si.verbose > 0 {
			Debug.Printf("Indexing article '%d' from feed id '%d'\n", a.Id, a.FeedId)
		}

		batch.Delete(strconv.FormatInt(a.Id, 10))
		count++

		if count >= si.batchSize {
			if err := si.index.Batch(batch); err != nil {
				si.logger.Printf("Error indexing article batch: %v\n", err)
			}
			batch = bleve.NewBatch()
			count = 0
		}
	}

	if count > 0 {
		if err := si.index.Batch(batch); err != nil {
			si.logger.Printf("Error indexing article batch: %v\n", err)
		}
	}
}

func (si SearchIndex) Search(u User, term, highlight string, feedIds []int64, paging ...int) ([]SearchResult, error) {
	var query bleve.Query

	query = bleve.NewQueryStringQuery(term)

	if len(feedIds) > 0 {
		queries := make([]bleve.Query, len(feedIds))
		conjunct := make([]bleve.Query, 2)

		for i, id := range feedIds {
			q := bleve.NewTermQuery(strconv.FormatInt(id, 10))
			q.SetField("FeedId")

			queries[i] = q
		}

		disjunct := bleve.NewDisjunctionQuery(queries)

		conjunct[0] = query
		conjunct[1] = disjunct

		Debug.Printf("Constructing query for term '%s' and feed ids [%v]\n", term, feedIds)
		query = bleve.NewConjunctionQuery(conjunct)
	}

	searchRequest := bleve.NewSearchRequest(query)

	if highlight != "" {
		searchRequest.Highlight = bleve.NewHighlightWithStyle(highlight)
	}

	limit, offset := pagingLimit(paging)
	searchRequest.Size = limit
	searchRequest.From = offset

	Debug.Printf("Searching for '%s'\n", term)
	searchResult, err := si.index.Search(searchRequest)

	if err != nil {
		return []SearchResult{}, err
	}

	if len(searchResult.Hits) == 0 {
		Debug.Printf("No results found for '%s'\n", term)
		return []SearchResult{}, nil
	}

	articleIds := []int64{}
	hitMap := map[string]*search.DocumentMatch{}

	for _, hit := range searchResult.Hits {
		if articleId, err := strconv.ParseInt(hit.ID, 10, 64); err == nil {
			articleIds = append(articleIds, articleId)
			hitMap[hit.ID] = hit
		}
	}

	articles, err := si.db.GetSpecificUserArticles(u, articleIds...)

	searchResults := []SearchResult{}
	for _, article := range articles {
		hit := hitMap[strconv.FormatInt(article.Id, 10)]
		searchResults = append(searchResults, SearchResult{article, *hit})
	}

	return searchResults, err
}

func prepareArticle(a Article) (string, indexArticle) {
	ia := indexArticle{FeedId: strconv.FormatInt(a.FeedId, 10),
		ArticleId:   strconv.FormatInt(a.Id, 10),
		Title:       html.UnescapeString(StripTags(a.Title)),
		Description: html.UnescapeString(StripTags(a.Description)),
		Link:        a.Link, Date: a.Date,
	}

	return strconv.FormatInt(a.Id, 10), ia
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
