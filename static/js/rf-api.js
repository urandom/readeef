(function() {
    var webSocket = null, messagePool = [], requestPool = [], receivers = [], initializing = false,
        heartbeatMisses = 0, heartbeatInterval = null;

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

    function dispatchData(r, data) {
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
    }

    function clearHeartbeat() {
        clearInterval(heartbeatInterval);
        heartbeatInterval = null;
        webSocket.close();
        webSocket = null;
    }

    Polymer({
        is: 'rf-api',
        properties: {
            user: Object,
            version: {
                type: Number,
                value: 1
            },
            url: String,
            method: String,
            args: Object,
            receiver: Boolean,
            tag: {
                type: String,
                value: ""
            },
            handlesErrors: Boolean,
        },
        listeners: {
            'nonce.response': 'onRequestResponse',
            'nonce.error': 'onRequestError'
        },

        retryTimeout: 10,

        attached: function() {
            if (this.receiver) {
                receivers.push({instance: this, method: this.method, tag: this.tag});
            }

            this._init();
        },

        send: function(data) {
            if (!webSocket || webSocket.readyState != WebSocket.OPEN) {
                messagePool.push({instance: this, data: data});

                this._init();
                return;
            }

            requestPool.push({instance: this, method: this.method, tag: this.tag});

            var payload = {method: this.method, tag: this.tag},
                arguments = {}, hasArguments = false;

            if (this.args) {
                hasArguments = true;
                Polymer.Base.mixin(arguments, this.args)
            }

            if (data) {
                hasArguments = true;
                Polymer.Base.mixin(arguments, data)
            }

            if (hasArguments) {
                payload.arguments = arguments;
            }

            webSocket.send(JSON.stringify(payload));
        },

        onRequestResponse: function(event) {
            if (!event.detail.response) {
                initializing = false;
                if (webSocket != null) {
                    clearHeartbeat();
                }
                this.fire('rf-api-error', 'closed');
                return;
            }

            if (!this.url) {
                this.url = this.$.nonce.getAttribute('data-api-pattern') + "v" + this.version + "/";
            }

            var date = new Date().getTime(),
                nonce = event.detail.response.Nonce,
                proto = 'ws';

            signature = generateSignature(encodeURI(this.url), 'GET', null,
                "", date, nonce, this.user.MD5API || "");

            if (location.protocol == "https:") {
                proto = 'wss';
            }

            webSocket = new WebSocket(
                proto + "://" + location.host + this.url +
                "?login=" + encodeURIComponent(this.user.Login) +
                "&signature=" + encodeURIComponent(signature) +
                "&date=" + encodeURIComponent(date) +
                "&nonce=" + encodeURIComponent(nonce));

            webSocket.onopen = function() {
                initializing = false;
                while (messagePool.length) {
                    var m = messagePool.shift();

                    m.instance.send(m.data);
                }

                if (heartbeatInterval === null) {
                    heartbeatMisses = 0
                    heartbeatInterval = setInterval(function() {
                        try {
                            heartbeatMisses++;
                            if (heartbeatMisses >= 3) {
                                throw new Error("Too many missed hearbeats.");
                            }

                            if (webSocket && webSocket.readyState == WebSocket.OPEN) {
                                webSocket.send(JSON.stringify({method: "heartbeat"}));
                            } else {
                                throw new Error("WebSocket is closed");
                            }
                        } catch (e) {
                            clearHeartbeat();
                        }
                    }, 5000);
                }
            };

            webSocket.onmessage = function(event) {
                var data = event.data;
                try {
                    data = JSON.parse(data)
                } catch(e) {}

                if (data.method == "heartbeat") {
                    heartbeatMisses = 0;
                    return;
                }

                for (var i = 0, r; r = receivers[i]; ++i) {
                    if (dispatchData(r, data)) {
                        return;
                    }
                }

                for (var i = 0, r; r = requestPool[i]; ++i) {
                    if (dispatchData(r, data)) {
                        requestPool.splice(i, 1)
                        return;
                    }
                }
            };

            webSocket.onclose = function(event) {
                initializing = false;
                setTimeout(function() {
                    this._init();
                }.bind(this), this.retryTimeout * 1000)
            }.bind(this);
        },

        onRequestError: function(event, detail) {
            this.fire('rf-api-error', detail);
        },

        _init: function() {
            if (!this.user || initializing) {
                return;
            }

            if (webSocket != null && webSocket.readyState == WebSocket.OPEN) {
                return;
            }

            if (webSocket != null) {
                clearHeartbeat();
            }

            initializing = true;

            return this.$.nonce.generateRequest();
        }

    });
})();
