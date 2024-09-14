package middleware

import "net/http"

// Middleware can be used to attach additional behaviour for http.Handler
type Middleware interface {
	Handle(w http.ResponseWriter, r *http.Request, next http.Handler)
}

// middlewareChain Redirects ServeHTTP call to Middleware.Handle method with nextHandler.
// Given middleware should attach additional logic and call ServeHTTP method on provided nextHandler.
// middlewareChain implements http.Handler, so next middleware can also be middlewareChain.
type middlewareChain struct {
	nextHandler http.Handler
	middleware  Middleware
}

func newMiddlewareChain(handler http.Handler, middleware Middleware) *middlewareChain {
	return &middlewareChain{
		nextHandler: handler,
		middleware:  middleware,
	}
}

func (m *middlewareChain) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.middleware.Handle(w, r, m.nextHandler)
}

// Apply returns handler with applied middlewares chain.
func Apply(handler http.Handler, middleware ...Middleware) http.Handler {
	current := handler
	for i := len(middleware) - 1; i >= 0; i-- {
		current = newMiddlewareChain(current, middleware[i])
	}
	return current
}
