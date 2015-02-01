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

    function toQueryString(params) {
        var r = [];
        for (var n in params) {
            var v = params[n];
            n = encodeURIComponent(n);
            if (Array.isArray(v)) {
                v.forEach(function(val) {
                    r.push(val == null ? n : (n + '=' + encodeURIComponent(val)));
                });
            } else {
                r.push(v == null ? n : (n + '=' + encodeURIComponent(v)));
            }
        }
        return r.join('&');
    }

    function urlQuery(url, params) {
        /* Make sure the order of the parameters is
         * consistent with that's in the message, don't
         * let the xhr do it */
        if (params == "") {
            return url;
        }
        url += (url.indexOf('?') > 0 ? '&' : '?') + params;

        return url
    }

    Polymer('rf-api', {
        version: 1,
        pathAction: '',
        webSocket: null,
        messagePool: [],

        ready: function() {
            this.$.request.xhr.toQueryString = toQueryString;
        },

        go: function() {
            if (!this.user) {
                return;
            }

            var self = this;

            this.$.nonce.processResponse = function(xhr) {
                var response = this.evalResponse(xhr);
                this.response = response;

                if (!self.url) {
                    self.url = self.getAttribute('data-api-pattern') + "v" + self.version + "/" + self.pathAction
                }

                if (self.socket) {
                    var params = toQueryString(self.$.request.getParams((self.xhrArgs || {}).params)),
                        date = new Date().getTime(),
                        nonce = response.Nonce,
                        url = urlQuery(self.url, params), urlParts;

                    if (self.webSocket != null) {
                        self.webSocket.close();
                    }

                    urlParts = url.split("?", 2);
                    urlParts[0] = encodeURI(urlParts[0]);

                    signature = generateSignature(urlParts.join("?"), 'GET', null,
                        "", date, nonce, self.user.MD5API || "");

                    self.webSocket = new WebSocket(
                        urlQuery("ws://" + location.host + url,
                            "login=" + encodeURIComponent(self.user.Login) +
                            "&signature=" + encodeURIComponent(signature) +
                            "&date=" + encodeURIComponent(date) +
                            "&nonce=" + encodeURIComponent(nonce)));

                    self.webSocket.onopen = function() {
                        while (self.messagePool.length) {
                            var m = self.messagePool.shift();

                            self.send(m);
                        }
                    };

                    self.webSocket.onmessage = function(event) {
                        var data = event.data;
                        try {
                            data = JSON.parse(data)
                        } catch(e) {}

                        self.fire('rf-api-message', data);
                    };
                } else {
                    xhr = self.$.request.xhr;

                    self.$.request.url = self.url;
                    self.$.request.xhr = {
                        request: function(args) {
                            var method = args.method.toUpperCase(),
                                params = xhr.toQueryString(args.params),
                                signature, urlParts;

                            if (xhr.isBodyMethod(method) && !args.body) {
                                args.body = params;
                            } else if (params && (!method || method == 'GET')) {
                                args.url = urlQuery(args.url, params);

                                delete args.params;
                            }

                            if (method == 'POST' && args.headers['Content-Type'].indexOf('charset=') == -1) {
                                args.headers['Content-Type'] += "; charset=UTF-8";
                            }

                            args.headers['X-Date'] = new Date().toUTCString();
                            args.headers['X-Nonce'] = response.Nonce;

                            urlParts = args.url.split("?", 2);
                            urlParts[0] = encodeURI(urlParts[0]);

                            signature = generateSignature(urlParts.join("?"), method, args.body,
                                args.headers['Content-Type'],
                                args.headers['X-Date'],
                                args.headers['X-Nonce'],
                                self.user.MD5API || ""
                            );

                            args.headers['Authorization'] = "Readeef " + self.user.Login + ":" + signature;

                            xhr.request(args);
                        }
                    };
                    self.$.request.go();
                    self.$.request.xhr = xhr;
                }
            }

            return this.$.nonce.go();
        },

        send: function(data) {
            if (!this.webSocket || this.webSocket.readyState != WebSocket.OPEN) {
                this.messagePool.push(data);

                this.go();
                return;
            }

            data = JSON.stringify(data);
            this.webSocket.send(data);
        },

        onRequestResponse: function(event, detail) {
            this.fire('rf-api-response', detail);
        },

        onRequestError: function(event, detail) {
            this.fire('rf-api-error', detail);
        },

        onRequestComplete: function(event, detail) {
            this.fire('rf-api-complete', detail);
        }
    });
})();
