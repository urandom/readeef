(function() {
    "use strict";

    Polymer('rf-article-display', {
        swiping: false,
        formatting: false,

        hasTransform: true,
        hasWillChange: true,

        observe: {
          'articles template': 'initialize'
        },

        eventDelegates: {
            trackstart: 'trackStart',
            trackx: 'trackx',
            trackend: 'trackEnd'
        },

        created: function() {
            this.hasTransform = 'transform' in this.style;
            this.hasWillChange = 'willChange' in this.style;
        },

        attached: function() {
            this.template = this.querySelector('template');
            this.template.setAttribute('id', 'article-template');

            this.templates = [];

            [0, 1, 2].forEach(function(index) {
                var template = document.createElement('template');
                template.setAttribute('ref', 'article-template');

                this.template.parentNode.insertBefore(template, this.template);
                HTMLTemplateElement.decorate(template);
                this.templates[index] = template;
            }.bind(this));

            this.addEventListener('transitionend', function() {
                this.$.viewport.removeAttribute('animate');
                for (var i = 0, t; t = this.templates[i]; ++i) {
                    if (t._element) {
                        t._element.removeAttribute('animate');
                    }
                }
            }.bind(this), false);
        },

        initialize: function() {
            if (!this.articles || !this.template) {
                return;
            }

            this._originalArticleProperties = Object.getOwnPropertyNames(this.articles[0] || {});
            this._physicalArticles = [];
            for (var i = 0; i < 3; ++i) {
                this._physicalArticles[i] = {};
                this.updateItem(i, i);
                this.cleanTemplateElement(this.templates[i]);
            }
        },

        updateItem: function(virtualIndex, physicalIndex) {
            var src = this.articles[virtualIndex],
                dest = this._physicalArticles[physicalIndex],
                middle = 1;

            if (!src) {
                return;
            }

            dest.selected = physicalIndex == middle;
            dest.wide = this.wide;
            for (var i = 0, p; p = this._originalArticleProperties[i]; ++i) {
                dest[p] = src[p];
            }

            dest._physicalIndex = physicalIndex;
            dest._virtualIndex = virtualIndex;
        },

        articleChanged: function() {
            if (!this.article) {
                return;
            }

            var virtual = this.articles.indexOf(this.article),
                article, physical = -1, middle = 1;

            if (virtual == -1) {
                return;
            }

            for (var i = 0, a; a = this._physicalArticles[i]; ++i) {
                a.selected = false;
                if (a.Id == this.article.Id) {
                    a.selected = true;
                    physical = i;
                }
            }

            if (physical == 0) {
                this._physicalArticles.pop();
                this._physicalArticles.unshift({});
                this.updateItem(virtual - 1, 0);

                this.templates.unshift(this.templates.pop());
                this.cleanTemplateElement(this.templates[0]);
            } else if (physical == 2) {
                this._physicalArticles.shift();
                this._physicalArticles.push({});
                this.updateItem(virtual + 1, 2);

                this.templates.push(this.templates.shift());
                this.cleanTemplateElement(this.templates[2]);
            } else if (physical == -1) {
                for (var i = virtual - middle, j = 0; j < 3; ++i, ++j) {
                    this.updateItem(i, j);
                    this.cleanTemplateElement(this.templates[j]);
                }
            }

            Platform.flush();

            this.$.viewport.setAttribute('animate', '');

            for (var i = 0, t; t = this.templates[i]; ++i) {
                if (t._element) {
                    t._element.setAttribute('animate', '');
                } else {
                    var model = this._physicalArticles[i],
                        syntax = this.template.bindingDelegate || this.templateInstance.model.element.syntax,
                        dom = t.createInstance(model, syntax);

                    if (this.$.viewport.children[i]) {
                        this.$.viewport.insertBefore(dom, this.$.viewport.children[i]);
                    } else {
                        this.$.viewport.appendChild(dom);
                    }

                    t._element = this.$.viewport.children[i];
                    this.initializeElement(t._element, model);
                }
            }

            // Force layout
            this.templates[0]._element.offsetTop;

            for (var i = 0, t; t = this.templates[i]; ++i) {
                if (i == middle) {
                    t._element.classList.add('selected');
                } else {
                    t._element.classList.remove('selected');
                }
            }
        },

        onFavoriteArticleToggle: function() {
            if ('onFavoriteArticleToggle' in this.templateInstance.model) {
                this.templateInstance.model.onFavoriteArticleToggle.apply(
                    this.templateInstance.model, arguments);

                this.updateItem(this._physicalArticles[1]._virtualIndex, 1);
            }
        },

        cleanTemplateElement: function(template) {
            if (template._element) {
                if (template._element.parentNode) {
                    template._element.parentNode.removeChild(template._element);
                }
                template._element = undefined;
            }
        },

        initializeElement: function(item, model) {
            if (!model.Description) {
                return;
            }

            var description = item.querySelector('.article-description'),
                imageStyler = function() {
                    if (image.width < 300) {
                        (image.parentNode || description).classList.add('clearfix');
                        image.style.float = "right";
                    }
                }, image;

            description.innerHTML = model.Description;
            image = description.querySelector('img');

            if (image) {
                if (image.complete) {
                    imageStyler();
                } else {
                    image.addEventListener('load', imageStyler);
                }
            }

            Array.prototype.forEach.call(
                description.querySelectorAll('img'),
                function(element) {
                    element.style.width = 'auto';
                    element.style.height = 'auto';
                }
            );
        },

        openCurrentArticle: function() {
            if (this.templates[1] && this.templates[1]._element) {
                this.templates[1]._element.querySelector('.article-link').openInBackground();
            }
        },

        swipingChanged: function() {
            if (this.swiping) {
                this.$.viewport.setAttribute('swipe', '');
            } else {
                this.$.viewport.removeAttribute('swipe');
            }
        },

        trackStart: function(event) {
            if (this.wide) {
                return;
            }

            this.swiping = true;
            this.swipeStart = new Date().getTime();

            this.articleWidth = this.templates[1]._element.offsetWidth;

            event.preventTap();
        },

        trackEnd: function(event) {
            if (this.swiping) {
                this.swiping = false;
                this.moveArticles(null);

                var absDx = Math.abs(event.dx),
                    speed = absDx / (new Date().getTime() - this.swipeStart);

                if ((absDx > this.articleWidth / 2) || (speed > 0.5 && absDx > 40)) {
                    if (event.dx > 0) {
                        this.fire('core-signal', {name: 'rf-previous-article'});
                    } else {
                        this.fire('core-signal', {name: 'rf-next-article'});
                    }
                }
            }
        },

        trackx: function(event) {
            if (this.swiping) {
                if ((!this._physicalArticles[2].Id && event.dx < 0)
                    || (!this._physicalArticles[0].Id && event.dx > 0)) {
                    return;
                }

                this.moveArticles(event.dx);
            }
        },

        transformForTranslateX: function (translateX) {
            if (translateX === null) {
                return '';
            }

            return this.hasWillChange ? 'translateX(' + translateX + 'px)' : 'translate3d(' + translateX + 'px, 0, 0)';
        },

        moveArticles: function(translateX) {
            var prop = this.hasTransform ? 'transform' : 'webkitTransform',
                moveIndex = 0, resetIndex = 2, alterTranslateX = translateX - this.articleWidth;

            if (translateX < 0) {
                moveIndex = 2;
                resetIndex = 0;
                alterTranslateX = this.articleWidth + translateX;
            }

            if (translateX === null) {
                alterTranslateX = null;
            }

            this.templates[resetIndex]._element.removeAttribute('swipe');
            if (translateX === null) {
                this.templates[moveIndex]._element.removeAttribute('swipe');
                this.templates[1]._element.removeAttribute('swipe');
            } else {
                this.templates[moveIndex]._element.setAttribute('swipe', '');
                this.templates[1]._element.setAttribute('swipe', '');
            }

            this.templates[moveIndex]._element.style[prop] = this.transformForTranslateX(alterTranslateX);
            this.templates[1]._element.style[prop] = this.transformForTranslateX(translateX);
            this.templates[resetIndex]._element.style[prop] = '';
        },

        onArticleFormat: function() {
            if (this.$.viewport.getAttribute('data-formatter-readability') != 'false') {
                this.formatting = true;
                this.$['article-format'].go();
            }
        },

        onArticleFormatResponse: function(event, data) {
            if (data.response && data.response.ArticleId == this.article.Id
                && data.response.Id == this.article.FeedId) {

                this._physicalArticles[1].Description = data.response.Content;
                this.initializeElement(this.templates[1]._element, this._physicalArticles[1]);
            }
            this.formatting = false;
        }
    });
})();
