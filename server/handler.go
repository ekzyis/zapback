package server

import (
	"net/http"

	"github.com/ekzyis/zapback/lightning"
	"github.com/ekzyis/zapback/pages"
	"github.com/ekzyis/zapback/pages/components"
	"github.com/labstack/echo/v4"
)

func index(sCtx Context) echo.HandlerFunc {
	return func(eCtx echo.Context) error {
		return pages.Render(pages.Index(), http.StatusOK, eCtx)
	}
}

func newGame(sCtx Context) echo.HandlerFunc {
	return func(eCtx echo.Context) error {
		return pages.Render(pages.NewGame(), http.StatusOK, eCtx)
	}
}

func createGame(sCtx Context) echo.HandlerFunc {
	return func(eCtx echo.Context) error {
		var form struct {
			ZapAmount int `form:"zap_amount"`
		}
		if err := eCtx.Bind(&form); err != nil {
			return err
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

		return pages.Render(components.InvoiceStatus(inv), http.StatusOK, eCtx)
	}
}
