package controller

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
)

type Controller interface {
	Routes() []Route
}

type Route struct {
	Method  string
	Pattern string
	Handler any // http.Handler compatible
}

// HandlerRegistrar provides strategy for registering handler with route pattern.
type HandlerRegistrar interface {
	RegisterHandler(mux any, handler http.Handler, route *Route) error
}

// HandlerMapper provides strategy for mapping Route.Handler value to http.Handler interface
type HandlerMapper interface {
	MapHandler(handler any) (http.Handler, error)
}

type HandlerMapperFunc func(handler any) (http.Handler, error)

type HandlerRegistrarFunc func(mux any, handler http.Handler, route *Route) error

func (r HandlerRegistrarFunc) RegisterHandler(mux any, handler http.Handler, route *Route) error {
	return r(mux, handler, route)
}

func (m HandlerMapperFunc) MapHandler(handler any) (http.Handler, error) {
	return m(handler)
}

type RegisterControllerOption struct {
	Mapper    HandlerMapper
	Registrar HandlerRegistrar
}

func WithHandlerMapperFunc(mapper HandlerMapperFunc) OptionModifier {
	return OptionModifierFunc(func(option *RegisterControllerOption) {
		option.Mapper = mapper
	})
}

func WithHandlerRegistrarFunc(registrar HandlerRegistrarFunc) OptionModifier {
	return OptionModifierFunc(func(option *RegisterControllerOption) {
		option.Registrar = registrar
	})
}

type OptionModifier interface {
	Modify(option *RegisterControllerOption)
}

type OptionModifierFunc func(option *RegisterControllerOption)

func (o OptionModifierFunc) Modify(option *RegisterControllerOption) {
	o(option)
}

func MustRegisterController(mux any, c Controller, options ...OptionModifier) {
	err := RegisterController(mux, c, options...)
	if err != nil {
		panic(err)
	}
}

func RegisterController(mux any, c Controller, options ...OptionModifier) error {
	option := &RegisterControllerOption{}
	for _, modifier := range options {
		modifier.Modify(option)
	}
	if option.Mapper == nil {
		option.Mapper = HandlerMapperFunc(mapHandler)
	}
	if option.Registrar == nil {
		option.Registrar = HandlerRegistrarFunc(registerHttpMux)
	}
	return registerController(mux, option.Mapper, option.Registrar, c)
}

func registerController(mux any, mapper HandlerMapper, routeRegistrar HandlerRegistrar, c Controller) error {
	routes := c.Routes()
	for _, route := range routes {
		handler, err := mapper.MapHandler(route.Handler)
		if err != nil {
			return err
		}
		err = routeRegistrar.RegisterHandler(mux, handler, &route)
		if err != nil {
			return err
		}
	}
	return nil
}

// registerHttpMux provides default implementation for HandlerRegistrar.
func registerHttpMux(mux any, handler http.Handler, route *Route) error {
	type Imux interface {
		Handle(pattern string, handler http.Handler)
	}
	if m, ok := mux.(Imux); ok {
		m.Handle(fmt.Sprintf("%s %s", route.Method, route.Pattern), handler)
		return nil
	}
	return fmt.Errorf("http.ServeMux expected, got %T", mux)
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
