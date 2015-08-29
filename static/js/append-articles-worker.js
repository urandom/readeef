importScripts("/dist/moment/min/moment.min.js");

self.addEventListener('message', function(event) {
    "use strict";

    var articles = event.data.current || [],
        inserts = [], insertIndex = 0,
        newArticles = event.data.newArticles,
        newerFirst = event.data.newerFirst,
        unreadOnly = event.data.unreadOnly,
        feeds = event.data.feeds,
        articleMap = {}, indexMap = {}, feedMap;

    for (var i = 0, a; a = articles[i]; ++i) {
        articleMap[a.Id] = a;
    }

    for (var i = 0, a, pre; a = newArticles[i]; ++i) {
        if (!articleMap[a.Id]) {
            if (feeds && feeds.length) {
                if (!feedMap) {
                    feedMap = {};
                    for (var j = 0, f; f = feeds[j]; ++j) {
                        feedMap[f.Id] = f;
                    }
                }

                var feed = feedMap[a.FeedId];
                if (!feed) {
                    continue;
                }
                a.FeedOrigin = feedMap[a.FeedId].Title;
            }

            a.Description = a.Description.replace(/<!--.*?-->/g, '')
                .replace(/<script.*?<\/script\s*>/g, '')
                .replace(/(<a )(.*?<\/a\s*>)/g, '$1 target="feed-article" $2');

            pre = "";
            while (pre != a.Description) {
                if (pre) {
                    a.Description = pre;
                }

                pre = a.Description.replace(/(<\w[^>]*?["'])javascript:([^>]*>)/g, '$1$2')
                    .replace(/(<\w[^>]*?)\s+on\w+=(['"]).*?\2([^>]*>)/g, '$1$3');
            }

            a.ShortDescription = a.Description
                .replace(/<\w[^>]*>/g, '').replace(/<\/[^>]*>/g, '').trim().replace(/\s\s+/g, ' ');

            a.Date = new Date(a.Date);
            a.RelativeDate = moment(a.Date).fromNow();

            for (var o; o = articles[insertIndex]; ++insertIndex) {
                if (newerFirst) {
                    if (o.Date <= a.Date) {
                        break;
                    }
                } else {
                    if (o.Date > a.Date) {
                        break;
                    }
                }
            }

            if (!inserts[inserts.length - 1] || inserts[inserts.length - 1].index != insertIndex) {
                inserts.push({index: insertIndex, articles: []});
            }

            inserts[inserts.length - 1].articles.push(a);
        }
    }

    for (var i = 0, insert; insert = inserts[i]; ++i) {
        articles.splice.apply(articles, [insert.index, 0].concat(insert.articles));
    }

    for (var i = 0, a; a = articles[i]; ++i) {
        indexMap[a.Id] = i;
    }

    self.postMessage({inserts: inserts, indexMap: indexMap});
});
