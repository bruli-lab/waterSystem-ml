package app

import (
	"context"

	"github.com/bruli-lab/go-core/cqs"
	"github.com/bruli-lab/go-core/event"
)

const PublishMessageCommandName = "publish_message"

type PublishMessageCommand struct {
	Message string
}

func (p PublishMessageCommand) Name() string {
	return PublishMessageCommandName
}

type PublishMessage struct {
	publisher Publisher
}

func (p PublishMessage) Handle(ctx context.Context, cmd cqs.Command) ([]event.Event, error) {
	co, ok := cmd.(PublishMessageCommand)
	if !ok {
		return nil, cqs.NewInvalidCommandError(PublishMessageCommandName, cmd.Name())
	}
	return nil, p.publisher.Publish(ctx, co.Message)
}

func NewPublishMessage(publisher Publisher) *PublishMessage {
	return &PublishMessage{publisher: publisher}
}
