package provider

import (
	"context"
	"fmt"
	"log"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

// ContextFromMessage enrich given context with span
func ContextFromMessage(ctx context.Context, p *WorkerMessage) context.Context {
	if ctx == nil {
		return ctx
	}

	tracer := opentracing.GlobalTracer()

	spanContext, err := tracer.Extract(opentracing.TextMap, opentracing.TextMapCarrier(p.Tracing))
	if err != nil {
		log.Printf(`[tracing] Error while extracting span from WorkerMessage: %v`, err)
		return nil
	}

	span := tracer.StartSpan(fmt.Sprintf(`%s/%s`, p.Source, p.Type), ext.RPCServerOption(spanContext))

	return opentracing.ContextWithSpan(ctx, span)
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
