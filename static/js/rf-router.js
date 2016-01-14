(function(root) {
    "use strict";

    var userTTL = 1000 * 60 * 60 * 24 * 15,
        state = {VALIDATING: 1 << 0};

    function randomTheme() {
        if (document.body.classList.contains('theme-__random__')) {
            var classes = ['blue', 'indigo', 'cyan', 'teal', 'green', 'light-green', 'lime', 'red', 'pink', 'purple', 'deep-purple', 'yellow', 'amber', 'deep-orange', 'grep'],
                index = Math.floor(Math.random() * classes.length - 1);

            for (var i = 0, c; c = document.body.classList[i]; ++i) {
                if (c != "theme-__random__" && c.indexOf('theme-') == 0) {
                    document.body.classList.remove(c);
                    break;
                }
            }

            document.body.classList.add('theme-' + classes[index]);
            Polymer.updateStyles();
        }
    };
    setInterval(randomTheme, 1800000);

	root.RouteBehavior = {
		attached: function() {
			Polymer.dom(this.root).querySelectorAll('excess-route').forEach(function(route) {
				if (route.routeAlias) {
					var meta = document.createElement('iron-meta');
					meta.key = route.routeAlias;
					meta.value = route;
				}
			});
		},

		urlFor: function(spec, params) {
			return Excess.RouteManager.getRoutePath(spec, params);
		},

		namedRoute: function(name) {
			return document.createElement('iron-meta').byKey(name);
		},

		isRouteActive: function(name) {
			var r = this.namedRoute(name);
			return r && r.active;
		},
	};

    Polymer({
        is: "rf-router",
        properties: {
            route: {
                type: String
            },
            user: {
                type: Object,
                readOnly: true,
                notify: true
			},
			topLevelNavigation: {
				type: String,
				observer: '_topLevelNavigationChanged',
			},
			loginRedirect: {
				type: String,
			},
        },
		behaviors: [
			RouteBehavior,
		],
		_routingStarted: false,
        _state: 0,

        attached: function() {
			Excess.RouteManager.start();
			this._routingStarted = true;

            this.async(function() {
                if (!this.user && (this._state & state.VALIDATING) != state.VALIDATING) {
                    if (this.topLevelNavigation == "splash" || !location.pathname) {
						Excess.RouteManager.transitionTo('@login');
                    } else if (this.topLevelNavigation != "login") {
						Excess.RouteManager.transitionTo('@login-from', {url: this.encodeURI(location.pathname)});
                    }
                }
            });

            document.addEventListener('rf-lazy-insert', function(event) {
                Polymer.updateStyles();
            }.bind(this));
        },

        onUserLoad: function(event, detail) {
            var storage = event.target;

            if (storage.value) {
                if (!storage.value.authTime || new Date().getTime() - storage.value.authTime > this.userTTL) {
                    storage.value = null;
                }
            }

            if (!detail.externalChange) {
                this.validateUser(storage.value);
            }
        },

        validateUser: function(user) {
            if (!user) {
                return;
            }

            this._state |= state.VALIDATING;

            var authCheck = this.$['auth-check'];
            var validateMessage = function(event) {
                authCheck.removeEventListener('rf-api-message', validateMessage);
                if (!event.detail.arguments.Auth) {
                    return this.connectionUnauthorized();
                }

                var user = event.detail.arguments.User;
                user.authTime = new Date().getTime();
                user.capabilities = event.detail.arguments.Capabilities;

                user.ProfileData = user.ProfileData || {};

                if (('language' in user.ProfileData) && user.ProfileData.language != this.dataset.language) {
                    location.href = location.href.replace('/' + this.dataset.language + '/', '/' + user.ProfileData.language + '/');
                }

                this._setUser(user);
                this._state &= ~state.VALIDATING;

                if (user.ProfileData.theme) {
                    document.body.classList.add('theme-' + user.ProfileData.theme);
                    randomTheme();
                }

                if (user.ProfileData.shareServices) {
                    user.ProfileData.shareServices.forEach(function(name) {
                        RfShareServices.get(name).active = true;
                    });
                }

				this.debounce('validate-redirect', this.validateRedirect);
            }.bind(this);

            authCheck.user = user;
            authCheck.addEventListener('rf-api-message', validateMessage);
            authCheck.send();
        },

		validateRedirect: function() {
			if (!this._routingStarted) {
				this.debounce('validate-redirect', this.validateRedirect, 50);
				return;
			}

			if (this.topLevelNavigation == "login" && this.loginRedirect) {
				var login = Polymer.dom(this.root).querySelector('rf-login');
				if (login) {
					login.hide();
				}

				try {
					Excess.RouteManager.transitionTo(this.decodeURI(this.loginRedirect));
				} catch(e) {
					Excess.RouteManager.transitionTo('@feed-all');
				}
			} else if (this.topLevelNavigation == "login") {
				var login = Polymer.dom(this.root).querySelector('rf-login');
				if (login) {
					login.hide();
				}
				Excess.RouteManager.transitionTo('@feed-all');
			} else if (this.topLevelNavigation != "feed" &&
					this.topLevelNavigation != "settings") {
				Excess.RouteManager.transitionTo('@feed-all');
			}
		},

        logout: function() {
            this.$.logout.send();
            this.$['auth-check'].close();
            this._setUser(null);
            this.async(function() {
				Excess.RouteManager.transitionTo('@login');
            });
        },

        connectionUnauthorized: function() {
            this._state &= ~state.VALIDATING;
			switch (this.topLevelNavigation) {
			case "feed":
			case "settings":
				Excess.RouteManager.transitionTo('@login-from', {url: this.encodeURI(location.pathname)});
				break;
			case "login":
                var login = Polymer.dom(this.root).querySelector('rf-login');
                if (login) {
                    login.invalid = true;
                }
				break;
			default:
				Excess.RouteManager.transitionTo('@login');
				break;
			}
        },

        unhandledAPIError: function(data) {
            this.$['api-error'].text = "Error: " + JSON.stringify(data.error) + ", type: " + data.errorType;
            this.$['api-error'].show();
        },

        encodeURI: function(uri) {
            return encodeURIComponent(uri).replace(/%/g, '$');
        },

        decodeURI: function(encodedURI) {
            return decodeURIComponent(encodedURI.replace(/\$/g, '%'));
        },

		_topLevelNavigationChanged: function(value, old) {
			if (value == "logout") {
				this.logout();
                return;
			}

            if (!this.user && (this._state & state.VALIDATING) != state.VALIDATING && value != "login") {
				if (this.topLevelNavigation == "splash" || !location.pathname) {
					Excess.RouteManager.transitionTo('@login');
				} else if (this.topLevelNavigation != "login") {
					Excess.RouteManager.transitionTo('@login-from', {url: this.encodeURI(location.pathname)});
				}
			}

			if (value == "login") {
                var login = Polymer.dom(this.root).querySelector('rf-login');
                if (login) {
                    login.show();
                }
			}
		},
    });

    root.NestedRouteBehavior = {
        defaultNestedRoute: function(parentName, nestedName, nestedParams) {
            if (!MoreRouting.isCurrentUrl(nestedName)) {
                MoreRouting.navigateTo(nestedName, nestedParams || {});
            }

            MoreRouting.getRouteByName(parentName).__subscribe(function(name, value) {
                if (name == "active") {
                    // Change the route with async, otherwise once it finishes,
                    // the current one will continue and will revert it
                    this.async(function() {
                        if (value) {
                            if (!MoreRouting.isCurrentUrl(nestedName)) {
                                var route = MoreRouting.getRoute(nestedName), toNotify = {};
                                for (var key in nestedParams) {
                                    if (route.params[key] === nestedParams[key]) {
                                        toNotify[key] = nestedParams[key];
                                    }
                                }

                                MoreRouting.navigateTo(MoreRouting.urlFor(nestedName, nestedParams || {}));

                                for (var key in toNotify) {
                                    route.params.__notify(key, toNotify[key]);
                                }
                            }
                        }
                    }.bind(this));
                }
            }.bind(this));
        },

        routeParamObserver: function(routeName, param, cb) {
            var route = MoreRouting.getRoute(routeName), debouncer;

            route.params.__subscribe(function(name, value) {
                if (name == param) {
                    debouncer = Polymer.Debounce(debouncer, cb.bind(this, value));
                }
            });

            route.__subscribe(function(key, value) {
                if (key == 'active' && route.params[param] !== undefined) {
                    if (value) {
                        debouncer = Polymer.Debounce(debouncer, cb.bind(this, route.params[param]));
                    } else {
                        debouncer = Polymer.Debounce(debouncer, cb.bind(this, null, true));
                    }
                }
            });

            if (route.active) {
                debouncer = Polymer.Debounce(debouncer, cb.bind(this, route.params[param]));
            }
        },

    };

})(window);
