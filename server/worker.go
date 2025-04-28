package server

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/ekzyis/zapback/db"
	"github.com/ekzyis/zapback/env"
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

			inv, err := sCtx.Ln.GetInvoice(invoice.PaymentHash)
			if err != nil {
				logf("failed to check invoice: %s: %s", invoice.PaymentHash, err.Error())
				continue
			}

			if !inv.ConfirmedAt.IsZero() || (env.Debug && time.Since(inv.CreatedAt) >= 5*time.Second) {
				logf("invoice confirmed: %s", invoice.PaymentHash)
				sCtx.Db.UpdateInvoice(invoice.Id, &db.UpdateInvoice{
					ConfirmedAt: sql.NullTime{
						Time:  inv.ConfirmedAt,
						Valid: true,
					},
				})
				continue
			}

			logf("invoice pending: %s", invoice.PaymentHash)
		}

		time.Sleep(5 * time.Second)
	}
}
