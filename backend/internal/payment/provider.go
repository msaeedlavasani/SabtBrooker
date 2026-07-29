package payment

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

type PaymentProvider interface {
	Request(ctx context.Context, amount int64, description, callbackURL, mobile, email string) (token, paymentURL string, err error)
	Verify(ctx context.Context, amount int64, token string) (refID string, err error)
}

type ZarinpalProvider struct {
	merchantID string
	sandbox    bool
	client     *http.Client
}

func NewZarinpalProvider(merchantID string, sandbox bool) *ZarinpalProvider {
	return &ZarinpalProvider{
		merchantID: merchantID,
		sandbox:    sandbox,
		client:     &http.Client{Timeout: 10 * time.Second},
	}
}

// Mocking Zarinpal for now as we don't have real keys, but structure is correct
func (p *ZarinpalProvider) Request(ctx context.Context, amount int64, description, callbackURL, mobile, email string) (string, string, error) {
	token := fmt.Sprintf("mock_token_%d", time.Now().Unix())
	url := fmt.Sprintf("https://www.zarinpal.com/pg/StartPay/%s", token)
	if p.sandbox {
		url = fmt.Sprintf("https://sandbox.zarinpal.com/pg/StartPay/%s", token)
	}
	return token, url, nil
}

func (p *ZarinpalProvider) Verify(ctx context.Context, amount int64, token string) (string, error) {
	return fmt.Sprintf("ref_%d", time.Now().Unix()), nil
}
