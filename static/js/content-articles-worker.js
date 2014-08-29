self.addEventListener('message', function(event) {
    "use strict";
    var articles = [], current = event.data.current, feeds = event.data.feeds, feedMap;

    for (var i = 0, a; a = current.Articles[i]; ++i) {
        a.ShortDescription = a.Description.replace(/<!--.*?-->/g, '')
            .replace(/<script.*?<\/script>/, '')
            .replace(/<\w[^>]*>/g, '').replace(/<\/[^>]*>/g, '').trim().replace(/\s\s+/g, ' ');

        if (current.Id == "__all__") {
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

    self.postMessage({articles: articles});
});
