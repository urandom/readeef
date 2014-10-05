(function() {
    "use strict";

    var urlParser = document.createElement('a');

    function createPseudoTagFeed(tag) {
        // TODO: i18n
        return {
            Id: "tag:" + tag,
            Title: tag == '__all__' ? 'All feed articles' : 'Articles from ' + tag,
            Description: "",
            Articles: null,
            Image: {},
            Link: "",
        }
    }

    function createFavoritePseudoFeed() {
        // TODO: i18n
        return {
            Id: "__favorite__",
            Title: "Favorite feed articles",
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
        selected: 'loading',
        responsiveWidth: '768px',
        userTTL: 1000 * 60 * 60 * 24 * 15,
        user: null,
        userSettings: null,
        currentFeedId: null,
        currentFeed: null,
        currentArticle: null,
        loadingArticles: false,
        loadingMoreArticles: false,
        feedIdMap: {},
        noMoreArticles: false,
        display: 'feed',
        limit: 50,
        offset: 0,
        lastUpdateTime: 0,
        userObserver: null,
        userSettingsObserver: null,
        updateAvailable: false,
        lastUpdateNotifyStart: 0,

        created: function() {
            this.feeds = [];
            this.tags = [];
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
                } else {
                    if (!oldValue
                        || oldValue.Login != newValue.Login
                        || oldValue.MD5API != newValue.MD5API) {
                        this.$['auth-check'].go();
                    }

                    this.userObserver = new ObjectObserver(this.user);
                    this.userObserver.open(function (added, removed, changed, getOldValueFn) {
                        var ajax = this.$['user-settings'];
                        Object.keys(changed).forEach(function(attribute) {
                            switch (attribute) {
                            case "FirstName":
                            case "LastName":
                            case "Email":
                                ajax.body = changed[attribute];
                                ajax.pathAction = "user-settings/" + attribute;
                                ajax.go();
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
                this.currentFeedId = newValue.currentFeedId;
                this.display = newValue.display || 'feed';
                CoreStyle.g.theme = newValue.theme || 'blue';

                this.userSettingsObserver = new ObjectObserver(this.userSettings);
                this.userSettingsObserver.open(function (added, removed, changed, getOldValueFn) {
                    var amalgamation = Polymer.extend(Polymer.extend(Polymer.extend({}, added), removed), changed);
                    if ('newerFirst' in amalgamation || 'unreadOnly' in amalgamation) {
                        this.updateFeedArticles();
                    }

                    if ('theme' in amalgamation) {
                        CoreStyle.g.theme = amalgamation.theme;
                    }

                    this.$['user-settings'].body = JSON.stringify(this.userSettings);
                    this.$['user-settings'].pathAction = "user-settings/ProfileData";
                    this.$['user-settings'].go();
                }.bind(this));
            }
        },

        displayChanged: function(oldValue, newValue) {
            this.userSettings.display = newValue;
        },

        currentFeedIdChanged: function(oldValue, newValue) {
            if (this.feeds && this.feeds.length) {
                if (newValue == "__favorite__") {
                    this.currentFeed = createFavoritePseudoFeed();
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
            this.updateFeedArticles();
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

        onAuthCheckComplete: function(event, response) {
            if (response.response == 403) {
                if (this.selected == 'login') {
                    this.$.login.invalid = true;
                }
                this.user = null;
            }
        },

        onAuthCheckResponse: function(event, data) {
            this.user.authTime = new Date().getTime();
            this.user.Admin = data.response.User.Admin;
            this.user.Email = data.response.User.Email;
            this.user.FirstName = data.response.User.FirstName;
            this.user.LastName = data.response.User.LastName;
            this.userSettings = data.response.ProfileData;

            if (this.selected == 'login' || this.selected == 'loading') {
                this.selected = 'scaffolding';
            }

            this.$['user-storage'].save();

            this.$['list-feeds'].go();
            this.$['feed-update-notifier'].go();
        },

        onUserLoad: function(event, detail, sender) {
            if (sender.value) {
                if (!sender.value.authTime || new Date().getTime() - this.user.authTime > this.userTTL) {
                    sender.value = null;
                }
            }

            if (!sender.value) {
                this.selected = 'login';
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
            this.$['list-feeds'].go();
        },

        onAllFeedsResponse: function(event, data) {
            if (data.response) {
                this.feeds = data.response.Feeds;

                this.updateTags();
            }
        },

        onFeedTap: function(event) {
            if (this.display != 'feed') {
                this.display = 'feed';
            }

            this.currentFeedId = event.target.getAttribute('name');
        },

        onFeedRefresh: function(event, detail, sender) {
            this.updateFeedArticles();
        },

        onFeedArticlesResponse: function(event, data) {
            if (data.response) {
                if (data.response.Articles && data.response.Articles.length) {
                    var worker = new Worker('/js/append-articles-worker.js');

                    worker.addEventListener('message', function(event) {
                        this.currentFeed.Articles = event.data.articles;
                        this.loadingArticles = false;
                        this.loadingMoreArticles = false;
                    }.bind(this));

                    worker.postMessage({
                        current: this.currentFeed.Articles,
                        newArticles: data.response.Articles
                    });
                } else {
                    this.noMoreArticles = true;
                    this.loadingArticles = false;
                    this.loadingMoreArticles = false;

                    if (!this.offset) {
                        this.currentFeed.Articles = null;
                    }
                }
                this.lastUpdateTime = new Date().getTime();
            }
        },

        onRequestArticles: function(event) {
            if (this.loadingMoreArticles || this.noMoreArticles || this.display != 'feed') {
                return;
            }

            this.loadingMoreArticles = true;
            this.offset += this.limit;
            this.$['feed-articles'].go();
        },

        updateFeedArticles: function() {
            if (!this.currentFeed) {
                return;
            }

            this.currentArticle = null;
            this.currentFeed.Articles = null;

            this.updateAvailable = false;
            this.noMoreArticles = false;
            this.offset = 0;

            this.loadingArticles = true;

            if (this.currentFeed.Id.toString().indexOf("search:") == 0) {
                this.noMoreArticles = true;
                this.$['feed-search'].go();
            } else {
                this.$['feed-articles'].go();
            }
        },

        onMarkAllAsRead: function() {
            this.$['feed-read-all'].go();
        },

        onFeedReadAllResponse: function(event, data) {
            if (data.response && data.response.Success) {
                this.updateFeedArticles();
            }
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
                this.currentFeedId = this.userSettings.preSearchFeedId || 'tag:__all__';
            } else {
                this.currentFeedId = "search:" + detail;
            }
        },

        onFeedUpdateNotify: function(event, data) {
            if (!this.user) {
                return;
            }

            if (data.response && data.response.Feed) {
                if (this.currentFeedId.toString().indexOf("tag:") == 0) {
                    var currentTag = this.currentFeedId.substring(4);

                    for (var i = 0, tag; tag = this.tags[i]; ++i) {
                        if (tag.name == currentTag) {
                            for (var j = 0, feed; feed = tag.feeds[j]; ++j) {
                                if (feed.Id == data.response.Feed.Id) {
                                    this.updateAvailable = true;
                                    break;
                                }
                            }
                            break;
                        }
                    }

                } else if (this.currentFeedId == data.response.Feed.Id) {
                    this.updateAvailable = true;
                }
            }
        },

        onFeedUpdateNotifyComplete: function() {
            if (this.lastUpdateNotifyStart) {
                if (new Date().getTime() - this.lastUpdateNotifyStart < 1000) {
                    this.job('update-notifier', this.onFeedUpdateNotifyComplete);
                    return;
                }
            }
            this.$['feed-update-notifier'].go();
            this.lastUpdateNotifyStart = new Date().getTime();
        },

        updateTags: function() {
            var tagList = [{name: '__all__', feeds: this.feeds}], tags = {};

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
        }

    });
})();
