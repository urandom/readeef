package search

import (
	"encoding/json"
	"net/url"
	"strconv"

	"gopkg.in/olivere/elastic.v3"

	"github.com/pkg/errors"
	"github.com/urandom/readeef"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/repo"
)

const (
	elasticIndexName   = "readeef"
	elasticArticleType = "article"
)

type elasticSearch struct {
	client    *elastic.Client
	log       readeef.Logger
	newIndex  bool
	batchSize int64
	service   repo.Service
}

func NewElastic(url string, size int64, service repo.Service, log readeef.Logger) (elasticSearch, error) {
	var client *elastic.Client
	var exists bool

	client, err := elastic.NewClient(elastic.SetURL(url))
	if err != nil {
		return nil, errors.Wrapf(err, "connecting to elastic server '%s'", url)
	}

	if exists, err = client.IndexExists(elasticIndexName).Do(); err != nil {
		return nil, err
	} else if !exists {
		if _, err = client.CreateIndex(elasticIndexName).Do(); err != nil {
			return nil, errors.Wrap(err, "creating index")
		}
	}

	return &elasticSearch{client: client, log: log, batchSize: size, service: service, newIndex: !exists}, nil
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

	search := e.client.Search().Index(elasticIndexName)

	var query elastic.Query

	if t, err := url.QueryUnescape(term); err == nil {
		term = t
	}
	query = elastic.NewCommonTermsQuery("_all", term)

	feedIDs := o.FeedIDs

	if len(feedIDs) == 0 {
		var err error
		if feedIDs, err = b.service.FeedRepo().IDs(); err != nil {
			return []content.Article{}, errors.WithMessage(err, "getting feed ids")
		} else if len(feedIDs) == 0 {
			return []content.Article{}, nil
		}
	}

	idFilter := elastic.NewBoolQuery()

	for _, id := range feedIDs {
		idFilter = idFilter.Should(elastic.NewTermQuery("feed_id", int64(id)))
	}

	query = elastic.NewBoolQuery().Must(query).Filter(idFilter)

	search.Query(query)
	search.Highlight(elastic.NewHighlight().PreTags("<mark>").PostTags("</mark>").Field("title").Field("description"))
	search.From(o.Offset).Size(o.Limit)

	switch o.SortField {
	case content.SortByDate:
		search.Sort("date", o.SortOrder == content.AscendingOrder)
	case content.SortById, content.DefaultSort:
		search.Sort("article_id", o.SortOrder == content.AscendingOrder)
	}

	var res *elastic.SearchResult
	res, err = search.Do()

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
				if id, err := strconv.ParseInt(a.ArticleID, 10, 64); err == nil {
					articleID := content.ArticleId(id)
					articleIDs = append(articleIDs, articleID)
					highlightMap[articleID] = hit.Highlight
				}
			}
		}
	}

	articles, err := b.service.ArticleRepo().All(content.IDs(articleIDs))
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
			e.log.Debugf("Indexing article '%d' of feed id '%d'\n", a.Id, a.FeedId)

			id, doc := prepareArticle(a)
			req = elastic.NewBulkIndexRequest().Index(elasticIndexName).Type(elasticArticleType).Id(id).Doc(doc)
		case BatchDelete:
			e.log.Debugf("Removing article '%d' of feed id '%d' from the index\n", a.Id, a.FeedId)

			req = elastic.NewBulkDeleteRequest().Index(elasticIndexName).Type(elasticArticleType).Id(strconv.FormatInt(int64(a.Id), 10))
		default:
			return errors.Errorf("unknown operation type %v", op)
		}

		bulk.Add(req)
		count++

		if count >= e.batchSize {
			if _, err := bulk.Do(); err != nil {
				return errors.Wrap(err, "indexing article batch")
			}
			bulk = e.client.Bulk()
			count = 0
		}
	}

	if count > 0 {
		if _, err := bulk.Do(); err != nil {
			return errors.Wrap(err, "indexing article batch")
		}
	}

	return nil
}
