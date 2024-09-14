package middleware

import (
	"net/http"
	"sync/atomic"
	"testing"
)

type mockHandler struct {
	interactionsCount atomic.Int32
	trackOrder        *atomic.Int32
	lastInvokedOrder  int32
	onServeHTTP       http.HandlerFunc
}

func (h *mockHandler) numInteractions() int32 {
	return h.interactionsCount.Load()
}

func (h *mockHandler) order() int32 {
	return h.lastInvokedOrder
}

func (h *mockHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.trackOrder != nil {
		h.lastInvokedOrder = h.trackOrder.Load()
		h.trackOrder.Add(1)
	}
	h.interactionsCount.Add(1)
	h.onServeHTTP(w, r)
}

type mockMiddleware struct {
	interactionsCount atomic.Int32
	onHandle          MiddlewareFunc
	lastInvokedOrder  int32
	trackOrder        *atomic.Int32
}

func (m *mockMiddleware) Handle(w http.ResponseWriter, r *http.Request, next http.Handler) {
	if m.trackOrder != nil {
		m.lastInvokedOrder = m.trackOrder.Load()
		m.trackOrder.Add(1)
	}
	m.interactionsCount.Add(1)
	m.onHandle(w, r, next)
}

func (m *mockMiddleware) numInteractions() int32 {
	return m.interactionsCount.Load()
}

func (m *mockMiddleware) order() int32 {
	return m.lastInvokedOrder
}

func assertInteractions(t *testing.T, interactable interface{ numInteractions() int32 }, expected int32) {
	if interactable.numInteractions() != expected {
		t.Fatalf("expected %d interactions, got %d, object: %T %v", expected, interactable.numInteractions(), interactable, interactable)
	}
}

func assertNoInteractions(t *testing.T, interactables ...interface{ numInteractions() int32 }) {
	for _, interactable := range interactables {
		assertInteractions(t, interactable, 0)
	}
}

func assertExactlyOneInteraction(t *testing.T, interactables ...interface{ numInteractions() int32 }) {
	for _, interactable := range interactables {
		assertInteractions(t, interactable, 1)
	}
}

func assertOrder(t *testing.T, orderable interface{ order() int32 }, expected int32) {
	if orderable.order() != expected {
		t.Fatalf("expected order to be %d, got %d", expected, orderable.order())
	}
}

func TestShouldInvokeInvokeInitialHandler(t *testing.T) {
	ord := atomic.Int32{}
	h := &mockHandler{
		trackOrder: &ord,
		onServeHTTP: func(writer http.ResponseWriter, request *http.Request) {
		}}
	m := &mockMiddleware{
		trackOrder: &ord,
		onHandle: func(w http.ResponseWriter, r *http.Request, next http.Handler) {
			assertNoInteractions(t, h)
			next.ServeHTTP(w, r)
			assertExactlyOneInteraction(t, h)
		},
	}
	handler := Apply(h, m)
	handler.ServeHTTP(nil, nil)
	assertExactlyOneInteraction(t, m)
	assertExactlyOneInteraction(t, h)
	assertOrder(t, m, 0)
	assertOrder(t, h, 1)
}

func TestShouldInvokeWithCorrectOrder(t *testing.T) {
	ord := atomic.Int32{}
	var (
		m0, m1, m2 *mockMiddleware
		h          *mockHandler
	)
	m0 = &mockMiddleware{
		trackOrder: &ord,
		onHandle: func(w http.ResponseWriter, r *http.Request, next http.Handler) {
			assertNoInteractions(t, m1, m2, h)
			next.ServeHTTP(w, r)
			assertExactlyOneInteraction(t, m1, m2, h)
		},
	}
	m1 = &mockMiddleware{
		trackOrder: &ord,
		onHandle: func(w http.ResponseWriter, r *http.Request, next http.Handler) {
			assertExactlyOneInteraction(t, m0)
			assertNoInteractions(t, m2, h)
			next.ServeHTTP(w, r)
			assertExactlyOneInteraction(t, m2, h)
		},
	}
	m2 = &mockMiddleware{
		trackOrder: &ord,
		onHandle: func(w http.ResponseWriter, r *http.Request, next http.Handler) {
			assertExactlyOneInteraction(t, m0, m1)
			assertNoInteractions(t, h)
			next.ServeHTTP(w, r)
			assertExactlyOneInteraction(t, h)
		},
	}
	h = &mockHandler{
		trackOrder: &ord,
		onServeHTTP: func(w http.ResponseWriter, r *http.Request) {
			assertExactlyOneInteraction(t, m0, m1, m2)
		},
	}
	handler := Apply(h, m0, m1, m2)
	handler.ServeHTTP(nil, nil)
	assertExactlyOneInteraction(t, m0, m1, m2)
	assertOrder(t, m0, 0)
	assertOrder(t, m1, 1)
	assertOrder(t, m2, 2)
	assertOrder(t, h, 3)
}
