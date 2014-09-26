(function() {
    "use strict";

    Polymer('rf-article-display', {
        observe: {
          'articles template': 'initialize'
        },

        attached: function() {
            this.template = this.querySelector('template');
            this.template.setAttribute('id', 'article-template');

            this.templates = [];

            [0, 1, 2].forEach(function(index) {
                var template = document.createElement('template');
                template.setAttribute('ref', 'article-template');

                this.template.parentNode.insertBefore(template, this.template);
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
                    var model = this._physicalArticles[i];

                    if (this.$.viewport.children[i]) {
                        this.$.viewport.insertBefore(t.createInstance(model), this.$.viewport.children[i]);
                    } else {
                        this.$.viewport.appendChild(t.createInstance(model));
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
        }
    });
})();
