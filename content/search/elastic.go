package search

import (
	"context"
	"encoding/json"
	"net/url"
	"strconv"
	"time"

	elastic "gopkg.in/olivere/elastic.v5"

	"github.com/pkg/errors"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/repo"
	"github.com/urandom/readeef/log"
)

const (
	elasticIndexName   = "readeef"
	elasticArticleType = "article"
)

type elasticSearch struct {
	client    *elastic.Client
	log       log.Log
	newIndex  bool
	batchSize int64
	service   repo.Service
}

func timeout(d time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), d)
}

func NewElastic(url string, size int64, service repo.Service, log log.Log) (elasticSearch, error) {
	var client *elastic.Client
	var exists bool

	client, err := elastic.NewClient(elastic.SetURL(url))
	if err != nil {
		return elasticSearch{}, errors.Wrapf(err, "connecting to elastic server '%s'", url)
	}

	ctx, cancel := timeout(2 * time.Second)
	defer cancel()
	if exists, err = client.IndexExists(elasticIndexName).Do(ctx); err != nil {
		return elasticSearch{}, err
	} else if !exists {
		if _, err = client.CreateIndex(elasticIndexName).Do(ctx); err != nil {
			return elasticSearch{}, errors.Wrap(err, "creating index")
		}
	}

	return elasticSearch{client: client, log: log, batchSize: size, service: service, newIndex: !exists}, nil
}

func (e elasticSearch) IsNewIndex() bool {
	return e.newIndex
}

func (e elasticSearch) Search(
	term string,
	u content.User,
	opts ...content.QueryOpt,
) ([]content.Article, error) {

	o := content.QueryOptions{}
	o.Apply(opts)

	search := e.client.Search(elasticIndexName)

	var query elastic.Query

	if t, err := url.QueryUnescape(term); err == nil {
		term = t
	}
	query = elastic.NewCommonTermsQuery("_all", term)

	feedIDs := o.FeedIDs

	if len(feedIDs) == 0 {
		var err error
		if feedIDs, err = e.service.FeedRepo().IDs(); err != nil {
			return []content.Article{}, errors.WithMessage(err, "getting feed ids")
		} else if len(feedIDs) == 0 {
			return []content.Article{}, nil
		}
	}

	filter := make([]elastic.Query, 1, 3)

	for _, id := range feedIDs {
		filter[0] = elastic.NewBoolQuery().Should(elastic.NewTermQuery("feed_id", int64(id)))
	}

	if !o.BeforeDate.IsZero() || !o.AfterDate.IsZero() {
		filter = append(
			filter,
			elastic.NewRangeQuery("date").Gt(o.AfterDate).Lt(o.BeforeDate),
		)
	}

	if o.BeforeID > 0 || o.AfterID > 0 {
		filter = append(
			filter,
			elastic.NewRangeQuery("id").Gt(o.AfterID).Lt(o.BeforeID),
		)
	}

	query = elastic.NewBoolQuery().Must(query).Filter(filter...)

	search.Query(query)
	search.Highlight(elastic.NewHighlight().PreTags("<mark>").PostTags("</mark>").Field("title").Field("description"))
	search.Size(o.Limit)

	switch o.SortField {
	case content.SortByDate:
		search.Sort("date", o.SortOrder == content.AscendingOrder)
	case content.SortByID, content.DefaultSort:
		search.Sort("article_id", o.SortOrder == content.AscendingOrder)
	}

	ctx, cancel := timeout(2 * time.Second)
	defer cancel()
	res, err := search.Do(ctx)

	if err != nil {
		return []content.Article{}, errors.Wrap(err, "performing search")
	}

	if res.TotalHits() == 0 {
		return []content.Article{}, nil
	}

	articleIDs := []content.ArticleID{}
	highlightMap := map[content.ArticleID]elastic.SearchHitHighlight{}

	if res.Hits != nil && res.Hits.Hits != nil {
		for _, hit := range res.Hits.Hits {
			a := indexArticle{}
			if err := json.Unmarshal(*hit.Source, &a); err == nil {
				articleID := content.ArticleID(a.ArticleID)
				articleIDs = append(articleIDs, articleID)
				highlightMap[articleID] = hit.Highlight
			}
		}
	}

	queryOpts := []content.QueryOpt{
		content.IDs(articleIDs),
		content.Sorting(o.SortField, o.SortOrder),
	}
	if o.UnreadFirst {
		queryOpts = append(queryOpts, content.UnreadFirst)
	}
	if o.UnreadOnly {
		queryOpts = append(queryOpts, content.UnreadOnly)
	}

	articles, err := e.service.ArticleRepo().ForUser(u, queryOpts...)
	if err != nil {
		return []content.Article{}, errors.WithMessage(err, "getting articles by ids")
	}

	for i := range articles {
		if highlight, ok := highlightMap[articles[i].ID]; ok {
			articles[i].Hit.Fragments = map[string][]string{}
			if len(highlight["title"]) > 0 {
				articles[i].Hit.Fragments["Title"] = highlight["title"]
			}
			if len(highlight["description"]) > 0 {
				articles[i].Hit.Fragments["Description"] = highlight["description"]
			}
		}
	}

	return articles, nil
}

func (e elasticSearch) BatchIndex(articles []content.Article, op indexOperation) error {
	if len(articles) == 0 {
		return nil
	}

	bulk := e.client.Bulk()
	count := int64(0)

	for i := range articles {
		a := articles[i]

		var req elastic.BulkableRequest
		switch op {
		case BatchAdd:
			e.log.Debugf("Indexing article %s", a)
			id, doc := prepareArticle(a)
			req = elastic.NewBulkIndexRequest().Index(elasticIndexName).Type(elasticArticleType).Id(id).Doc(doc)
		case BatchDelete:
			e.log.Debugf("Removing article %d of feed id %d from the index", a.ID, a.FeedID)

			req = elastic.NewBulkDeleteRequest().Index(elasticIndexName).Type(elasticArticleType).Id(strconv.FormatInt(int64(a.ID), 10))
		default:
			return errors.Errorf("unknown operation type %v", op)
		}

		bulk.Add(req)
		count++

		ctx, cancel := timeout(time.Duration(count) * time.Second)
		defer cancel()
		if count >= e.batchSize {
			if _, err := bulk.Do(ctx); err != nil {
				return errors.Wrap(err, "indexing article batch")
			}
			bulk = e.client.Bulk()
			count = 0
		}
	}

	if count > 0 {
		ctx, cancel := timeout(time.Duration(count) * time.Second)
		defer cancel()
		if _, err := bulk.Do(ctx); err != nil {
			return errors.Wrap(err, "indexing article batch")
		}
	}

	return nil
}

func (e elasticSearch) RemoveFeed(id content.FeedID) error {
	q := elastic.NewTermQuery("feed_id", int64(id))

	ctx, cancel := timeout(10 * time.Second)
	defer cancel()

	if _, err := e.client.DeleteByQuery(elasticIndexName).Query(q).Do(ctx); err != nil {
		return errors.Wrapf(err, "deleting articles for feed %d", id)
	}

	return nil
}
