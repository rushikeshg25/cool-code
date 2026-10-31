package agent

import (
	"context"
	"github.com/rushikeshg25/cool-code/internal/llm"
	"github.com/rushikeshg25/cool-code/internal/types"
	"strings"
	"testing"
)

type queueOnReplyProvider struct {
	fakeProvider
	enqueue func()
}

func (p *queueOnReplyProvider) Complete(ctx context.Context, req llm.Request) (llm.Message, error) {
	if p.calls == 0 {
		p.enqueue()
	}
	return p.fakeProvider.Complete(ctx, req)
}
func TestQueuedFollowupAfterTextOnlyResponse(t *testing.T) {
	provider := &queueOnReplyProvider{}
	p := newTestProcessor(t, t.TempDir(), provider, types.ModeAgent)
	provider.enqueue = func() { p.EnqueueMessage("use TypeScript, not Java") }
	if _, err := p.ProcessQuery(context.Background(), "make a plan", nil); err != nil {
		t.Fatal(err)
	}
	if provider.calls != 2 {
		t.Fatalf("provider calls=%d; text-only response stranded queued follow-up", provider.calls)
	}
	found := false
	for _, msg := range provider.lastReq.Messages {
		if strings.Contains(msg.Text, "use TypeScript, not Java") {
			found = true
		}
	}
	if !found {
		t.Fatal("follow-up missing from next request")
	}
}
