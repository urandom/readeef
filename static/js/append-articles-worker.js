importScripts("/dist/moment/min/moment.min.js");

self.addEventListener('message', function(event) {
    "use strict";

    var articles = event.data.current || [],
        newArticles = event.data.newArticles,
        feeds = event.data.feeds,
        articleMap = {}, feedMap;

    for (var i = 0, a; a = articles[i]; ++i) {
        delete a.First;
        delete a.Last;

        articleMap[a.Id] = a;
    }

    for (var i = 0, a, pre; a = newArticles[i]; ++i) {
        if (!articleMap[a.Id]) {
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

            a.RelativeDate = moment(a.Date).fromNow();

            if (feeds && feeds.length) {
                if (!feedMap) {
                    feedMap = {};
                    for (var j = 0, f; f = feeds[j]; ++j) {
                        feedMap[f.Id] = f;
                    }
                }

                a.FeedOrigin = feedMap[a.FeedId].Title;
            }
            articles.push(a);
        }
    }

    articles[0].First = true;
    articles[articles.length - 1].Last = true;

    self.postMessage({articles: articles});
});
