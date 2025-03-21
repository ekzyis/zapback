package lightning

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type Voltage struct {
	url            *url.URL
	organizationId string
	envId          string
	walletId       string
	apiKey         string
}

var _ Lightning = (*Voltage)(nil)

type VoltageCreateInvoiceParams struct {
	AmountMsats int64  `json:"amount_msats"`
	Currency    string `json:"currency"`
	Description string `json:"description"`
	Id          string `json:"id"`
	PaymentKind string `json:"payment_kind"`
	WalletId    string `json:"wallet_id"`
}

type VoltageCreateInvoiceResponse struct {
	Bip21Uri  string `json:"bip21_uri"`
	CreatedAt string `json:"created_at"`
	Currency  string `json:"currency"`
	Data      struct {
		AmountMsats    int64  `json:"amount_msats"`
		Memo           string `json:"memo"`
		PaymentRequest string `json:"payment_request"`
	} `json:"data"`
	Direction      string `json:"direction"`
	EnvironmentId  string `json:"environment_id"`
	Error          any    `json:"error"`
	Id             string `json:"id"`
	OrganizationId string `json:"organization_id"`
	Status         string `json:"status"`
	Type           string `json:"type"`
	UpdatedAt      string `json:"updated_at"`
	WalletId       string `json:"wallet_id"`
}

func NewVoltage(opts ...func(*Voltage) *Voltage) *Voltage {
	v := &Voltage{}
	for _, opt := range opts {
		opt(v)
	}
	return v
}

func WithVoltageUrl(u string) func(*Voltage) *Voltage {
	return func(v *Voltage) *Voltage {
		u, err := url.Parse(u)
		if err != nil {
			log.Fatal(err)
		}
		v.url = u
		return v
	}
}

func WithVoltageOrganizationId(organizationId string) func(*Voltage) *Voltage {
	return func(v *Voltage) *Voltage {
		v.organizationId = organizationId
		return v
	}
}

func WithVoltageEnvId(envId string) func(*Voltage) *Voltage {
	return func(v *Voltage) *Voltage {
		v.envId = envId
		return v
	}
}

func WithVoltageWalletId(walletId string) func(*Voltage) *Voltage {
	return func(v *Voltage) *Voltage {
		v.walletId = walletId
		return v
	}
}

func WithVoltageApiKey(apiKey string) func(*Voltage) *Voltage {
	return func(v *Voltage) *Voltage {
		v.apiKey = apiKey
		return v
	}
}

func (v *Voltage) CreateInvoice(msats int64, description string) (PaymentRequest, error) {
	params := VoltageCreateInvoiceParams{
		AmountMsats: msats,
		Currency:    "btc",
		Description: description,
		Id:          uuid.New().String(),
		PaymentKind: "bolt11",
		WalletId:    v.walletId,
	}

	jsonData, err := json.Marshal(params)
	if err != nil {
		return "", err
	}

	endpoint := v.url.JoinPath(
		"organizations", v.organizationId,
		"environments", v.envId,
		"payments")

	req, err := http.NewRequest("POST", endpoint.String(), bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Api-Key", v.apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		return "", fmt.Errorf("voltage %s", resp.Status)
	}

	var invoiceResp *VoltageCreateInvoiceResponse
	for {
		time.Sleep(1 * time.Second)
		invoiceResp, err = v.getInvoice(params.Id)
		if err != nil {
			log.Println("voltage:", err)
			continue
		}
		if invoiceResp.Data.PaymentRequest == "" {
			// XXX voltage can return HTTP/200 but not include the payment request yet, lol
			log.Println("voltage: no payment request yet")
			continue
		}
		break
	}

	return PaymentRequest(invoiceResp.Data.PaymentRequest), nil
}

func (v *Voltage) GetInvoice(paymentHash string) (*Invoice, error) {
	// TODO: implement
	return nil, nil
}

func (v *Voltage) getInvoice(paymentId string) (*VoltageCreateInvoiceResponse, error) {
	endpoint := v.url.JoinPath(
		"organizations", v.organizationId,
		"environments", v.envId,
		"payments", paymentId)

	req, err := http.NewRequest("GET", endpoint.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Api-Key", v.apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, echo.NewHTTPError(resp.StatusCode)
	}

	var invoiceResp *VoltageCreateInvoiceResponse
	if err := json.Unmarshal(body, &invoiceResp); err != nil {
		return nil, err
	}

	return invoiceResp, nil
}
