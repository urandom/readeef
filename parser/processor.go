package parser

type Processor interface {
	Process(Feed) Feed
}
