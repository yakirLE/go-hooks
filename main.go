package main

import (
	"fmt"
	"reflect"
)

// ================ main service ======================
type ProxyService struct {
	internalService *InternalService
	wrapper         *ServiceWrapper
}

func NewService(f1 string, f2 int) *ProxyService {
	svc := InternalService{
		FieldA: f1,
		FieldB: f2,
	}

	wrapper := ServiceWrapper{&svc}

	return &ProxyService{
		internalService: &svc,
		wrapper:         &wrapper,
	}
}

func (s ProxyService) DoSomethingA(a string) error {
	res := invokeMethodWithHooks(s.wrapper, "InternalDoSomethingA", "hookValueA", a)
	err := convertReflectValue[error](res[0])
	if err == nil {
		return nil
	}

	return *err
}

func (s ProxyService) DoSomethingB() (string, error) {
	res := invokeMethodWithHooks(s.wrapper, "InternalDoSomethingB", "hookValueB")
	str := convertReflectValue[string](res[0])
	err := convertReflectValue[error](res[1])
	var rerr error
	if err != nil {
		rerr = *err
	}

	return *str, rerr
}

func (s ProxyService) DoSomethingC(a int) {
	invokeMethodWithHooks(s.wrapper, "InternalDoSomethingC", "hookValueC", a)
}

// ===================================================

func invokeMethodWithHooks(hookable Hookable, methodName string, args ...any) []reflect.Value {
	method := reflect.ValueOf(hookable).MethodByName(methodName)
	hookable.BeforeHook(args[0].(string))
	args = args[1:]
	inputArgs := make([]reflect.Value, len(args))
	for i, arg := range args {
		inputArgs[i] = reflect.ValueOf(arg)
	}

	result := method.Call(inputArgs)
	hookable.AfterHook()

	return result
}

func convertReflectValue[T any](value reflect.Value) *T {
	if value.Kind() == reflect.Invalid ||
		!value.IsValid() ||
		(value.Kind() == reflect.Interface && value.IsNil()) {
		return nil
	}

	actualValue := value.Interface().(T)
	return &actualValue
}

// =============== internal service ===================
type InternalService struct {
	FieldA string
	FieldB int
}

func (s InternalService) InternalDoSomethingA(a string) error {
	fmt.Printf("did A with %s\n", a)
	return nil
}

func (s InternalService) InternalDoSomethingB() (string, error) {
	fmt.Println("did B")
	return "bye", fmt.Errorf("error B")
}

func (s InternalService) InternalDoSomethingC(a int) {
	fmt.Println("did C with", a)
}

// ====================================================

// ================ hookable wrapper ==================
type ServiceWrapper struct {
	*InternalService
}

func (w *ServiceWrapper) BeforeHook(s string) {
	println("this is before hook", s)
}

func (w *ServiceWrapper) AfterHook() {
	println("this is after hook")
}

type Hookable interface {
	BeforeHook(string)
	AfterHook()
}

// ====================================================

func main() {
	myService := NewService("yakir", 33)
	err := myService.DoSomethingA("levi")
	fmt.Println("A error", err)

	s, err := myService.DoSomethingB()
	fmt.Println("B string", s, "error", err)

	myService.DoSomethingC(2023)
}
