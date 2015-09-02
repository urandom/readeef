package search

import (
	"net/url"
	"reflect"
	"strconv"

	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/base"
	"github.com/urandom/readeef/content/data"
	"github.com/urandom/webfw"
	"gopkg.in/olivere/elastic.v2"
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

func NewElastic(url string, size int64, logger webfw.Logger) (e *Elastic, err error) {
	var client *elastic.Client
	var exists bool

	client, err = elastic.NewClient(elastic.SetURL(url))
	if err != nil {
		return
	}

	if exists, err = client.IndexExists(elasticIndexName).Do(); err != nil {
		return
	} else if !exists {
		if _, err = client.CreateIndex(elasticIndexName).Do(); err != nil {
			return
		}
	}

	e = &Elastic{client: client, logger: logger, batchSize: size}
	return
}

func (e Elastic) IsNewIndex() bool {
	return e.newIndex
}

func (e Elastic) IndexAllArticles(repo content.Repo) error {
	e.logger.Infoln("Indexing all articles")

	for _, f := range repo.AllFeeds() {
		articles := f.AllArticles()
		if f.HasErr() {
			return f.Err()
		}

		e.batchIndex(articles)
	}

	return repo.Err()
}

func (e Elastic) UpdateFeed(feed content.Feed) {
	e.logger.Infof("Updating article search index for feed '%s'\n", feed)

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

	e.batchIndex(articles)
}

func (e Elastic) DeleteFeed(feed content.Feed) error {
	articles := feed.AllArticles()

	if !feed.HasErr() {
		e.logger.Infof("Removing all articles from the search index for feed '%s'\n", feed)

		e.batchDelete(articles)
	} else {
		return feed.Err()
	}
	return nil
}

func (e Elastic) Search(
	term string, u content.User, feedIds []data.FeedId, limit, offset int,
) (ua []content.UserArticle, err error) {
	search := e.client.Search().Index(elasticIndexName)

	var query elastic.Query

	if t, err := url.QueryUnescape(term); err == nil {
		term = t
	}
	query = elastic.NewCommonQuery("_all", term)

	if len(feedIds) > 0 {
		idFilter := elastic.NewOrFilter()

		for _, id := range feedIds {
			idFilter = idFilter.Add(elastic.NewTermFilter("feed_id", int64(id)))
		}

		filteredQuery := elastic.NewFilteredQuery(query)
		query = filteredQuery.Filter(idFilter)
	}

	search.Query(query)
	search.Highlight(elastic.NewHighlight().PreTags("<mark>").PostTags("</mark>").Field("_all"))
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
	hitMap := map[data.ArticleId]indexArticle{}

	for _, item := range res.Each(reflect.TypeOf(indexArticle{})) {
		if a, ok := item.(indexArticle); ok {
			if id, err := strconv.ParseInt(a.ArticleId, 10, 64); err == nil {
				articleId := data.ArticleId(id)
				articleIds = append(articleIds, articleId)
				hitMap[articleId] = a
			}
		}
	}

	ua = u.ArticlesById(articleIds)
	if u.HasErr() {
		return ua, u.Err()
	}

	for i := range ua {
		data := ua[i].Data()

		if hit, ok := hitMap[data.Id]; ok {
			data.Hit.Fragments = map[string][]string{
				"Title":       []string{hit.Title},
				"Description": []string{hit.Description},
			}
			ua[i].Data(data)
		}
	}

	return
}

func (e Elastic) batchIndex(articles []content.Article) {
	if len(articles) == 0 {
		return
	}

	bulk := e.client.Bulk()
	count := int64(0)

	for i := range articles {
		data := articles[i].Data()

		e.logger.Debugf("Indexing article '%d' from feed id '%d'\n", data.Id, data.FeedId)

		id, doc := prepareArticle(data)
		req := elastic.NewBulkIndexRequest().Index(elasticIndexName).Type(elasticArticleType).Id(id).Doc(doc)
		bulk.Add(req)
		count++

		if count >= e.batchSize {
			if _, err := bulk.Do(); err != nil {
				e.logger.Printf("Error indexing article batch: %v\n", err)
			}
			bulk = e.client.Bulk()
			count = 0
		}
	}

	if count > 0 {
		if _, err := bulk.Do(); err != nil {
			e.logger.Printf("Error indexing article batch: %v\n", err)
		}
	}
}

func (e Elastic) batchDelete(articles []content.Article) {
	if len(articles) == 0 {
		return
	}

	bulk := e.client.Bulk()
	count := int64(0)

	for i := range articles {
		data := articles[i].Data()

		e.logger.Debugf("Indexing article '%d' from feed id '%d'\n", data.Id, data.FeedId)

		req := elastic.NewBulkDeleteRequest().Index(elasticIndexName).Type(elasticArticleType).Id(strconv.FormatInt(int64(data.Id), 10))
		bulk.Add(req)
		count++

		if count >= e.batchSize {
			if _, err := bulk.Do(); err != nil {
				e.logger.Printf("Error indexing article batch: %v\n", err)
			}
			bulk = e.client.Bulk()
			count = 0
		}
	}

	if count > 0 {
		if _, err := bulk.Do(); err != nil {
			e.logger.Printf("Error indexing article batch: %v\n", err)
		}
	}
}
