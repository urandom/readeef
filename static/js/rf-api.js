(function() {
    Polymer('rf-api', {
        version: 1,
        pathAction: '',

        ready: function() {
            this.$.request.xhr.toQueryString = function(params) {
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
            };
        },

        go: function() {
            if (!this.user) {
                return;
            }

            var self = this;

            this.$.nonce.processResponse = function(xhr) {
                var response = this.evalResponse(xhr);
                this.response = response;

                var xhr = self.$.request.xhr;

                if (!self.url) {
                    self.url = "{% .apiPattern %}v" + self.version + "/" + self.pathAction
                }

                self.$.request.url = self.url;
                self.$.request.xhr = {
                    request: function(args) {
                        var method = (args.method.toUpperCase() || 'GET'),
                            params = xhr.toQueryString(args.params),
                            urlParts, url, message, messageHash, body, bodyHash;

                        if (xhr.isBodyMethod(method) && !args.body) {
                            args.body = params;
                        } else if (params && method == 'GET') {
                            /* Make sure the order of the parameters is
                             * consistent with that's in the message, don't
                             * let the xhr do it */
                            args.url += (args.url.indexOf('?') > 0 ? '&' : '?') + params;

                            delete args.params;
                        }

                        if (method == 'POST' && args.headers['Content-Type'].indexOf('charset=') == -1) {
                            args.headers['Content-Type'] += "; charset=UTF-8";
                        }

                        body = args.body;
                        if (body != null) {
                            body = body.toString();
                        }
                        bodyHash = CryptoJS.MD5(body).toString(CryptoJS.enc.Base64);

                        args.headers['X-Date'] = new Date().toUTCString();
                        args.headers['X-Nonce'] = response.Nonce;

                        urlParts = args.url.split("?", 2);
                        urlParts[0] = encodeURI(urlParts[0]);
                        url = urlParts.join("?");

                        message = url + "\n" + method + "\n"
                                + bodyHash + "\n" + args.headers['Content-Type'] + "\n"
                                + args.headers['X-Date'] + "\n" + args.headers['X-Nonce'] + "\n";

                        messageHash = CryptoJS.HmacSHA256(message, self.user.MD5API || "").toString(CryptoJS.enc.Base64);

                        args.headers['Authorization'] = "Readeef " + self.user.Login + ":" + messageHash;

                        xhr.request(args);
                    }
                };
                self.$.request.go();
                self.$.request.xhr = xhr;
            }

            return this.$.nonce.go();
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
