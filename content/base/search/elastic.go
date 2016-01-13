package search

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"gopkg.in/olivere/elastic.v3"

	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/base"
	"github.com/urandom/readeef/content/data"
	"github.com/urandom/webfw"
)

const (
	elasticIndexName   = "readeef"
	elasticArticleType = "article"
)

type Elastic struct {
	base.ArticleSorting
	client    *elastic.Client
	logger    webfw.Logger
	newIndex  bool
	batchSize int64
}

func NewElastic(url string, size int64, logger webfw.Logger) (content.SearchProvider, error) {
	var client *elastic.Client
	var exists bool

	client, err := elastic.NewClient(elastic.SetURL(url))
	if err != nil {
		return nil, fmt.Errorf("Cannot connect to elastic server '%s': %v", url, err)
	}

	if exists, err = client.IndexExists(elasticIndexName).Do(); err != nil {
		return nil, err
	} else if !exists {
		if _, err = client.CreateIndex(elasticIndexName).Do(); err != nil {
			return nil, err
		}
	}

	return &Elastic{client: client, logger: logger, batchSize: size, newIndex: !exists}, nil
}

func (e Elastic) IsNewIndex() bool {
	return e.newIndex
}

func (e Elastic) IndexAllFeeds(repo content.Repo) error {
	e.logger.Infoln("Indexing all articles")

	for _, f := range repo.AllFeeds() {
		articles := f.AllArticles()
		if f.HasErr() {
			return f.Err()
		}

		if err := e.BatchIndex(articles, data.BatchAdd); err != nil {
			return err
		}
	}

	return repo.Err()
}

func (e Elastic) Search(
	term string, u content.User, feedIds []data.FeedId, limit, offset int,
) (ua []content.UserArticle, err error) {
	search := e.client.Search().Index(elasticIndexName)

	var query elastic.Query

	if t, err := url.QueryUnescape(term); err == nil {
		term = t
	}
	query = elastic.NewCommonTermsQuery("_all", term)

	if len(feedIds) > 0 {
		idFilter := elastic.NewBoolQuery()

		for _, id := range feedIds {
			idFilter = idFilter.Should(elastic.NewTermQuery("feed_id", int64(id)))
		}

		query = elastic.NewBoolQuery().Must(query).Filter(idFilter)
	}

	search.Query(query)
	search.Highlight(elastic.NewHighlight().PreTags("<mark>").PostTags("</mark>").Field("title").Field("description"))
	search.From(offset).Size(limit)

	switch e.Field() {
	case data.SortByDate:
		search.Sort("date", e.Order() == data.AscendingOrder)
	case data.SortById, data.DefaultSort:
		search.Sort("article_id", e.Order() == data.AscendingOrder)
	}

	var res *elastic.SearchResult
	res, err = search.Do()

	if err != nil {
		return
	}

	if res.TotalHits() == 0 {
		return
	}

	articleIds := []data.ArticleId{}
	highlightMap := map[data.ArticleId]elastic.SearchHitHighlight{}

	if res.Hits != nil && res.Hits.Hits != nil {
		for _, hit := range res.Hits.Hits {
			a := indexArticle{}
			if err := json.Unmarshal(*hit.Source, &a); err == nil {
				if id, err := strconv.ParseInt(a.ArticleId, 10, 64); err == nil {
					articleId := data.ArticleId(id)
					articleIds = append(articleIds, articleId)
					highlightMap[articleId] = hit.Highlight
				}
			}
		}
	}

	ua = u.ArticlesById(articleIds)
	if u.HasErr() {
		return ua, u.Err()
	}

	for i := range ua {
		data := ua[i].Data()

		if highlight, ok := highlightMap[data.Id]; ok {
			data.Hit.Fragments = map[string][]string{}
			if len(highlight["title"]) > 0 {
				data.Hit.Fragments["Title"] = highlight["title"]
			}
			if len(highlight["description"]) > 0 {
				data.Hit.Fragments["Description"] = highlight["description"]
			}
			ua[i].Data(data)
		}
	}

	return
}

func (e Elastic) BatchIndex(articles []content.Article, op data.IndexOperation) error {
	if len(articles) == 0 {
		return nil
	}

	bulk := e.client.Bulk()
	count := int64(0)

	for i := range articles {
		d := articles[i].Data()

		var req elastic.BulkableRequest
		switch op {
		case data.BatchAdd:
			e.logger.Debugf("Indexing article '%d' of feed id '%d'\n", d.Id, d.FeedId)

			id, doc := prepareArticle(d)
			req = elastic.NewBulkIndexRequest().Index(elasticIndexName).Type(elasticArticleType).Id(id).Doc(doc)
		case data.BatchDelete:
			e.logger.Debugf("Removing article '%d' of feed id '%d' from the index\n", d.Id, d.FeedId)

			req = elastic.NewBulkDeleteRequest().Index(elasticIndexName).Type(elasticArticleType).Id(strconv.FormatInt(int64(d.Id), 10))
		default:
			return fmt.Errorf("Unknown operation type %v", op)
		}

		bulk.Add(req)
		count++

		if count >= e.batchSize {
			if _, err := bulk.Do(); err != nil {
				return fmt.Errorf("Error indexing article batch: %v\n", err)
			}
			bulk = e.client.Bulk()
			count = 0
		}
	}

	if count > 0 {
		if _, err := bulk.Do(); err != nil {
			return fmt.Errorf("Error indexing article batch: %v\n", err)
		}
	}

	return nil
}
