self.addEventListener('message', function(event) {
    "use strict";

    var articles = event.data.current || [],
        newArticles = event.data.newArticles,
        articleMap = {};

    for (var i = 0, a; a = articles[i]; ++i) {
        delete a.First;
        delete a.Last;

        articleMap[a.Id] = a;
    }

    for (var i = 0, a; a = newArticles[i]; ++i) {
        if (!articleMap[a.Id]) {
            articles.push(a);
        }
    }

    articles[0].First = true;
    articles[articles.length - 1].Last = true;

    self.postMessage({articles: articles});
});
