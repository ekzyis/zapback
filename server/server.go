package server

import (
	"bytes"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/ekzyis/zapback/pages"
)

type Server struct {
	*echo.Echo
}

func New(sCtx Context) *Server {
	var (
		e *echo.Echo
		s *Server
	)
	e = echo.New()

	e.Static("/", "public")

	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format:           "${time_custom} ${method} ${uri} ${status}\n",
		CustomTimeFormat: "2006-01-02 15:04:05.00000-0700",
	}))
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"http://localhost:4444" /* TODO: add prod domain */},
		AllowCredentials: true,
		AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	}))

	e.HTTPErrorHandler = httpErrorHandler(sCtx)

	s = &Server{e}

	s.GET("/", index(sCtx))

	s.GET("/game", newGame(sCtx))
	s.POST("/game", createGame(sCtx))

	s.GET("/invoice/:payment_hash", invoiceStatus(sCtx))

	go worker(sCtx)

	return s
}

func httpErrorHandler(sCtx Context) echo.HTTPErrorHandler {
	return func(err error, eCtx echo.Context) {
		var (
			code = http.StatusInternalServerError
			page = pages.Error
			buf  *bytes.Buffer
		)

		eCtx.Logger().Error(err)

		if httpError, ok := err.(*echo.HTTPError); ok {
			code = httpError.Code
		}

		// make sure that HTMX selects and targets correct element
		eCtx.Response().Header().Add("HX-Retarget", "#content")
		eCtx.Response().Header().Add("HX-Reselect", "#content")

		if err = pages.Render(page(code), code, eCtx); err != nil {
			eCtx.Logger().Error(err)
			code = http.StatusInternalServerError
		}

		if err = eCtx.HTML(code, buf.String()); err != nil {
			eCtx.Logger().Error(err)
		}
	}
}
