package server

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

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

func gameForm(sCtx Context) echo.HandlerFunc {
	return func(eCtx echo.Context) error {
		return pages.Render(pages.GameForm(nil), http.StatusOK, eCtx)
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
			return pages.Render(pages.GameForm(formError), http.StatusBadRequest, eCtx)
		}

		desc := "zapback: new game"
		msats := int64(form.ZapAmount * 1000)

		pr, err := sCtx.Ln.CreateInvoice(msats, desc, time.Now().Add(5*time.Minute))
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
			Type:           db.InvoiceTypeCreate,
		})
		if err != nil {
			return err
		}

		if err := tx.Commit(); err != nil {
			return err
		}

		// we will use this cookie to render a different page for the inviter vs invitee when they visit /game/:code
		// we will also unset the cookie on /game/:code
		eCtx.SetCookie(&http.Cookie{
			Name:     "inviter",
			Value:    "true",
			Path:     fmt.Sprintf("/game/%s", game.Code),
			Secure:   true,
			SameSite: http.SameSiteLaxMode,
			HttpOnly: true,
		})

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

		cookie, _ := eCtx.Cookie("inviter")
		inviter := false
		if cookie != nil {
			inviter = cookie.Value == "true"
		}

		if inviter {
			eCtx.SetCookie(&http.Cookie{
				Name:     "inviter",
				Value:    "",
				Path:     fmt.Sprintf("/game/%s", code),
				MaxAge:   -1,
				Expires:  time.Unix(0, 0),
				Secure:   true,
				SameSite: http.SameSiteLaxMode,
				HttpOnly: true,
			})
			return pages.Render(pages.GameSendInvite(game), http.StatusOK, eCtx)
		}

		tx, err := sCtx.Db.BeginTx(eCtx.Request().Context(), nil)
		if err != nil {
			return err
		}

		if started, err := tx.HasGameStarted(game.Id); err != nil {
			return err
		} else if started {
			turn, err := tx.GetGameTurn(game.Id)
			if err != nil {
				return err
			}

			expired := turn.Invoice.ExpiresAt.Before(time.Now())
			if expired {
				winner, err := tx.GetGameWinner(game.Id)
				if err != nil {
					return fmt.Errorf("failed to get game winner: %s", err.Error())
				}

				if err := tx.Commit(); err != nil {
					return err
				}

				return pages.Render(pages.GameFinished(game, turn, winner), http.StatusOK, eCtx)
			}

			if err := tx.Commit(); err != nil {
				return err
			}

			return pages.Render(pages.Game(game, turn), http.StatusOK, eCtx)
		}

		if err := tx.Commit(); err != nil {
			return err
		}

		return pages.Render(pages.GameInvite(game, nil), http.StatusOK, eCtx)
	}
}

func startGame(sCtx Context) echo.HandlerFunc {
	return func(eCtx echo.Context) error {
		code := eCtx.Param("code")

		var form struct {
			LightningAddress string `form:"lnaddr"`
		}
		if err := eCtx.Bind(&form); err != nil {
			return err
		}

		game, err := sCtx.Db.GetGame(code)
		if err != nil {
			return err
		}

		tx, err := sCtx.Db.BeginTx(eCtx.Request().Context(), nil)
		if err != nil {
			return err
		}

		if started, err := tx.HasGameStarted(game.Id); err != nil {
			return err
		} else if started {
			return echo.NewHTTPError(http.StatusBadRequest, "game already started")
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
			return pages.Render(pages.GameInvite(game, formError), http.StatusBadRequest, eCtx)
		}

		desc := "zapback: start game"
		msats := game.ZapAmount
		expiresAt := time.Now().Add(5 * time.Minute)
		pr, err := sCtx.Ln.CreateInvoice(msats, desc, expiresAt)
		if err != nil {
			return err
		}

		decoded, err := lightning.DecodePaymentRequest(pr)
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
			Type:           db.InvoiceTypeInvite,
		})
		if err != nil {
			if strings.Contains(err.Error(), "duplicate key value violates unique constraint \"invoice_unique_invite_per_game\"") {
				formError["lnaddr"] = "invite no longer available"
				eCtx.Response().Header().Add("HX-Retarget", "#content")
				eCtx.Response().Header().Add("HX-Reselect", "#content")
				eCtx.Logger().Error(formError)
				return pages.Render(pages.GameInvite(game, formError), http.StatusBadRequest, eCtx)
			}

			return err
		}

		if err := tx.Commit(); err != nil {
			return err
		}

		return pages.RenderModal(components.Invoice(inv), http.StatusOK, eCtx)
	}
}

func loadGame(sCtx Context) echo.HandlerFunc {
	return func(eCtx echo.Context) error {
		code := eCtx.FormValue("code")
		if code == "" {
			var err types.FormError
			// if POST, then we're submitting the form, else we're just loading the page
			if eCtx.Request().Method == "POST" {
				err = types.FormError{
					"code": "required",
				}
			}
			return pages.Render(
				pages.LoadGame(err),
				http.StatusOK,
				eCtx,
			)
		}

		return eCtx.Redirect(http.StatusSeeOther, fmt.Sprintf("/game/%s", code))
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
