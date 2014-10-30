(function() {
    "use strict";

    Polymer('rf-content-list', {
        itemHeight: 48,
        listHeight: 0,
        pathObserver: null,
        listPosition: 0,
        articleRead: false,
        
        ready: function() {
            this.formattingEnabled = this.formattingEnabled == "true";

            document.addEventListener('keypress', this.onContentKeypress.bind(this), false);
        },

        nextArticle: function() {
            if (this.article) {
                var index = this.articles.indexOf(this.article);

                if (index < this.articles.length - 1) {
                    this.article = this.articles[index + 1];
                }
            } else {
                this.article = this.articles[0];
            }
        },

        previousArticle: function() {
            if (this.article) {
                var index = this.articles.indexOf(this.article);

                if (index) {
                    this.article = this.articles[index - 1];
                }
            } else {
                this.article = this.articles[this.articles.length - 1];
            }
        },

        created: function() {
            this.articles = [];
        },

        domReady: function() {
            var contentPanel = document.querySelector('rf-app').$.scaffolding.$['content-panel'];

            this.$['articles-list'].scrollTarget = contentPanel;

            contentPanel.addEventListener('scroll', function(event) {
                if (!this.articles.length || this.$.pages.selected == "detail") {
                    return;
                }

                if (contentPanel.scroller.offsetHeight + contentPanel.scroller.scrollTop + 50 > contentPanel.scroller.scrollHeight) {
                    this.asyncFire('core-signal', {name: 'rf-request-articles'});
                }
            }.bind(this));
        },

        feedChanged: function(oldValue, newValue) {
            this.listHeight = 0;

            this.article = null;

            var processArticles = (function processArticles() {
                if (this.feed.Articles && this.feed.Articles.length) {
                    var worker = new Worker('/js/content-articles-worker.js'),
                        data = {current: this.feed},
                        feedId = this.feed.Id.toString();

                    worker.addEventListener('message', function(event) {
                        this.articles = event.data.articles;
                    }.bind(this));

                    if (feedId.indexOf("tag:") == 0 || feedId.indexOf("search:") == 0 || feedId.indexOf("popular:tag:") == 0 || feedId == '__favorite__' || feedId == 'popular:__all__') {
                        data.feeds = this.feeds;
                    }

                    worker.postMessage(data);
                } else if (this.articles.length) {
                    this.articles = [];
                }
            }).bind(this);

            if (newValue) {
                processArticles();

                if (this.pathObserver) {
                    this.pathObserver.close();
                }

                this.pathObserver = new PathObserver(newValue, 'Articles');
                this.pathObserver.open(function(newValue) {
                    processArticles();
                }.bind(this));
            }
        },

        articleChanged: function(oldValue, newValue) {
            var scroller = document.querySelector('rf-app').$.scaffolding.$['content-panel'].scroller;

            if (newValue) {
                if (!oldValue) {
                    this.listPosition = scroller.scrollTop;
                }
                scroller.scrollTop = 0;

                this.$.pages.selected = 'detail';

                if (!newValue.Read) {
                    this.articleRead = true;
                    this.$['article-read'].go();
                }

                if (newValue.Last) {
                    this.asyncFire('core-signal', {name: 'rf-request-articles'});
                }
            } else {
                this.$.pages.selected = 'list';

                this.async(function() {
                    this.$['articles-list'].refresh(true);
                    scroller.scrollTop = this.listPosition;
                });
            }
        },

        articlesChanged: function(oldValue, newValue) {
            if (newValue && this.article && newValue.indexOf(this.article) == -1) {
                var found = false;
                for (var i = 0, a; a = newValue[i]; ++i) {
                    if (a.Id == this.article.Id) {
                        this.article = a;
                        found = true;
                        break;
                    }
                }

                if (!found) {
                    this.article = null;
                }
            }
        },

        onArticleActivate: function(event, detail) {
            this.article = detail.data;
        },

        onFavoriteToggle: function(event) {
            event.stopPropagation();
            event.preventDefault();

            var articleId = event.target.getAttribute('data-article-id');
            for (var i = 0, a; a = this.articles[i]; ++i) {
                if (a.Id == articleId) {
                    a.Favorite = !a.Favorite;

                    var fav = new RfAPI();
                    fav.pathAction = "article/favorite/" + a.Id + "/" + a.Favorite;
                    fav.user = this.user;
                    fav.method = "post";

                    fav.go()

                    this.$['articles-list'].refresh(true);

                    break;
                }
            }
        },

        onReadArticleToggle: function() {
            this.articleRead = !this.article.Read;
            this.$['article-read'].go();
            this.$['articles-list'].refresh(true);
        },

        onArticleReadResponse: function(event, data) {
            if (data.response && data.response.Success) {
                if (this.article && this.article.Id == data.response.ArticleId) {
                    this.article.Read = data.response.Read;
                } else {
                    for (var i = 0, a; a = this.articles[i]; ++i) {
                        if (a.Id == data.response.ArticleId) {
                            a.Read = data.response.Read;
                            break;
                        }
                    }
                }
            }
        },

        onContentKeypress: function(event) {
            if (this.$.pages.offsetWidth == 0 && this.$.pages.offsetHeight == 0) {
                return;
            }

            var code = event.keyCode || event.charCode, key = event.keyIdentifier;

            if (key == "U+004A" || code == 106 || code == 74) { // j
                this.nextArticle();
            } else if (key == "U+004B" || code == 107 || code == 75) { // k
                this.previousArticle();
            } else if (key == "U+0048" || code == 104 || code == 72) { // h
                this.article = null;
            } else if (key == "U+0056" || code == 118 || code == 86) { // v
                if (this.article) {
                    this.$['article-display'].openCurrentArticle();
                }
            } else if (key == "U+004D" || code == 109 || code == 77) { // m
                if (this.article) {
                    this.onReadArticleToggle();
                }
            } else if (key == "U+0043" || code == 99 || code == 67) { // c
                if (this.article && this.formattingEnabled) {
                    this.fire('core-signal', {name: "rf-article-format"});
                }
            } else if (key == "U+0052" || code == 114 || code == 82) { // r
                this.fire('core-signal', {name: "rf-feed-refresh"});
            } else if (key == "U+0046" || code == 102 || code == 70) { // f
                if (this.article) {
                    this.fire('core-signal', {name: "rf-article-favorite"});
                }
            }
        }
    });
})();
