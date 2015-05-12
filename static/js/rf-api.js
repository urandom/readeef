(function() {
    var webSocket = null, messagePool = [], requestPool = [], receivers = [], initializing = false;

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
                    r.instance.asyncFire('core-signal', {name: 'rf-connection-unauthorized'});
                } else {
                    if (r.instance.hasAttribute('on-rf-api-error')) {
                        r.instance.fire('rf-api-error', data);
                    } else {
                        r.instance.asyncFire('core-signal', {name: 'rf-api-error', data: data});
                    }
                }
            } else {
                r.instance.fire('rf-api-message', data);
            }
            return true;
        }

        return false;
    }

    Polymer('rf-api', {
        version: 1,
        retryTimeout: 10,
        tag: "",

        ready: function() {
            if (this.receiver) {
                receivers.push({instance: this, method: this.method, tag: this.tag});
            }
        },

        initialize: function() {
            if (!this.user || initializing) {
                return;
            }

            if (webSocket != null && webSocket.readyState == WebSocket.OPEN) {
                return;
            }

            if (webSocket != null) {
                webSocket.close();
            }

            initializing = true;

            var self = this;

            this.$.nonce.processResponse = function(xhr) {
                var response = this.evalResponse(xhr);
                this.response = response;

                if (!self.url) {
                    self.url = self.getAttribute('data-api-pattern') + "v" + self.version + "/";
                }

                var date = new Date().getTime(),
                    nonce = response.Nonce,
                    proto = 'ws';

                signature = generateSignature(encodeURI(self.url), 'GET', null,
                    "", date, nonce, self.user.MD5API || "");

                if (location.protocol == "https:") {
                    proto = 'wss';
                }

                webSocket = new WebSocket(
                    proto + "://" + location.host + self.url +
                    "?login=" + encodeURIComponent(self.user.Login) +
                    "&signature=" + encodeURIComponent(signature) +
                    "&date=" + encodeURIComponent(date) +
                    "&nonce=" + encodeURIComponent(nonce));

                webSocket.onopen = function() {
                    initializing = false;
                    while (messagePool.length) {
                        var m = messagePool.shift();

                        m.instance.send(m.data);
                    }
                };

                webSocket.onmessage = function(event) {
                    var data = event.data;
                    try {
                        data = JSON.parse(data)
                    } catch(e) {}

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
                        self.initialize();
                    }, self.retryTimeout * 1000)
                };
            }

            return this.$.nonce.go();
        },

        send: function(data) {
            if (!webSocket || webSocket.readyState != WebSocket.OPEN) {
                messagePool.push({instance: this, data: data});

                this.initialize();
                return;
            }

            requestPool.push({instance: this, method: this.method, tag: this.tag});

            var payload = {method: this.method, tag: this.tag};

            if (!data && this.arguments) {
                data = this.arguments;
                try {
                    data = JSON.parse(this.arguments);
                } catch (e) {}
            }
            if (data) {
                payload.arguments = data;
            }

            webSocket.send(JSON.stringify(payload));
        },

        onRequestError: function(event, detail) {
            this.fire('rf-api-error', detail);
        },
    });
})();
