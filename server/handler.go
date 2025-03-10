package server

import (
	"net/http"

	"github.com/ekzyis/zapback/pages"
	"github.com/labstack/echo/v4"
)

func index(sCtx Context) echo.HandlerFunc {
	return func(eCtx echo.Context) error {
		return pages.Render(pages.Index(), http.StatusOK, eCtx)
	}
}
