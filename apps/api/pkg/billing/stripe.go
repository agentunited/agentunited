package billing

import (
	"context"
	"time"

	"github.com/stripe/stripe-go/v81"
	portalsession "github.com/stripe/stripe-go/v81/billingportal/session"
	checkoutsession "github.com/stripe/stripe-go/v81/checkout/session"
	"github.com/stripe/stripe-go/v81/customer"
	stripeSubscription "github.com/stripe/stripe-go/v81/subscription"
)

// StripeProvider implements billing.Service using Stripe API
type StripeProvider struct {
	webhookSecret string
}

// NewStripeProvider creates a new Stripe billing provider
func NewStripeProvider(secretKey, webhookSecret string) *StripeProvider {
	stripe.Key = secretKey
	return &StripeProvider{webhookSecret: webhookSecret}
}

// CreateCustomer creates a new Stripe customer
func (p *StripeProvider) CreateCustomer(ctx context.Context, email, name string) (string, error) {
	params := &stripe.CustomerParams{
		Email: &email,
		Name:  &name,
	}
	c, err := customer.New(params)
	if err != nil {
		return "", err
	}
	return c.ID, nil
}

// CreateCheckoutSession creates a Stripe checkout session for subscription
func (p *StripeProvider) CreateCheckoutSession(ctx context.Context, customerID, priceID, successURL, cancelURL string) (string, error) {
	params := &stripe.CheckoutSessionParams{
		Customer:   &customerID,
		SuccessURL: &successURL,
		CancelURL:  &cancelURL,
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    &priceID,
				Quantity: stripe.Int64(1),
			},
		},
		Mode: stripe.String(string(stripe.CheckoutSessionModeSubscription)),
	}
	s, err := checkoutsession.New(params)
	if err != nil {
		return "", err
	}
	return s.URL, nil
}

// GetPortalURL generates a customer billing portal URL
func (p *StripeProvider) GetPortalURL(ctx context.Context, customerID, returnURL string) (string, error) {
	params := &stripe.BillingPortalSessionParams{
		Customer:  &customerID,
		ReturnURL: &returnURL,
	}
	s, err := portalsession.New(params)
	if err != nil {
		return "", err
	}
	return s.URL, nil
}

// GetSubscription retrieves subscription details from Stripe
func (p *StripeProvider) GetSubscription(ctx context.Context, subscriptionID string) (*Subscription, error) {
	s, err := stripeSubscription.Get(subscriptionID, nil)
	if err != nil {
		return nil, err
	}

	var periodEnd *time.Time
	if s.CurrentPeriodEnd > 0 {
		t := time.Unix(s.CurrentPeriodEnd, 0)
		periodEnd = &t
	}

	plan := "free"
	if len(s.Items.Data) > 0 && s.Items.Data[0].Price != nil {
		plan = s.Items.Data[0].Price.LookupKey
		if plan == "" {
			plan = string(s.Items.Data[0].Price.Recurring.Interval)
		}
	}

	return &Subscription{
		ID:               s.ID,
		CustomerID:       s.Customer.ID,
		SubscriptionID:   s.ID,
		Plan:             plan,
		Status:           string(s.Status),
		CurrentPeriodEnd: periodEnd,
	}, nil
}
