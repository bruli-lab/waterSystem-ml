package app

import (
	"context"

	"github.com/bruli-lab/go-core/cqs"
	"github.com/bruli-lab/go-core/event"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
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
	tracer    trace.Tracer
}

func (p PublishMessage) Handle(ctx context.Context, cmd cqs.Command) ([]event.Event, error) {
	ctx, span := p.tracer.Start(ctx, "app.PublishMessage")
	defer span.End()
	co, ok := cmd.(PublishMessageCommand)
	if !ok {
		err := cqs.NewInvalidCommandError(PublishMessageCommandName, cmd.Name())
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	return nil, p.publisher.Publish(ctx, co.Message)
}

func NewPublishMessage(publisher Publisher, tracer trace.Tracer) *PublishMessage {
	return &PublishMessage{publisher: publisher, tracer: tracer}
}
