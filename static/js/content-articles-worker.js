self.addEventListener('message', function(event) {
    "use strict";
    var articles = [];

    for (var i = 0, a; a = event.data[i]; ++i) {
        a.ShortDescription = a.Description.replace(/<!--.*?-->/g, '').replace(/<\w[^>]*>/g, '').replace(/<\/[^>]*>/g, '').trim().replace(/\s\s+/g, ' ');
        articles.push(a);
    }

    self.postMessage({articles: articles});
});
