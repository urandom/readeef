package monitor

import (
	"fmt"
	"sync"

	"github.com/urandom/readeef/content"
	"github.com/urandom/readeef/content/thumbnail"
	"github.com/urandom/readeef/log"
)

type Thumbnailer struct {
	generator thumbnail.Generator
	log       log.Log
}

func NewThumbnailer(g thumbnail.Generator, l log.Log) Thumbnailer {
	return Thumbnailer{generator: g, log: l}
}

func (t Thumbnailer) FeedUpdated(feed content.Feed, articles []content.Article) error {
	t.log.Debugln("Generating thumbnailer processors")

	processors := t.generateProcessors(articles)
	numProcessors := 20
	done := make(chan struct{})
	errc := make(chan error)

	defer close(done)

	var wg sync.WaitGroup

	wg.Add(numProcessors)
	for i := 0; i < numProcessors; i++ {
		go func() {
			err := t.process(done, processors)
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
		return err
	}

	return nil
}

func (t Thumbnailer) FeedDeleted(feed content.Feed) error {
	return nil
}

func (t Thumbnailer) generateProcessors(articles []content.Article) <-chan content.Article {
	processors := make(chan content.Article)

	go func() {
		defer close(processors)

		for _, a := range articles {
			processors <- a
		}
	}()

	return processors
}

func (t Thumbnailer) process(done <-chan struct{}, processors <-chan content.Article) error {
	for a := range processors {
		select {
		case <-done:
			return nil
		default:
			if err := t.generator.Generate(a); err != nil {
				return fmt.Errorf("Error generating thumbnail: %v\n", err)
			}
		}
	}

	return nil
}
