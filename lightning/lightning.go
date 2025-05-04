package lightning

import (
	"time"

	decodepay "github.com/nbd-wtf/ln-decodepay"
)

type PaymentRequest string

type Lightning interface {
	CreateInvoice(msats int64, description string, expiresAt time.Time) (PaymentRequest, error)
	GetInvoice(paymentHash string) (*Invoice, error)
	// TODO: PayRequest(pr PaymentRequest) error
}

type Invoice struct {
	PaymentHash    string    `json:"paymentHash"`
	Preimage       string    `json:"preimage"`
	Msats          int64     `json:"msats"`
	Description    string    `json:"description"`
	PaymentRequest string    `json:"paymentRequest"`
	CreatedAt      time.Time `json:"createdAt"`
	ConfirmedAt    time.Time `json:"confirmedAt"`
	ExpiresAt      time.Time `json:"expiresAt"`
	ExpiresIn      int
}

func DecodePaymentRequest(pr PaymentRequest) (*Invoice, error) {
	decoded, err := decodepay.Decodepay(string(pr))
	if err != nil {
		return nil, err
	}

	expiresAt := time.Unix(int64(decoded.CreatedAt+decoded.Expiry), 0)
	expiresIn := int(time.Until(expiresAt).Seconds())

	return &Invoice{
		PaymentHash:    decoded.PaymentHash,
		Msats:          decoded.MSatoshi,
		Description:    decoded.Description,
		PaymentRequest: string(pr),
		CreatedAt:      time.Unix(int64(decoded.CreatedAt), 0),
		ExpiresAt:      expiresAt,
		ExpiresIn:      expiresIn,
	}, nil
}
