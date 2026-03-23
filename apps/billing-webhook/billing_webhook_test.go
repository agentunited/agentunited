package main

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stripe/stripe-go/v81/webhook"
)

type mockStore struct {
	upserts int
	last    SubscriptionRecord
	err     error
}

func (m *mockStore) UpsertSubscription(_ context.Context, s SubscriptionRecord) error {
	m.upserts++
	m.last = s
	return m.err
}
func (m *mockStore) GetWorkspaceEmail(_ context.Context, _ string) (string, error) {
	return "owner@example.com", nil
}

type mockCache struct {
	processed bool
	marked    bool
	updated   bool
	err       error
}

func (m *mockCache) AlreadyProcessed(_ context.Context, _ string) (bool, error) {
	return m.processed, m.err
}
func (m *mockCache) MarkProcessed(_ context.Context, _ string, _ time.Duration) error {
	m.marked = true
	return nil
}
func (m *mockCache) UpdateRelayPlan(_ context.Context, _ string, _ SubscriptionRecord) error {
	m.updated = true
	return nil
}

type mockEmail struct{ sent bool }

func (m *mockEmail) SendUpgradeConfirmation(_ context.Context, _, _ string) error {
	m.sent = true
	return nil
}

func stripeSig(secret string, body []byte) string {
	signed := webhook.GenerateTestSignedPayload(&webhook.UnsignedPayload{Payload: body, Secret: secret})
	return signed.Header
}

func TestInvalidSignature(t *testing.T) {
	h := NewHandler("whsec_test", &mockStore{}, &mockCache{}, &mockEmail{})
	req := httptest.NewRequest(http.MethodPost, "/webhook/stripe", bytes.NewBufferString(`{"id":"evt_1"}`))
	req.Header.Set("Stripe-Signature", "bad")
	rr := httptest.NewRecorder()
	h.StripeWebhook(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 got %d", rr.Code)
	}
}

func TestIdempotencySkip(t *testing.T) {
	secret := "whsec_test"
	body := []byte(`{"id":"evt_1","type":"checkout.session.completed","data":{"object":{"client_reference_id":"ws_1","metadata":{"plan":"pro"}}}}`)
	store := &mockStore{}
	cache := &mockCache{processed: true}
	h := NewHandler(secret, store, cache, &mockEmail{})
	req := httptest.NewRequest(http.MethodPost, "/webhook/stripe", bytes.NewReader(body))
	req.Header.Set("Stripe-Signature", stripeSig(secret, body))
	rr := httptest.NewRecorder()
	h.StripeWebhook(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", rr.Code)
	}
	if store.upserts != 0 {
		t.Fatalf("expected no DB upsert on duplicate")
	}
}

func TestSubscriptionDeletedSetsFreeCanceled(t *testing.T) {
	secret := "whsec_test"
	body := []byte(`{"id":"evt_2","type":"customer.subscription.deleted","data":{"object":{"id":"sub_1","customer":"cus_1","status":"canceled","metadata":{"workspace_id":"ws_1","plan":"team"}}}}`)
	store := &mockStore{}
	cache := &mockCache{}
	h := NewHandler(secret, store, cache, &mockEmail{})
	req := httptest.NewRequest(http.MethodPost, "/webhook/stripe", bytes.NewReader(body))
	req.Header.Set("Stripe-Signature", stripeSig(secret, body))
	rr := httptest.NewRecorder()
	h.StripeWebhook(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", rr.Code)
	}
	if store.last.Plan != "free" || store.last.Status != "canceled" {
		t.Fatalf("unexpected sub: %+v", store.last)
	}
}

func TestDBErrorReturns500(t *testing.T) {
	secret := "whsec_test"
	body := []byte(`{"id":"evt_3","type":"checkout.session.completed","data":{"object":{"client_reference_id":"ws_1","metadata":{"plan":"pro"}}}}`)
	store := &mockStore{err: context.DeadlineExceeded}
	cache := &mockCache{}
	h := NewHandler(secret, store, cache, &mockEmail{})
	req := httptest.NewRequest(http.MethodPost, "/webhook/stripe", bytes.NewReader(body))
	req.Header.Set("Stripe-Signature", stripeSig(secret, body))
	rr := httptest.NewRecorder()
	h.StripeWebhook(rr, req)
	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 got %d", rr.Code)
	}
}
