package server

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/ekzyis/zapback/db"
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
			return pages.Render(pages.NewGame(formError), http.StatusBadRequest, eCtx)
		}

		desc := "zapback: new game"
		msats := int64(form.ZapAmount * 1000)

		pr, err := sCtx.Ln.CreateInvoice(msats, desc)
		if err != nil {
			return err
		}

		decoded, err := lightning.DecodePaymentRequest(pr)
		if err != nil {
			return err
		}

		tx, err := sCtx.Db.BeginTx(eCtx.Request().Context(), nil)
		if err != nil {
			return err
		}

		game, err := tx.CreateGame(&db.CreateGame{ZapAmount: msats})
		if err != nil {
			return err
		}

		player, err := tx.CreatePlayer(&db.CreatePlayer{
			LightningAddress: form.LightningAddress,
			GameId:           game.Id,
		})
		if err != nil {
			return err
		}

		inv, err := tx.CreateInvoice(&db.CreateInvoice{
			CreatedAt:      decoded.CreatedAt,
			ExpiresAt:      decoded.ExpiresAt,
			PaymentHash:    decoded.PaymentHash,
			PaymentRequest: string(pr),
			MsatsRequested: msats,
			Description:    desc,
			PlayerId:       player.Id,
			GameId:         game.Id,
		})
		if err != nil {
			return err
		}

		if err := tx.Commit(); err != nil {
			return err
		}

		return pages.RenderModal(components.Invoice(inv), http.StatusOK, eCtx)
	}
}

func game(sCtx Context) echo.HandlerFunc {
	return func(eCtx echo.Context) error {
		code := eCtx.Param("code")
		game, err := sCtx.Db.GetGame(code)
		if err != nil {
			return err
		}

		// TODO: render different page if game started
		return pages.Render(pages.Game(game, nil), http.StatusOK, eCtx)
	}
}

func startGame(sCtx Context) echo.HandlerFunc {
	return func(eCtx echo.Context) error {
		code := eCtx.Param("code")

		game, err := sCtx.Db.GetGame(code)
		if err != nil {
			return err
		}

		var form struct {
			LightningAddress string `form:"lnaddr"`
		}
		if err := eCtx.Bind(&form); err != nil {
			return err
		}

		formError := types.FormError{}

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
			return pages.Render(pages.Game(game, formError), http.StatusBadRequest, eCtx)
		}

		desc := "zapback: start game"
		msats := game.ZapAmount

		pr, err := sCtx.Ln.CreateInvoice(msats, desc)
		if err != nil {
			return err
		}

		decoded, err := lightning.DecodePaymentRequest(pr)
		if err != nil {
			return err
		}

		tx, err := sCtx.Db.BeginTx(eCtx.Request().Context(), nil)
		if err != nil {
			return err
		}

		player, err := tx.CreatePlayer(&db.CreatePlayer{
			LightningAddress: form.LightningAddress,
			GameId:           game.Id,
		})
		if err != nil {
			return err
		}

		inv, err := tx.CreateInvoice(&db.CreateInvoice{
			CreatedAt:      decoded.CreatedAt,
			ExpiresAt:      decoded.ExpiresAt,
			PaymentHash:    decoded.PaymentHash,
			PaymentRequest: string(pr),
			MsatsRequested: msats,
			Description:    desc,
			PlayerId:       player.Id,
			GameId:         game.Id,
		})
		if err != nil {
			return err
		}

		if err := tx.Commit(); err != nil {
			return err
		}

		return pages.RenderModal(components.Invoice(inv), http.StatusOK, eCtx)
	}
}

func invoiceStatus(sCtx Context) echo.HandlerFunc {
	return func(eCtx echo.Context) error {
		inv, err := sCtx.Db.GetInvoice(eCtx.Param("payment_hash"))
		if err != nil {
			return err
		}

		var redirectUrl *url.URL
		if inv.GameCode.Valid {
			redirectUrl, err = url.Parse(fmt.Sprintf("/game/%s", inv.GameCode.String))
			if err != nil {
				return err
			}
		}

		return pages.Render(components.InvoiceStatus(inv, redirectUrl), http.StatusOK, eCtx)
	}
}
