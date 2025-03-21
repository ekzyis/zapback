package pages

import (
	"context"

	"github.com/a-h/templ"
	"github.com/ekzyis/zapback/env"
	"github.com/ekzyis/zapback/pages/components"
	pCtx "github.com/ekzyis/zapback/pages/context"
	"github.com/labstack/echo/v4"
)

func GetEnv(ctx context.Context) string {
	if u, ok := ctx.Value(pCtx.Env).(string); ok {
		return u
	}
	return "development"
}

func Render(t templ.Component, statusCode int, eCtx echo.Context) error {
	buf := templ.GetBuffer()
	defer templ.ReleaseBuffer(buf)

	rCtx := context.WithValue(eCtx.Request().Context(), pCtx.Env, env.Env)

	if err := t.Render(rCtx, buf); err != nil {
		return err
	}

	return eCtx.HTML(statusCode, buf.String())
}

func RenderModal(child templ.Component, statusCode int, eCtx echo.Context) error {
	buf := templ.GetBuffer()
	defer templ.ReleaseBuffer(buf)

	rCtx := templ.WithChildren(eCtx.Request().Context(), child)

	if err := components.Modal(true).Render(rCtx, buf); err != nil {
		return err
	}

	return eCtx.HTML(statusCode, buf.String())
}
