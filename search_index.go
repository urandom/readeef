package readeef

import (
	"fmt"
	"log"
	"os"

	"github.com/blevesearch/bleve"
)

type SearchIndex struct {
	index  bleve.Index
	logger *log.Logger
}

func NewSearchIndex(config Config, db DB, logger *log.Logger) (SearchIndex, error) {
	si := SearchIndex{logger: logger}

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
