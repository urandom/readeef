(function() {
    "use strict";

    var urlParser = document.createElement('a');

    function createPseudoTagFeed(tag) {
        // TODO: i18n
        return {
            Id: "tag:" + tag,
            Title: 'Articles from ' + tag,
            Description: "",
            Articles: null,
            Image: {},
            Link: "",
        }
    }

    function createAllPseudoFeed() {
        // TODO: i18n
        return {
            Id: "all",
            Title: "All feed articles",
            Description: "",
            Articles: null,
            Image: {},
            Link: "",
        }
    }

    function createFavoritePseudoFeed() {
        // TODO: i18n
        return {
            Id: "favorite",
            Title: "Favorite feed articles",
            Description: "",
            Articles: null,
            Image: {},
            Link: "",
        }
    }

    function createPopularPseudoFeed(tag) {
        // TODO: i18n
        return {
            Id: "popular:" + tag,
            Title: "Popular feed articles",
            Description: "",
            Articles: null,
            Image: {},
            Link: "",
        }
    }

    function createSearchPseudoFeed(query) {
        // TODO: i18n
        return {
            Id: "search:" + query,
            Title: "Search results for '" + query + "'",
            Description: "",
            Articles: null,
            Image: {},
            Link: "",
        }
    }

    Polymer('rf-app', {
        selected: 'splash',
        settingsTab: 'general',
        responsiveWidth: '768px',
        userTTL: 1000 * 60 * 60 * 24 * 15,
        user: null,
        userSettings: null,
        currentFeedId: null,
        currentFeed: null,
        currentArticle: null,
        loadingArticles: false,
        loadingMoreArticles: false,
        noMoreArticles: false,
        display: 'feed',
        limit: 50,
        offset: 0,
        lastUpdateTime: 0,
        userObserver: null,
        userSettingsObserver: null,
        updateAvailable: false,
        lastUpdateNotifyStart: 0,
        requestArticle: null,

        created: function() {
            this.feeds = [];
            this.tags = [];
            this.feedIdMap = {};
            this.shareServices = {};
        },

        attached: function() {
            var shareServices = {};

            for (var element in this.$) {
                if (this.$[element].tagName.toLowerCase() == "rf-share-service") {
                    shareServices[element] = this.$[element];
                }
            }

            this.shareServices = shareServices;

            window.addEventListener("popstate", function(event) {
                var path = location.pathname;
                if (path.indexOf("/web/login") == 0)  {
                    this.user = null;
                } else if (path.indexOf("/web/") == 0) {
                    if (path.indexOf("/web/settings/") == 0) {
                        this.display = "settings";
                        this.settingsTab = path.substring("/web/settings/".length);
                    } else {
                        var displayChanged = false,
                            match = path.match(/^\/web\/feed\/([^\/]+)(?:\/article\/([^\/]+))?/);

                        if (this.display != "feed") {
                            this.display = "feed";
                            displayChanged = true;
                        }

                        if (match) {
                            var id = parseInt(match[2]);

                            this.requestArticle = !isNaN(id) ? { Id: id } : null;
                            if (this.currentFeedId == match[1]) {
                                if (this.requestArticle) {
                                    if (this.currentFeed && this.currentFeed.Articles) {
                                        for (var i = 0, a; a = this.currentFeed.Articles[i]; ++i) {
                                            if (a.Id == this.requestArticle.Id) {
                                                this.currentArticle = a;
                                                this.requestArticle = null;
                                                return;
                                            }
                                        }
                                    }
                                    if (!displayChanged) {
                                        this.updateFeedArticles();
                                    }
                                } else {
                                    this.currentArticle = null;
                                }
                            } else {
                                this.currentFeedId = match[1];
                            }
                        }
                    }
                }
            }.bind(this));
        },

        userChanged: function(oldValue, newValue) {
            this.async(function() {
                if (this.userSettingsObserver) {
                    this.userSettingsObserver.close();
                }

                if (!newValue) {
                    this.selected = 'login';
                    this.feeds = [];
                    this.tags = [];
                    this.userSettings = null;
                    if ("/web/login" != location.pathname) {
                        history.pushState(null, null, "/web/login")
                    }
                } else {
                    if (!oldValue
                        || oldValue.Login != newValue.Login
                        || oldValue.MD5API != newValue.MD5API) {
                        this.$['auth-check'].send();
                    }

                    this.userObserver = new ObjectObserver(this.user);
                    this.userObserver.open(function (added, removed, changed, getOldValueFn) {
                        var api = this.$['user-settings'];
                        Object.keys(changed).forEach(function(attribute) {
                            switch (attribute) {
                            case "FirstName":
                            case "LastName":
                            case "Email":
                                api.send({attribute: attribute, value: changed[attribute]})
                                break;
                            }
                        });
                    }.bind(this));
                }

            })
        },

        userSettingsChanged: function(oldValue, newValue) {
            if (this.userSettingsObserver) {
                this.userSettingsObserver.close();
            }

            if (newValue) {
                this.display = newValue.display || 'feed';
                this.currentFeedId = newValue.currentFeedId;
                CoreStyle.g.theme = newValue.theme || 'blue';

                var updateShareServices = function() {
                    var shareServices = Polymer.mixin({}, this.shareServices);

                    for (var service in shareServices) {
                        shareServices[service].enabled = false;
                    }

                    for (var i = 0, s; s = (this.userSettings.shareServices || [])[i]; ++i) {
                        if (shareServices[s]) {
                            shareServices[s].enabled = !!this.$[s];
                        }
                    }

                    this.shareServices = shareServices;
                }.bind(this);

                this.userSettingsObserver = new ObjectObserver(this.userSettings);
                this.userSettingsObserver.open(function (added, removed, changed, getOldValueFn) {
                    var amalgamation = Polymer.mixin({}, added, removed, changed);
                    if ('newerFirst' in amalgamation || 'unreadOnly' in amalgamation) {
                        this.updateFeedArticles();
                    }

                    if ('theme' in amalgamation) {
                        CoreStyle.g.theme = amalgamation.theme;
                    }

                    if ('shareServices' in amalgamation) {
                        updateShareServices();
                    }

                    this.$['user-settings'].send({
                        attribute: "ProfileData",
                        value: this.userSettings
                    });
                }.bind(this));

                updateShareServices();
            }
        },

        displayChanged: function(oldValue, newValue) {
            this.userSettings.display = newValue;
            if (newValue == "feed") {
                this.updateFeedArticles();
                if ("/web/feed/" + this.currentFeedId != location.pathname) {
                    history.pushState(null, null, "/web/feed/" + this.currentFeedId);
                }
            }
        },

        currentFeedIdChanged: function(oldValue, newValue) {
            if (this.feeds && this.feeds.length) {
                if (newValue == "favorite") {
                    this.currentFeed = createFavoritePseudoFeed();
                } else if (newValue == "all") {
                    this.currentFeed = createAllPseudoFeed();
                } else if (newValue.indexOf("popular:") == 0) {
                    this.currentFeed = createPopularPseudoFeed(newValue.substring(8));
                } else if (newValue.indexOf("tag:") == 0) {
                    this.currentFeed = createPseudoTagFeed(newValue.substring(4));
                } else if (newValue.indexOf("search:") == 0) {
                    this.searchTerm = newValue.substring(7);
                    this.currentFeed = createSearchPseudoFeed(this.searchTerm);

                    if (oldValue.indexOf("search:") != 0) {
                        this.userSettings.preSearchFeedId = oldValue;
                    }
                } else {
                    this.currentFeed = this.feedIdMap[newValue];
                }

                this.userSettings.currentFeedId = newValue;
            }
        },

        currentFeedChanged: function(oldValue, newValue) {
            if (this.display == "feed") {
                this.updateFeedArticles();
                if ("/web/feed/" + this.currentFeedId != location.pathname) {
                    history.pushState(null, null, "/web/feed/" + this.currentFeedId);
                }
            }
        },

        currentArticleChanged: function(oldValue, newValue) {
            if (newValue) {
                if ("/web/feed/" + this.currentFeedId + "/article/" + newValue.Id != location.pathname) {
                    history.pushState(null, null, "/web/feed/" + this.currentFeedId + "/article/" + newValue.Id);
                }
            } else {
                if ("/web/feed/" + this.currentFeedId != location.pathname) {
                    history.pushState(null, null, "/web/feed/" + this.currentFeedId);
                }
            }
        },

        feedsChanged: function(oldValue, newValue) {
            var self = this;

            if (newValue) {
                newValue.forEach(function(feed) {
                    self.feedIdMap[feed.Id] = feed;
                });
            }

            if (!this.currentFeed && this.currentFeedId) {
                this.currentFeedIdChanged(this.currentFeedId, this.currentFeedId);
            }
        },

        domain: function(value) {
            urlParser.href = value;

            return urlParser.host;
        },

        encodeURIComponent: function(value) {
            return encodeURIComponent(value);
        },

        onConnectionUnauthorized: function(event, data) {
            if (this.selected == 'login') {
                this.$.login.invalid = true;
            }
            this.user = null;
        },

        onAuthCheckMessage: function(event, data) {
            var path = location.pathname;
            if (path.indexOf("/web/login") == 0) {
                if (location.search.indexOf("uri=") > -1) {
                    var match = location.search.match(/\buri=([^&]+)/), uri;

                    if (match) {
                        path = decodeURIComponent(match[1]);
                    } else {
                        path = "";
                    }
                } else {
                    path = "";
                }
            }

            var profileData = data.arguments.ProfileData;
            if (path.indexOf("/web/") == 0) {
                if (path.indexOf("/web/settings/") == 0) {
                    profileData.display = "settings";
                    this.settingsTab = path.substring("/web/settings/".length);
                } else {
                    profileData.display = "feed";

                    var match = path.match(/^\/web\/feed\/([^\/]+)(?:\/article\/([^\/]+))?/);
                    if (match) {
                        profileData.currentFeedId = match[1];
                        var id = parseInt(match[2]);

                        this.requestArticle = !isNaN(id) ? { Id: id } : null;
                    }
                }
            }

            this.user.authTime = new Date().getTime();
            this.user.Admin = data.arguments.User.Admin;
            this.user.Email = data.arguments.User.Email;
            this.user.FirstName = data.arguments.User.FirstName;
            this.user.LastName = data.arguments.User.LastName;
            this.userSettings = profileData;

            if (this.selected == 'login' || this.selected == 'splash') {
                this.selected = 'scaffolding';
            }

            this.$['user-storage'].save();

            this.$['list-feeds'].send();
        },

        onUserLoad: function(event, detail, sender) {
            if (sender.value) {
                if (!sender.value.authTime || new Date().getTime() - this.user.authTime > this.userTTL) {
                    sender.value = null;
                }
            }

            if (!sender.value) {
                this.selected = 'login';

                var query = "";
                if (location.pathname) {
                    query += "?uri=" + encodeURIComponent(location.pathname);
                }

                if ("/web/login" + query != location.pathname) {
                    history.pushState(null, null, "/web/login" + query)
                }
            }
        },

        onDisplaySettings: function() {
            this.display = 'settings';
        },

        onSignOut: function() {
            this.user = null;
        },

        onAddFeed: function() {
            this.$['add-feed-dialog'].toggle();
        },

        onFeedsChanged: function() {
            this.$['list-feeds'].send();
        },

        onAllFeedsMessage: function(event, data) {
            this.feeds = data.arguments.Feeds;

            this.updateTags();
        },

        onFeedTap: function(event, detail, sender) {
            if (this.display != 'feed') {
                this.display = 'feed';
            }

            this.currentFeedId = sender.getAttribute('name');
        },

        onFeedRefresh: function(event, detail, sender) {
            this.updateFeedArticles();
        },

        onFeedArticlesMessage: function(event, data) {
            if (data.arguments.Articles && data.arguments.Articles.length) {
                var worker = new Worker('/js/append-articles-worker.js');

                worker.addEventListener('message', function(event) {
                    window.requestAnimationFrame(function() {
                        this.currentFeed.Articles = event.data.articles;
                        this.loadingArticles = false;
                        this.loadingMoreArticles = false;
                    }.bind(this));
                }.bind(this));

                worker.postMessage({
                    current: this.currentFeed.Articles,
                    newArticles: data.arguments.Articles
                });
            } else {
                this.noMoreArticles = true;
                this.loadingArticles = false;
                this.loadingMoreArticles = false;

                if (!this.offset) {
                    window.requestAnimationFrame(function() {
                        this.currentFeed.Articles = null;
                    }.bind(this));
                }
            }
            this.lastUpdateTime = new Date().getTime();
        },

        onRequestArticles: function(event) {
            if (this.loadingMoreArticles || this.noMoreArticles || this.display != 'feed') {
                return;
            }

            this.loadingMoreArticles = true;
            this.offset += this.limit;
            this.$['feed-articles'].send();
        },

        updateFeedArticles: function() {
            if (!this.currentFeed) {
                return;
            }

            this.currentArticle = null;

            this.updateAvailable = false;
            this.noMoreArticles = false;
            this.offset = 0;

            this.loadingArticles = true;

            if (this.requestArticle) {
                var api = new RfAPI();

                api.addEventListener('rf-api-message', function(event) {
                    var article = event.detail.arguments.Article,
                        feedId = this.currentFeed.Id,
                        valid = true;

                    article.First = true;
                    article.Last = true;

                    if (feedId != "all" && (!isNaN(feedId) || feedId.indexOf("popular:") == -1 && feedId.indexOf("search:") == -1)) {
                        if (feedId == "favorite") {
                            if (!article.Favorite) {
                                valid = false;
                            }
                        } else if (!isNaN(feedId)) {
                            if (feedId != article.FeedId) {
                                valid = false;
                            }
                        } else if (feedId.indexOf("tag:") == 0) {
                            var tag = feedId.substring(4), found = false;
                            for (var i = 0, t; t = this.tags[i]; ++i) {
                                if (t.name == tag) {
                                    for (var j = 0, f; f = t.feeds[j]; ++j) {
                                        if (f.Id == article.FeedId) {
                                            found = true;
                                            break;
                                        }
                                    }
                                    break;
                                }
                            }

                            if (!found) {
                                valid = false;
                            }
                        }
                    }

                    if (valid) {
                        this.currentFeed.Articles = [article];
                        this.currentArticle = article;
                    }

                    if (this.currentFeed.Id.toString().indexOf("search:") == 0) {
                        this.noMoreArticles = true;
                        this.$['feed-search'].send();
                    } else {
                        this.$['feed-articles'].send();
                    }

                }.bind(this));

                api.method = "get-article";
                api.send({id: this.requestArticle.Id});

                this.appendChild(api);
            } else {
                if (this.currentFeed.Id.toString().indexOf("search:") == 0) {
                    this.noMoreArticles = true;
                    this.$['feed-search'].send();
                } else {
                    this.$['feed-articles'].send();
                }

                window.requestAnimationFrame(function() {
                    this.currentFeed.Articles = null;
                }.bind(this));
            }

        },

        onMarkAllAsRead: function() {
            this.$['feed-read-all'].send();
        },

        onFeedReadAllMessage: function(event, data) {
            this.updateFeedArticles();
        },

        onTagCollapseToggle: function(event, detail, sender) {
            var tag = sender.getAttribute('data-tag'),
                collapse = this.$.scaffolding.querySelector(
                    'core-collapse[data-tag="' + sender.getAttribute('data-tag') + '"]'
                );

            if (collapse) {
                collapse.toggle();
            }

            event.stopPropagation();
        },

        onFeedTagsChange: function() {
            this.updateTags();
        },

        onFeedSearch: function(event, detail) {
            if (detail === "") {
                this.currentFeedId = this.userSettings.preSearchFeedId || 'all';
            } else {
                this.currentFeedId = "search:" + detail;
            }
        },

        onFeedUpdateNotify: function(event, data) {
            if (!this.user) {
                return;
            }

            if (this.currentFeedId.toString().indexOf("tag:") == 0) {
                var tag = this.getTag(this.currentFeedId.substring(4));

                if (tag.feeds.some(function(feed) { return feed.Id == data.arguments.Feed.Id })) {
                    this.updateAvailable = true;
                }
            } else if (this.currentFeedId == 'all' && this.currentFeedId == data.arguments.Feed.Id) {
                this.updateAvailable = true;
            }
        },

        onApiError: function(event, data) {
            if (this.$.error.opened) {
                this.$.error.dismiss();
            }

            this.$.error.text = "Error: " + JSON.stringify(data.error) + ", type: " + data.errorType
            this.$.error.show()
        },

        updateTags: function() {
            var tagList = [], tags = {};

            this.feeds.forEach(function(feed) {
                if (feed.Tags && feed.Tags.length) {
                    for (var i = 0, tag; tag = feed.Tags[i]; ++i) {
                        if (!tags[tag]) {
                            tags[tag] = [];
                        }

                        tags[tag].push(feed);
                    }
                }
            });

            Object.keys(tags).sort().forEach(function(tag) {
                tagList.push({name: tag, feeds: tags[tag]});
            });

            this.tags = tagList;
        },

        getTag: function(name) {
            for (var i = 0, t; t = this.tags[i]; ++i) {
                if (t.name == name) {
                    return t;
                }
            }
        }

    });
})();
