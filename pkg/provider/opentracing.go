package provider

import (
	"context"
	"fmt"
	"log"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

// ContextFromMessage enrich given context with span
func ContextFromMessage(ctx context.Context, p *WorkerMessage) (context.Context, opentracing.Span) {
	tracer := opentracing.GlobalTracer()

	if ctx == nil {
		return ctx, tracer.StartSpan(fmt.Sprintf(`%s/%s`, p.Source, p.Type))
	}

	spanContext, err := tracer.Extract(opentracing.TextMap, opentracing.TextMapCarrier(p.Tracing))
	if err != nil {
		log.Printf(`[tracing] Error while extracting span from WorkerMessage: %v`, err)
		return nil, tracer.StartSpan(fmt.Sprintf(`%s/%s`, p.Source, p.Type))
	}

	span := tracer.StartSpan(fmt.Sprintf(`%s/%s`, p.Source, p.Type), ext.RPCServerOption(spanContext))

	return opentracing.ContextWithSpan(ctx, span), span
}

// ContextToMessage enrich message with tracing from context
func ContextToMessage(ctx context.Context, p *WorkerMessage) {
	if ctx == nil {
		return
	}

	tracer := opentracing.GlobalTracer()

	p.Tracing = make(map[string]string)
	if err := tracer.Inject(opentracing.SpanFromContext(ctx).Context(), opentracing.TextMap, opentracing.TextMapCarrier(p.Tracing)); err != nil {
		log.Printf(`[tracing] Error while injecting span to WorkerMessage: %v`, err)
	}
}
