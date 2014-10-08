(function() {
    "use strict";

    var urlParser = document.createElement('a');

    Polymer('rf-settings', {
        selectedTab: 'general',
        loading: false,
        removedFeed: null,
        taggedFeed: null,
        removedUser: null,
        addedUser: null,
        toggleActiveUser: null,
        toggleActiveUserState: false,
        users: null,

        g: CoreStyle.g,

        attached: function() {
            this.cleanFields();
            this.users = [];
        },

        displayChanged: function(oldValue, newValue) {
            if (newValue == 'settings') {
                if (this.user.Admin) {
                    this.$['user-list'].go();
                }
            } else {
                this.cleanFields();
            }
        },

        domain: function(value) {
            urlParser.href = value;

            return urlParser.host;
        },

        splitTags: function(value) {
            return value ? value.join(", ") : value;
        },

        onAddFeedUrlKeypress: function(event, detail, sender) {
            var code = event.keyCode || event.charCode, key = event.keyIdentifier;

            if (key == 'Enter' || code == 13) {
                sender.blur();

                if (!this.url) {
                    this.$['feed-url'].required = true;
                    return;
                }

                this.$['find-feeds'].asyncFire('tap');
            }
        },

        onFindFeed: function() {
            if (this.$.opml.files.length) {
                var file = this.$.opml.files[0], fileReader = new FileReader();

                fileReader.onload = function(event) {
                    var contents = event.target.result;

                    this.$['discover-opml'].body = contents;
                    this.$['discover-opml'].contentType = file.type;

                    this.$['discover-opml'].go();
                }.bind(this);

                fileReader.readAsText(file);
            } else {
                if (!this.url) {
                    return;
                }

                if (window.google && google.feeds) {
                    google.feeds.findFeeds(this.url, function(response) {
                        if (response.status.code == 200) {
                            var feeds = [], urls = {};

                            for (var i = 0, e; e = response.entries[i]; ++i) {
                                if (!urls[e.url]) {
                                    feeds.push({
                                        Link: e.url,
                                        Title: this.stripTags(e.title),
                                        Description: this.stripTags(e.contentSnippet)
                                    });
                                    urls[e.url] = true;
                                }
                            }

                            feeds[0].selected = true;
                            this.onDiscoverFeedResponse(null, {response: {Feeds: feeds, skipSelection: true}});
                        } else {
                            this.onDiscoverFeedError();
                        }
                    }.bind(this));
                } else {
                    this.$['discover-feed'].params = JSON.stringify({"url": this.url});
                    this.$['discover-feed'].go();
                }
            }
            this.loading = true;
        },

        onAddFeed: function() {
            var params = {url: []};
            for (var i = 0, f; f = this.discoveredFeeds[i]; ++i) {
                if (f.selected) {
                    params.url.push(f.Link);
                }
            }

            if (!params.url.length) {
                /* TODO: show that nothing was selected */
                return;
            }

            this.$['add-feed'].params = JSON.stringify(params)
            this.$['add-feed'].go();
            this.loading = true;
        },

        onDiscoverFeedResponse: function(event, data) {
            if (data.response) {
                if (data.response.Feeds) {
                    if (!data.response.skipSelection) {
                        data.response.Feeds.forEach(function(f) {
                            f.selected = true;
                        });
                    }
                } else if (data.response.Error) {
                    this.$['feed-url'].error = this.$['feed-url'].getAttribute("data-" + data.response.ErrorType);
                    this.$['feed-url'].invalid = true;
                }
                this.discoveredFeeds = data.response.Feeds;
            } else {
                this.discoveredFeeds = [];
            }
            this.loading = false;
        },

        onDiscoverFeedError: function(event) {
            this.$['feed-url'].error = this.$['feed-url'].getAttribute("data-error-internal");
            this.$['feed-url'].invalid = true;
            this.loading = false;
        },

        onAddFeedResponse: function(event, data) {
            if (data.response && data.response.Success) {
                this.fire('core-signal', {name: 'rf-feeds-added'});
            }

            this.cleanFields();
        },

        onRemoveFeed: function(event, detail, sender) {
            this.removedFeed = sender.templateInstance.model.feed.Id;
            this.$['remove-feed'].go();
        },

        onRemoveFeedResponse: function(event, data) {
            if (data.response && data.response.Success) {
                this.fire('core-signal', {name: 'rf-feeds-removed'});
            }
        },

        onChangePassword: function() {
            this.$['set-password-dialog'].toggle();
        },

        onApplyPasswordChange: function() {
            var invalid = false;
            ["password", "new-password", "confirm-new-password"].forEach(function(id) {
                if (!this.$[id].value) {
                    this.$[id].required = true;
                    this.$[id].value = null;
                    invalid = true;
                }
            }.bind(this));

            if (this.$["new-password"].value != this.$["confirm-new-password"].value) {
                this.$["confirm-new-password"].invalid = true;
                this.$["confirm-new-password"].error = "Make sure the new password fields match.";
                invalid = true;
            }

            if (invalid) {
                return;
            }

            this.$['password-change'].body = JSON.stringify({
                "Current": this.$.password.value,
                "New": this.$["new-password"].value
            });
            this.$['password-change'].go();
        },

        onPasswordChangeResponse: function(event, data) {
            if (data.response && data.response.Success) {
                this.user = null;
            }
        },

        onPasswordDialogKeypress: function(event) {
            var code = event.keyCode || event.charCode, key = event.keyIdentifier;

            if (key == 'Enter' || code == 13) {
                if (event.target === this.$.password) { 
                    this.$["new-password"].focusAction();
                } else if (event.target === this.$["new-password"]) {
                    this.$["confirm-new-password"].focusAction();
                } else if (event.target === this.$["confirm-new-password"]) {
                    this.$["apply-password-change"].asyncFire('tap');
                }
            }
        },

        onThemeSelect: function(event, detail, sender) {
            var theme = sender.className.replace(/^theme /, '');

            this.settings.theme = theme;
        },

        cleanFields: function() {
            this.url = "";
            this.discoveredFeeds = null;
            this.loading = false;

            ["password", "new-password", "confirm-new-password", "add-user-login", "add-user-password"].forEach(function(id) {
                this.$[id].required = false;
                this.$[id].invalid = false;
                this.$[id].value = "";
                this.$[id].error = "";
            }.bind(this));
        },

        onUserListResponse: function(event, data) {
            if (data.response && data.response.Users) {
                this.users = data.response.Users.filter(function(user) {
                    return user.Login != this.user.Login;
                }.bind(this));
            }
        },

        onCreateUser: function() {
            this.$['add-user-dialog'].toggle();
        },

        onNewUserDialogKeypress: function(event) {
            var code = event.keyCode || event.charCode, key = event.keyIdentifier;

            if (key == 'Enter' || code == 13) {
                if (event.target === this.$['add-user-login']) { 
                    this.$['add-user-password'].focusAction();
                } else if (event.target === this.$['add-user-password']) {
                    this.$['add-user-apply'].asyncFire('tap');
                }
            }
        },

        onApplyCreateUser: function() {
            var invalid = false;
            ["add-user-login", "add-user-password"].forEach(function(id) {
                if (!this.$[id].value) {
                    this.$[id].required = true;
                    this.$[id].value = null;
                    invalid = true;
                }
            }.bind(this));

            if (invalid) {
                return;
            }

            this.addedUser = this.$['add-user-login'].value;
            this.$['user-add'].body = this.$['add-user-password'].value;
            this.$['user-add'].go();
        },

        onUserAddResponse: function(event, data) {
            if (data.response && data.response.Success) {
                this.$['user-list'].go();
            }

            this.$['add-user-dialog'].toggle();
        },

        onRemoveUser: function(event, detail, sender) {
            this.removedUser = sender.templateInstance.model.user.Login;
            this.$['user-remove'].go();
        },

        onUserRemoveResponse: function(event, data) {
            if (data.response && data.response.Success) {
                this.users = this.users.filter(function(user) {
                    return user.Login != data.response.Login;
                });
            }
        },

        onToggleActiveUser: function(event, detail, sender) {
            this.toggleActiveUser = sender.templateInstance.model.user.Login;
            this.toggleActiveUserState = sender.checked;
            this.$['user-toggle-active'].go();
        },

        onUserToggleActiveResponse: function(event, data) {
            if (!data.response.Success) {
                this.users = this.users.map(function(user) {
                    if (user.Login == data.response.Login) {
                        user.Active = !user.Active;
                    }
                    return user;
                });
            }
        },

        onFeedTagsChange: function(event, detail, sender) {
            if (typeof sender.value != "string") {
                return;
            }

            var tags = sender.value.split(/\s*,\s*/);

            sender.templateInstance.model.feed.Tags = tags;

            this.taggedFeed = sender.templateInstance.model.feed.Id;
            this.$['feed-tags'].body = JSON.stringify(tags)
            this.$['feed-tags'].go();
        },

        onFeedTagsResponse: function(event, data) {
            if (data.response && data.response.Success) {
                var feed = this.feeds.filter(function(feed) {
                    if (feed.Id == data.response.Id) {
                        return feed;
                    }
                });
                this.fire('core-signal', {name: 'rf-feed-tags-changed', data: feed});
            }
        },

        onDisplayFeedErrors: function(event, detail, sender) {
            sender.parentNode.querySelector('paper-toast').toggle();
        },

        stripTags: function(html) {
            var div = document.createElement("div");
            div.innerHTML = html;

            return div.textContent || "";
        },

        shareServiceEnabled: function(id) {
            if (!this.settings || !this.settings.shareServices) {
                return false;
            }

            return this.settings.shareServices.indexOf(id) != -1;
        },

        onShareServiceCheckChange: function() {
            this.settings.shareServices = this.$['share-services'].querySelectorAll(
                'paper-toggle-button[data-service-id]'
            ).array().filter(function(e) {
                return e.checked;
            }).map(function(e) {
                return e.getAttribute('data-service-id');
            });
        }
    });
})();
