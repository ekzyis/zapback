package server

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/ekzyis/zapback/db"
	"github.com/ekzyis/zapback/env"
	"github.com/ekzyis/zapback/lightning"
)

func logf(format string, a ...any) {
	log.Printf("worker: %s", fmt.Sprintf(format, a...))
}

func worker(sCtx Context) error {
	for {
		invoices, err := sCtx.Db.GetPendingInvoices()
		if err != nil {
			return err
		}

		for _, invoice := range invoices {
			logf("check invoice: %s", invoice.PaymentHash)
			expired := invoice.ExpiresAt.Before(time.Now())
			if expired {
				logf("invoice expired: %s", invoice.PaymentHash)
				continue
			}

			lightningInvoice, err := sCtx.Ln.GetInvoice(invoice.PaymentHash)
			if err != nil {
				logf("failed to check invoice: %s: %s", invoice.PaymentHash, err.Error())
				continue
			}

			debugConfirmed := (env.Debug &&
				time.Since(lightningInvoice.CreatedAt) >= 5*time.Second &&
				invoice.Description.Valid && invoice.Description.String != "zapback: play game")

			if !lightningInvoice.ConfirmedAt.IsZero() || debugConfirmed {
				confirmedAt := lightningInvoice.ConfirmedAt
				if debugConfirmed {
					confirmedAt = time.Now()
					logf("[debug] invoice confirmed: %s", invoice.PaymentHash)
				} else {
					logf("invoice confirmed: %s", invoice.PaymentHash)
				}
				lightningInvoice.ConfirmedAt = confirmedAt

				if err := confirmInvoice(sCtx, &invoice, lightningInvoice); err != nil {
					return err
				}

				continue
			}

			logf("invoice pending: %s", invoice.PaymentHash)
		}

		time.Sleep(5 * time.Second)
	}
}

func confirmInvoice(sCtx Context, dbInv *db.Invoice, lnInv *lightning.Invoice) error {
	tx, err := sCtx.Db.BeginTx(context.Background(), nil)
	if err != nil {
		logf("failed to begin transaction: %s", err.Error())
		return err
	}

	err = tx.UpdateInvoice(dbInv.Id, &db.UpdateInvoice{
		ConfirmedAt: sql.NullTime{
			Time:  lnInv.ConfirmedAt,
			Valid: true,
		},
	})
	if err != nil {
		logf("failed to update invoice: %s", err.Error())
		return tx.Rollback()
	}

	hasGameStarted, err := tx.HasGameStarted(dbInv.GameId)
	if err != nil {
		logf("failed to check if game started: %s", err.Error())
		return tx.Rollback()
	}

	if !hasGameStarted {
		return tx.Commit()
	}

	// game has started, create new invoice for next turn
	if err := nextInvoice(sCtx, tx, dbInv); err != nil {
		logf("failed to create next invoice: %s", err.Error())
		return tx.Rollback()
	}

	return tx.Commit()
}

func nextInvoice(sCtx Context, tx *db.Tx, dbInv *db.Invoice) error {
	logf("creating next invoice ...")

	turn, err := tx.GetGameTurn(dbInv.GameId)
	if err != nil {
		return fmt.Errorf("failed to get game turn: %s", err.Error())
	}

	pr, err := sCtx.Ln.CreateInvoice(dbInv.MsatsRequested, "zapback: play game")
	if err != nil {
		return fmt.Errorf("failed to create lightning invoice: %s", err.Error())
	}

	decoded, err := lightning.DecodePaymentRequest(pr)
	if err != nil {
		return fmt.Errorf("failed to decode pr: %s", err.Error())
	}

	// TODO: make sure invoice expires in 24 hours
	_, err = tx.CreateInvoice(&db.CreateInvoice{
		CreatedAt:      decoded.CreatedAt,
		ExpiresAt:      decoded.ExpiresAt,
		PaymentHash:    decoded.PaymentHash,
		PaymentRequest: string(pr),
		GameId:         dbInv.GameId,
		PlayerId:       turn.Player.Id,
		MsatsRequested: decoded.Msats,
		Description:    decoded.Description,
	})
	if err != nil {
		return fmt.Errorf("failed to insert invoice: %s", err.Error())
	}

	logf("created next invoice: %s", decoded.PaymentHash)

	return nil
}
