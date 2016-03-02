(function() {
    function generateSignature(uri, method, body, contentType, date, nonce, secret) {
        var message, bodyHash;

        method = method ? method.toUpperCase() : 'GET';

        if (body != null) {
            body = body.toString();
        }
        bodyHash = CryptoJS.MD5(body).toString(CryptoJS.enc.Base64);

        message = uri + "\n" + method + "\n"
                + bodyHash + "\n" + contentType + "\n"
                + date + "\n" + nonce + "\n";

        return CryptoJS.HmacSHA256(message, secret).toString(CryptoJS.enc.Base64);
    }

    Polymer({
        is: 'rf-connection',
        properties: {
			user: {
				type: Object,
				observer: '_userChanged',
			},
            nonceUrl: String,
            websocketUrl: String,
        },
        listeners: {
            'nonce.response': 'onRequestResponse',
            'nonce.error': 'onRequestError'
        },
		behaviors: [
			ConnectionManagerBehavior,
		],

        retryTimeout: 10,
		initializing: false,
		disconnected: false,
		heartbeatMisses: 0,
		heartbeatInterval: null,
		websocket: null,

		ready: function() {
			if (!this.nonceUrl) {
				throw new Error("No nonce url provided");
			}

			if (!this.websocketUrl) {
				throw new Error("No websocket url provided");
			}

			this.instances = [];
			this.receivers = [];
			this.requestPool = [];
			this.messagePool = [];
		},

		registerAPI: function(api) {
			this.instances.push(api);
			if (api.receiver) {
				this.receivers.push({instance: api, method: api.method, tag: api.tag});
			}
		},

        send: function(api, data) {
            if (!this.websocket || this.websocket.readyState != WebSocket.OPEN) {
                this.messagePool.push({instance: api, data: data});

                this._init();
                return;
            }

            this.requestPool.push({instance: api, method: api.method, tag: api.tag});

            var payload = {method: api.method, tag: api.tag},
                arguments = {}, hasArguments = false;

            if (api.args) {
                hasArguments = true;
                Polymer.Base.mixin(arguments, api.args)
            }

            if (data) {
                hasArguments = true;
                Polymer.Base.mixin(arguments, data)
            }

            if (hasArguments) {
                payload.arguments = arguments;
            }

            this.websocket.send(JSON.stringify(payload));
        },

        close: function() {
            if (this.websocket != null) {
                this._close();
            }
        },

        onRequestResponse: function(event) {
            if (!event.detail.response || !this.user) {
                this.initializing = false;
                if (this.websocket != null) {
                    this._close();
                }
                this.fire('rf-api-error', 'closed');
                return;
            }

            var date = new Date().getTime(),
                nonce = event.detail.response.Nonce,
                proto = 'ws';

            signature = generateSignature(encodeURI(this.websocketUrl), 'GET', null,
                "", date, nonce, this.user.MD5API || "");

            if (location.protocol == "https:") {
                proto = 'wss';
            }

            this.websocket = new WebSocket(
                proto + "://" + location.host + this.websocketUrl +
                "?login=" + encodeURIComponent(this.user.Login) +
                "&signature=" + encodeURIComponent(signature) +
                "&date=" + encodeURIComponent(date) +
                "&nonce=" + encodeURIComponent(nonce));

            this.websocket.onopen = function() {
                this.initializing = false;
                while (this.messagePool.length) {
                    var m = this.messagePool.shift();

					this.send(m.instance, m.data);
                }

                if (this.heartbeatInterval === null) {
                    this.heartbeatMisses = 0
                    this.heartbeatInterval = setInterval(function() {
                        try {
                            this.heartbeatMisses++;
                            if (this.heartbeatMisses >= 3) {
                                throw new Error("Too many missed hearbeats.");
                            }

                            if (this.websocket && this.websocket.readyState == WebSocket.OPEN) {
                                this.websocket.send(JSON.stringify({method: "heartbeat"}));
                            } else {
                                throw new Error("WebSocket is closed");
                            }
                        } catch (e) {
                            this._close();
                        }
					}.bind(this), 30000);
                }

				if (this.disconnected) {
					this.disconnected = false;
					for (var i = 0, inst; inst = this.instances[i]; ++i) {
						inst.fire('rf-api-reconnect');
					}
				}
			}.bind(this);

            this.websocket.onmessage = function(event) {
                var data = event.data;
                try {
                    data = JSON.parse(data)
                } catch(e) {}

                if (data.method == "heartbeat") {
                    this.heartbeatMisses = 0;
                    return;
                }

                for (var i = 0, r; r = this.receivers[i]; ++i) {
                    if (this._dispatchData(r, data)) {
                        return;
                    }
                }

                for (var i = 0, r; r = this.requestPool[i]; ++i) {
                    if (this._dispatchData(r, data)) {
                        this.requestPool.splice(i, 1)
                        return;
                    }
                }
			}.bind(this);

            this.websocket.onclose = function(event) {
				this.disconnected = true;
				this.async(function() {
					this.initializing = false;
                    this._init();
                }, this.retryTimeout * 1000);
            }.bind(this);
        },

        onRequestError: function(event, detail) {
            this.fire('rf-api-error', detail);

			this.disconnected = true;
			this.async(function() {
				this.initializing = false;
				this._init();
			}, this.retryTimeout * 1000);
        },

		_userChanged: function(user, old) {
			if (user && old && user.Login == old.Login && user.MD5API == old.MD5API) {
				return
			}

			if (this.websocket != null) {
				this._close();
			}

			if (this.user) {
				this._init();
			}
		},

        _init: function() {
            if (!this.user || this.initializing) {
                return;
            }

            if (this.websocket != null && this.websocket.readyState == WebSocket.OPEN) {
                return;
            }

            this.initializing = true;

            if (this.websocket != null) {
                this._close();
            }

			return this.$.nonce.generateRequest();
        },

		_close: function() {
			clearInterval(this.heartbeatInterval);
			this.heartbeatInterval = null;
			this.websocket.close();
			this.websocket = null;
		},

		_dispatchData: function(r, data) {
			if (r.method == data.method && r.tag == data.tag) {
				if (data.error) {
					if (data.errorType == "error-unauthorized") {
						r.instance.async(function() {
							Polymer.dom(document).querySelector('rf-router').connectionUnauthorized();
						});
					} else {
						if (r.instance.handlesErrors) {
							r.instance.fire('rf-api-error', data);
						} else {
							Polymer.dom(document).querySelector('rf-router').unhandledAPIError(data);
						}
					}
				} else {
					r.instance.async(function() {
						r.instance.fire('rf-api-message', data);
					});
				}
				return true;
			}

			return false;
		},

    });
})();
