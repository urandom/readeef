package readeef

import (
	"bytes"
	"fmt"
	"html"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/search"
)

type SearchIndex struct {
	index    bleve.Index
	logger   *log.Logger
	db       DB
	newIndex bool
}

type SearchResult struct {
	Article
	Hit search.DocumentMatch
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

	return si, nil
}

func (si SearchIndex) IndexAllArticles() {
	Debug.Println("Indexing all articles")

	if articles, err := si.db.GetAllArticles(); err == nil {
		for _, a := range articles {
			if err := si.Index(a); err != nil {
				si.logger.Printf("Error indexing article %s from feed %d: %v\n", a.Id, a.FeedId, err)
			}
		}
	} else {
		si.logger.Printf("Error getting all articles: %v\n", err)
	}
}

func (si SearchIndex) IsNewIndex() bool {
	return si.newIndex
}

func (si SearchIndex) UpdateFeed(feed Feed) {
	Debug.Printf("Updating article search index for feed '%s'\n", feed.Link)
	for _, a := range feed.Articles {
		if feed.lastUpdatedArticleIds[a.Id] {
			if err := si.Index(a); err != nil {
				si.logger.Printf(
					"Error indexing article %s from feed %d: %v\n",
					a.Id, a.FeedId, err)
			}
		}
	}
}

func (si SearchIndex) Index(a Article) error {
	a.Title = html.UnescapeString(stripTags(a.Title))
	a.Description = html.UnescapeString(stripTags(a.Description))
	return si.index.Index(fmt.Sprintf("%d:%s", a.FeedId, a.Id), a)
}

func (si SearchIndex) Delete(a Article) error {
	return si.index.Delete(fmt.Sprintf("%d:%s", a.FeedId, a.Id))
}

func (si SearchIndex) Search(term, highlight string, paging ...int) ([]SearchResult, error) {
	query := bleve.NewQueryStringQuery(term)
	searchRequest := bleve.NewSearchRequest(query)

	if highlight != "" {
		searchRequest.Highlight = bleve.NewHighlightWithStyle(highlight)
	}

	limit, offset := pagingLimit(paging)
	searchRequest.Size = limit
	searchRequest.From = offset

	Debug.Printf("Searching for '%s'\n", query.Query)
	searchResult, err := si.index.Search(searchRequest)

	if err != nil {
		return []SearchResult{}, err
	}

	if len(searchResult.Hits) == 0 {
		Debug.Printf("No results found for '%s'\n", query.Query)
		return []SearchResult{}, nil
	}

	feedArticleIds := []FeedArticleIds{}
	hitMap := map[string]*search.DocumentMatch{}

	for _, hit := range searchResult.Hits {
		ids := strings.SplitN(hit.ID, ":", 2)
		feedId, err := strconv.ParseInt(ids[0], 10, 64)
		if err == nil {
			feedArticleIds = append(feedArticleIds, FeedArticleIds{feedId, ids[1]})
			hitMap[hit.ID] = hit
		}
	}

	articles, err := si.db.GetSpecificArticles(feedArticleIds...)

	searchResults := make([]SearchResult, len(articles))
	for i, article := range articles {
		hit := hitMap[fmt.Sprintf("%d:%s", article.FeedId, article.Id)]
		searchResults[i] = SearchResult{article, *hit}
	}

	return searchResults, err
}

func stripTags(text string) string {
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
