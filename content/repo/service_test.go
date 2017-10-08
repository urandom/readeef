package repo_test

import "testing"

func TestService(t *testing.T) {
	skipTest(t)

	if service.UserRepo() == nil {
		t.Fatal("service.UserRepo() = nil")
	}

	if service.TagRepo() == nil {
		t.Fatal("service.TagRepo() = nil")
	}

	if service.FeedRepo() == nil {
		t.Fatal("service.FeedRepo() = nil")
	}

	if service.SubscriptionRepo() == nil {
		t.Fatal("service.SubscriptionRepo() = nil")
	}

	if service.ArticleRepo() == nil {
		t.Fatal("service.ArticleRepo() = nil")
	}

	if service.ExtractRepo() == nil {
		t.Fatal("service.ExtractRepo() = nil")
	}

	if service.ThumbnailRepo() == nil {
		t.Fatal("service.ThumbnailRepo() = nil")
	}

	if service.ScoresRepo() == nil {
		t.Fatal("service.ScoresRepo() = nil")
	}
}
