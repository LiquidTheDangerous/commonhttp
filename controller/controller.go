package controller

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"

	"github.com/LiquidTheDangerous/commonhttp/middleware"
)

type Registrar interface {
	RegisterController(mux any, c Controller) error
}

type Controller interface {
	Routes() []Route
}

type Route struct {
	Method           string
	Pattern          string
	Handler          any              // Provide WithHandlerMapperFunc to map this value to http.Handler. By default, RegisterController function tries to cast to http.Handler
	HandlerRegistrar HandlerRegistrar // route can provide registrar behaviour. If present, WithHandlerRegistrarFunc Option will be not applied
}

// HandlerRegistrar provides strategy for registering handler with route pattern.
type HandlerRegistrar interface {
	RegisterHandler(mux any, handler http.Handler, route *Route) error
}

func (r HandlerRegistrarFunc) RegisterHandler(mux any, handler http.Handler, route *Route) error {
	return r(mux, handler, route)
}

type HandlerRegistrarFunc func(mux any, handler http.Handler, route *Route) error

// HandlerMapper provides strategy for mapping Route.Handler value to http.Handler interface
type HandlerMapper interface {
	MapHandler(handler any) (http.Handler, error)
}

type HandlerMapperFunc func(handler any) (http.Handler, error)

func (m HandlerMapperFunc) MapHandler(handler any) (http.Handler, error) {
	return m(handler)
}

type HandlerDecorator interface {
	Decorate(h http.Handler) http.Handler
}

type HandlerDecoratorFunc func(handler http.Handler) http.Handler

func (m HandlerDecoratorFunc) Decorate(h http.Handler) http.Handler {
	return m(h)
}

type RegisterControllerOption struct {
	Mapper         HandlerMapper
	RouteRegistrar HandlerRegistrar
	Decorators     []HandlerDecorator
}

func WithHandlerMapperFunc(mapper HandlerMapperFunc) OptionModifier {
	return OptionModifierFunc(func(option *RegisterControllerOption) {
		option.Mapper = mapper
	})
}

func WithHandlerRegistrarFunc(registrar HandlerRegistrarFunc) OptionModifier {
	return OptionModifierFunc(func(option *RegisterControllerOption) {
		option.RouteRegistrar = registrar
	})
}

func WithHandlerDecoratorFunc(decorator HandlerDecoratorFunc) OptionModifier {
	return OptionModifierFunc(func(option *RegisterControllerOption) {
		option.Decorators = append(option.Decorators, decorator)
	})
}

func WithMiddlewares(m ...middleware.Middleware) OptionModifier {
	return OptionModifierFunc(func(option *RegisterControllerOption) {
		decorator := HandlerDecoratorFunc(func(handler http.Handler) http.Handler {
			return middleware.Apply(handler)
		})
		option.Decorators = append(option.Decorators, decorator)
	})
}

type OptionModifier interface {
	Modify(option *RegisterControllerOption)
}

type OptionModifierFunc func(option *RegisterControllerOption)

func (o OptionModifierFunc) Modify(option *RegisterControllerOption) {
	o(option)
}

type defaultControllerRegistrar struct {
	RegisterControllerOption
}

// NewDefaultControllerRegistrar returns default implementation of Registrar.
// Every option will be applied to corresponding controller registration
func NewDefaultControllerRegistrar(options ...OptionModifier) Registrar {
	r := &defaultControllerRegistrar{}
	for _, option := range options {
		option.Modify(&r.RegisterControllerOption)
	}
	return r
}

func (d *defaultControllerRegistrar) RegisterController(mux any, c Controller) error {
	mapper := d.Mapper
	decorators := d.Decorators
	routeRegistrar := d.RouteRegistrar
	routes := c.Routes()
	for _, route := range routes {

		handler, err := mapper.MapHandler(route.Handler)
		if err != nil {
			return err
		}
		if decorators != nil && len(decorators) > 0 {
			handler = decorate(handler, decorators)
		}
		if route.HandlerRegistrar != nil {
			routeRegistrar = route.HandlerRegistrar
		}
		err = routeRegistrar.RegisterHandler(mux, handler, &route)
		if err != nil {
			return err
		}
	}
	return nil
}

func MustRegisterController(mux any, c Controller, options ...OptionModifier) {
	err := RegisterController(mux, c, options...)
	if err != nil {
		panic(err)
	}
}

func RegisterController(mux any, c Controller, options ...OptionModifier) error {
	return NewDefaultControllerRegistrar(options...).RegisterController(mux, c)
}

func decorate(handler http.Handler, decorators []HandlerDecorator) http.Handler {
	for i := len(decorators); i >= 0; i-- {
		handler = decorators[i].Decorate(handler)
	}
	return handler
}

// registerHttpMux provides default implementation for HandlerRegistrar.
// its register http.Handler with pattern '<Route.Method> <Route.Pattern>'.
// works only with http.ServeMux like.
func registerHttpMux(mux any, handler http.Handler, route *Route) error {
	type IMux interface {
		Handle(pattern string, handler http.Handler)
	}
	if m, ok := mux.(IMux); ok {
		m.Handle(fmt.Sprintf("%s %s", route.Method, route.Pattern), handler)
		return nil
	}
	return fmt.Errorf("http.ServeMux like expected, got %T", mux)
}

// mapHandler provides default implementation for HandlerMapper.
// First try to cast to http.Handler, then try to cast to function with signature func(http.ResponseWriter,*http.Request)
func mapHandler(v any) (http.Handler, error) {
	if h, ok := v.(http.Handler); ok {
		return h, nil
	}
	f, signOk := tryReflectFunc(v)
	if signOk {
		return http.HandlerFunc(f), nil
	}
	return nil, errors.New("failed to map handler")
}

// tryReflectFunc checks that v value is a function and its signature matches http.HandlerFunc signature.
func tryReflectFunc(v any) (func(w http.ResponseWriter, r *http.Request), bool) {
	value := reflect.ValueOf(v)
	if value.Kind() != reflect.Func {
		return nil, false
	}
	arg0 := value.Type().In(0)
	if reflect.TypeFor[http.ResponseWriter]() != arg0 {
		return nil, false
	}
	arg1 := value.Type().In(1)
	if arg1.Kind() != reflect.Ptr {
		return nil, false
	}
	arg1 = arg1.Elem()
	if arg1 != reflect.TypeFor[http.Request]() {
		return nil, false
	}
	return value.Interface().(func(w http.ResponseWriter, r *http.Request)), true
}
