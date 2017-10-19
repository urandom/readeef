package monitor

import (
	"sync"

	"github.com/pkg/errors"
	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/repo/eventable"
	"github.com/urandom/readeef/content/thumbnail"
	"github.com/urandom/readeef/log"
)

func Thumbnailer(service eventable.Service, generator thumbnail.Generator, log log.Log) {
	for event := range service.Listener() {
		switch data := event.Data.(type) {
		case eventable.FeedUpdateData:
			go processThumbnailerEvent(data, generator, log)
		}
	}
}

func processThumbnailerEvent(data eventable.FeedUpdateData, generator thumbnail.Generator, log log.Log) {
	log.Infof("Generating article thumbnails for feed %s", data.Feed)

	processors := generateProcessors(data.NewArticles)
	numProcessors := 20
	done := make(chan struct{})
	errc := make(chan error)

	defer close(done)

	var wg sync.WaitGroup

	wg.Add(numProcessors)
	for i := 0; i < numProcessors; i++ {
		go func() {
			err := process(generator, done, processors)
			if err != nil {
				errc <- err
			}
			wg.Done()
		}()
	}

	go func() {
		wg.Wait()
		close(errc)
	}()

	for err := range errc {
		log.Printf("Error generating thumbnails for feed %s articles: %+v", data.Feed, err)
	}
}

func generateProcessors(articles []content.Article) <-chan content.Article {
	processors := make(chan content.Article)

	go func() {
		defer close(processors)

		for _, a := range articles {
			processors <- a
		}
	}()

	return processors
}

func process(generator thumbnail.Generator, done <-chan struct{}, processors <-chan content.Article) error {
	for a := range processors {
		select {
		case <-done:
			return nil
		default:
			if err := generator.Generate(a); err != nil {
				return errors.Wrapf(err, "generating thumbnail for article %s", a)
			}
		}
	}

	return nil
}
