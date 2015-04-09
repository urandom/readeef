(function() {
    "use strict";

    var urlParser = document.createElement('a');

    function shareServicesList(value) {
        var services = [], serviceCategories = [];

        for (var service in value) {
            services.push(value[service]);
        }

        services.sort(function(a, b) {
            if (a.category && !b.category) {
                return -1;
            } else if (!a.category && b.category) {
                return 1;
            }

            var cat = a.category.localeCompare(b.category);

            if (cat == 0) {
                return a.title.localeCompare(b.title)
            } else {
                return cat;
            }
        });

        for (var i = 0, s, c; s = services[i]; ++i) {
            if (c != s.category) {
                c = s.category;
                serviceCategories.push({name: c, services: []});
            }

            serviceCategories[serviceCategories.length - 1].services.push(s);
        }

        return serviceCategories;
    }

    Polymer('rf-settings', {
        loading: false,
        users: null,

        g: CoreStyle.g,

        attached: function() {
            this.cleanFields();
            this.users = [];
            this.shareServiceList = [];
        },

        displayChanged: function(oldValue, newValue) {
            if (newValue == 'settings') {
                if (this.user.Admin) {
                    this.$['list-users'].send();
                }
                if ("/web/settings/" + this.settingsTab != location.pathname) {
                    history.pushState(null, null, "/web/settings/" + this.settingsTab);
                }
            } else {
                this.cleanFields();
            }
        },

        settingsTabChanged: function(oldValue, newValue) {
            if (oldValue && newValue && "/web/settings/" + this.settingsTab != location.pathname) {
                history.pushState(null, null, "/web/settings/" + this.settingsTab);
            }
        },

        domain: function(value) {
            urlParser.href = value;

            return urlParser.host;
        },

        splitTags: function(value) {
            return value ? value.join(", ") : value;
        },

        shareServicesChanged: function() {
            if (this.shareServices && Object.keys(this.shareServices).length && !this.shareServiceList.length) {
                this.shareServiceList = shareServicesList(this.shareServices);
            }
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

                    this.$['parse-opml'].send({opml: contents});
                }.bind(this);

                fileReader.readAsText(file);
            } else {
                if (!this.url) {
                    return;
                }

                if (window.google && google.feeds && !/https?:\/\//.test(this.url)) {
                    google.feeds.findFeeds(this.url, function(response) {
                        if (response.status.code == 200) {
                            if (response.entries.length) {
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
                                this.onDiscoverFeedsMessage(null, {success: true, arguments: {Feeds: feeds, SkipSelection: true}});
                            } else {
                                this.onDiscoverFeedsError();
                            }
                        } else {
                            this.onDiscoverFeedsError();
                        }
                    }.bind(this));
                } else {
                    this.$['discover-feeds'].send({link: this.url});
                }
            }
            this.loading = true;
        },

        onManageFeedsBack: function() {
            if (this.noSelectedFeeds) {
                this.noSelectedFeeds = false;
            } else if (this.addFeedError) {
                this.addFeedError = null;
            } else {
                this.discoveredFeeds = null;
            }
        },

        onAddFeed: function() {
            var links = [];
            for (var i = 0, f; f = this.discoveredFeeds[i]; ++i) {
                if (f.selected) {
                    links.push(f.Link);
                }
            }

            if (!links.length) {
                this.noSelectedFeeds = true;
                return;
            }

            this.$['add-feed'].send({links: links});
            this.loading = true;
        },

        onDiscoverFeedsMessage: function(event, data) {
            if (!data.arguments.SkipSelection) {
                data.arguments.Feeds.forEach(function(f) {
                    f.selected = true;
                });
            }
            this.discoveredFeeds = data.arguments.Feeds;
            this.loading = false;
        },

        onDiscoverFeedsError: function(event, data) {
            this.$['feed-url'].error = this.$['feed-url'].getAttribute("data-" + data.arguments.ErrorType);
            this.$['feed-url'].invalid = true;
            this.loading = false;
        },

        onAddFeedMessage: function(event, data) {
            this.fire('core-signal', {name: 'rf-feeds-added'});

            this.cleanFields();
        },

        onAddFeedError: function(event, data) {
            this.addFeedError = data.error;
            this.loading = false;
        },

        onRemoveFeed: function(event, detail, sender) {
            this.$['remove-feed'].send({id: sender.templateInstance.model.feed.Id});
        },

        onRemoveFeedMessage: function(event, data) {
            this.fire('core-signal', {name: 'rf-feeds-removed'});
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

            this.$['password-change'].send({
                attribute: "Password",
                value: {
                    "Current": this.$.password.value,
                    "New": this.$["new-password"].value
                }
            });
        },

        onPasswordChangeMessage: function(event, data) {
            this.user = null;
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

        onListUsersMessage: function(event, data) {
            this.users = data.arguments.Users.filter(function(user) {
                return user.Login != this.user.Login;
            }.bind(this));
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

            this.$['add-user'].send({
                login: this.$['add-user-login'].value,
                password: this.$['add-user-password'].value
            });
        },

        onAddUserMessage: function(event, data) {
            this.$['list-users'].send();

            this.$['add-user-dialog'].toggle();
        },

        onRemoveUser: function(event, detail, sender) {
            this.$['remove-user'].send({login: sender.templateInstance.model.user.Login});
        },

        onUserRemoveResponse: function(event, data) {
            this.users = this.users.filter(function(user) {
                return user.Login != data.arguments.Login;
            });
        },

        onToggleActiveUser: function(event, detail, sender) {
            this.$['user-toggle-active'].send({
                login: sender.templateInstance.model.user.Login,
                attribute: "Active",
                value: sender.checked.toString()
            });
        },

        onUserToggleActiveResponse: function(event, data) {
            this.users = this.users.map(function(user) {
                if (user.Login == data.arguments.Login) {
                    user.Active = !user.Active;
                }
                return user;
            });
        },

        onFeedTagsChange: function(event, detail, sender) {
            if (typeof sender.value != "string") {
                return;
            }

            var tags = sender.value.split(/\s*,\s*/);

            sender.templateInstance.model.feed.Tags = tags;

            this.$['set-feed-tags'].send({id: sender.templateInstance.model.feed.Id, tags: tags});
        },

        onSetFeedTagsMessage: function(event, data) {
            var feed = this.feeds.filter(function(feed) {
                if (feed.Id == data.arguments.Id) {
                    return feed;
                }
            });
            this.fire('core-signal', {name: 'rf-feed-tags-changed', data: feed});
        },

        onDisplayFeedErrors: function(event, detail, sender) {
            sender.parentNode.querySelector('paper-toast').toggle();
        },

        stripTags: function(html) {
            var div = document.createElement("div");
            div.innerHTML = html;

            return div.textContent || "";
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
