package readeef

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/blevesearch/bleve"
)

type SearchIndex struct {
	index  bleve.Index
	logger *log.Logger
	db     DB
}

func NewSearchIndex(config Config, db DB, logger *log.Logger) (SearchIndex, error) {
	si := SearchIndex{logger: logger, db: db}

	_, err := os.Stat(config.SearchIndex.BlevePath)
	if err == nil {
		index, err := bleve.Open(config.SearchIndex.BlevePath)

		if err != nil {
			return si, err
		}

		si.index = index
	} else if os.IsNotExist(err) {
		mapping := bleve.NewIndexMapping()
		index, err := bleve.New(config.SearchIndex.BlevePath, mapping)

		if err != nil {
			return si, err
		}

		si.index = index

		go func() {
			Debug.Println("Indexing all articles")

			if articles, err := db.GetAllArticles(); err == nil {
				for _, a := range articles {
					if err := si.Index(a); err != nil {
						logger.Printf("Error indexing article %s from feed %d: %v\n", a.Id, a.FeedId, err)
					}
				}
			} else {
				logger.Printf("Error getting all articles: %v\n", err)
			}
		}()
	} else {
		return si, err
	}

	return si, nil
}

func (si SearchIndex) UpdateListener(original <-chan Feed) chan Feed {
	updateFeed := make(chan Feed)

	go func() {
		for {
			select {
			case feed := <-original:
				for _, a := range feed.Articles {
					if feed.lastUpdatedArticleIds[a.Id] {
						if err := si.Index(a); err != nil {
							si.logger.Printf(
								"Error indexing article %s from feed %d: %v\n",
								a.Id, a.FeedId, err)
						}
					}
				}

				updateFeed <- feed
			}
		}
	}()

	return updateFeed
}

func (si SearchIndex) Index(a Article) error {
	return si.index.Index(fmt.Sprintf("%d:%s", a.FeedId, a.Id), a)
}

func (si SearchIndex) Search(term string, highlight ...string) ([]Article, error) {
	query := bleve.NewQueryStringQuery("bleve")
	searchRequest := bleve.NewSearchRequest(query)
	if len(highlight) > 0 {
		searchRequest.Highlight = bleve.NewHighlightWithStyle(highlight[0])
	}
	searchResult, err := si.index.Search(searchRequest)

	if err != nil {
		return []Article{}, err
	}

	if len(searchResult.Hits) == 0 {
		return []Article{}, nil
	}

	feedArticleIds := []FeedArticleIds{}
	for _, hit := range searchResult.Hits {
		ids := strings.SplitN(hit.ID, ":", 2)
		feedId, err := strconv.ParseInt(ids[0], 10, 64)
		if err == nil {
			feedArticleIds = append(feedArticleIds, FeedArticleIds{feedId, ids[1], hit.Score})
		}
	}

	return si.db.GetSpecificArticles(feedArticleIds...)
}
