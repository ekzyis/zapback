package server

import (
	"net/http"
	"time"

	"github.com/ekzyis/zapback/env"
	"github.com/ekzyis/zapback/lightning"
	"github.com/ekzyis/zapback/lightning/lnurl"
	"github.com/ekzyis/zapback/pages"
	"github.com/ekzyis/zapback/pages/components"
	"github.com/ekzyis/zapback/pages/types"
	"github.com/labstack/echo/v4"
)

func index(sCtx Context) echo.HandlerFunc {
	return func(eCtx echo.Context) error {
		return pages.Render(pages.Index(), http.StatusOK, eCtx)
	}
}

func newGame(sCtx Context) echo.HandlerFunc {
	return func(eCtx echo.Context) error {
		return pages.Render(pages.NewGame(nil), http.StatusOK, eCtx)
	}
}

func createGame(sCtx Context) echo.HandlerFunc {
	return func(eCtx echo.Context) error {
		var form struct {
			ZapAmount        int    `form:"zap_amount"`
			LightningAddress string `form:"lnaddr"`
			Players          int    `form:"players"` // TODO: allow more than two players
		}
		if err := eCtx.Bind(&form); err != nil {
			return err
		}

		formError := types.FormError{}

		if form.ZapAmount < 1 {
			formError["zap_amount"] = "must be greater than 0"
		}

		if form.LightningAddress == "" {
			formError["lnaddr"] = "required"
		} else if err := lnurl.VerifyLNURLp(form.LightningAddress); err != nil {
			// XXX expose detailed error message?
			formError["lnaddr"] = "invalid lightning address"
		}

		if len(formError) > 0 {
			eCtx.Response().Header().Add("HX-Retarget", "#content")
			eCtx.Response().Header().Add("HX-Reselect", "#content")
			eCtx.Logger().Error(formError)
			return pages.Render(
				pages.NewGame(formError),
				http.StatusBadRequest,
				eCtx,
			)
		}

		pr, err := sCtx.Ln.CreateInvoice(int64(form.ZapAmount*1000), "zapback: new game")
		if err != nil {
			return err
		}

		decoded, err := lightning.DecodePaymentRequest(pr)
		if err != nil {
			return err
		}

		return pages.RenderModal(components.Invoice(decoded), http.StatusOK, eCtx)
	}
}

func invoiceStatus(sCtx Context) echo.HandlerFunc {
	return func(eCtx echo.Context) error {
		inv, err := sCtx.Ln.GetInvoice(eCtx.Param("payment_hash"))
		if err != nil {
			return err
		}

		// mark invoices as paid after 5 seconds in debug mode
		if env.Debug && time.Since(inv.CreatedAt) >= 5*time.Second {
			inv.ConfirmedAt = time.Now()
		}

		return pages.Render(components.InvoiceStatus(inv), http.StatusOK, eCtx)
	}
}
