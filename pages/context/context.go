package context

type ContextKey string

var (
	Env     ContextKey = "env"
	Session ContextKey = "session"
	Commit  ContextKey = "commit"
)
