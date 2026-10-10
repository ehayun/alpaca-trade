package summary

import (
	"strings"
	"testing"
	"time"

	"alpaca-app/internal/model"
)

func TestPrompt(t *testing.T) {
	ts := time.Date(2026, 10, 9, 21, 3, 0, 0, time.UTC)
	got := Prompt("AAPL", []model.News{
		{Headline: "Apple cuts orders", Summary: " Orders cut 15%. ", CreatedAt: ts},
		{Headline: "No summary here", CreatedAt: ts},
	})
	want := "Recent news for AAPL:\n" +
		"\n- 2026-10-09 21:03 UTC | Apple cuts orders\n  Orders cut 15%.\n" +
		"\n- 2026-10-09 21:03 UTC | No summary here\n"
	if got != want {
		t.Fatalf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestSummarizeWithoutKey(t *testing.T) {
	t.Setenv("ANTHROPIC_API_KEY", "")
	_, err := Summarize(t.Context(), "AAPL")
	if err == nil || !strings.Contains(err.Error(), "ANTHROPIC_API_KEY") {
		t.Fatalf("err = %v, want missing-key user error", err)
	}
}
