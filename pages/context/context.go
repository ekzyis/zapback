package context

import "context"

type ContextKey string

var (
	Env     ContextKey = "env"
	Session ContextKey = "session"
	Commit  ContextKey = "commit"
	Origin  ContextKey = "origin"
)

func GetOrigin(ctx context.Context) string {
	if u, ok := ctx.Value(Origin).(string); ok {
		return u
	}
	// https://developer.mozilla.org/en-US/docs/Web/API/URL/origin
	return "https://zapback.ekzy.is"
}
