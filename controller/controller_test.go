package controller

import (
	"net/http"
	"testing"
)

type ControllerWithFunction struct {
}

func (s *ControllerWithFunction) Routes() []Route {
	return []Route{{"GET", "", func(w http.ResponseWriter, r *http.Request) {

	}, nil}}
}

type MockMux struct {
	onRegister func(pattern string, handler http.Handler)
}

func (s *MockMux) Handle(pattern string, handler http.Handler) {

	s.onRegister(pattern, handler)
}

func TestShouldRegisterControllerWithFunction(t *testing.T) {
	var registeredPattern string
	var registeredHandler http.Handler
	mux := &MockMux{
		onRegister: func(pattern string, handler http.Handler) {
			registeredPattern = pattern
			registeredHandler = handler
		},
	}

	err := RegisterController(mux, &ControllerWithFunction{})
	if err != nil {
		t.Errorf("error registering controller: %v", err)
	}
	if registeredPattern != "GET /api/function" {
		t.Errorf("expected 'GET /api/function', got '%s'", registeredPattern)
	}
	if registeredHandler == nil {
		t.Error("expected a handler to be registered")
	}
}
