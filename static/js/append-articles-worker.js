self.addEventListener('message', function(event) {
    "use strict";

    var articles = event.data.current || [],
        newArticles = event.data.newArticles,
        articleMap = {};

    for (var i = 0, a; a = articles[i]; ++i) {
        articleMap[a.Id] = a;
    }

    for (var i = 0, a; a = newArticles[i]; ++i) {
        if (!articleMap[a.Id]) {
            articles.push(a);
        }
    }

    self.postMessage({articles: articles});
});
