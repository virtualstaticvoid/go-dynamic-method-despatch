package main

import (
	"errors"
	"fmt"
	msg "go.virtualstaticvoid.com/methoddespatch/messages"
	"reflect"
)

func main() {
	var err error

	client := NewSubscriber()

	err = client.SubscribeTo(&msg.Message1{}, HandleMessage1)
	if err != nil {
		fmt.Printf("Error %q\n", err)
	}
	err = client.SubscribeTo(&msg.Message2{}, HandleMessage2)
	if err != nil {
		fmt.Printf("Error %q\n", err)
	}

	err = client.SubscribeTo(&msg.Message1{}, HandleMessageInvalidArgType)
	if err != nil {
		fmt.Printf("Error %q\n", err)
	}

	err = client.SubscribeTo(&msg.Message1{}, HandleMessageInvalidReturn)
	if err != nil {
		fmt.Printf("Error %q\n", err)
	}

	err = client.Publish(&msg.Message1{SomeData: "foobar"})
	if err != nil {
		fmt.Printf("Error %q\n", err)
	}

	err = client.Publish(&msg.Message2{Other: 99})
	if err != nil {
		fmt.Printf("Error %q\n", err)
	}

	err = client.Publish(&msg.Message3{Never: 191919})
	if err != nil {
		fmt.Printf("Error %q\n", err)
	}

}

type Subscriber interface {
	SubscribeTo(message interface{}, handler interface{}) error
	Publish(message interface{}) error
}

type subscriber struct {
	handlers map[reflect.Type]interface{}
}

func NewSubscriber() Subscriber {
	s := &subscriber{}
	s.handlers = make(map[reflect.Type]interface{})
	return s
}

func (s *subscriber) SubscribeTo(message interface{}, handler interface{}) error {
	messageType := reflect.TypeOf(message)
	handlerType := reflect.TypeOf(handler)

	// handler must be a "func(messageType) error"
	if handlerType.Kind() == reflect.Func &&
		handlerType.NumIn() == 1 && handlerType.NumOut() == 1 &&
		messageType.AssignableTo(handlerType.In(0)) &&
		reflect.TypeOf(errors.New("")).AssignableTo(handlerType.Out(0)) {

		// passed contract check!
		s.handlers[messageType] = handler
		return nil
	}
	return fmt.Errorf("%q cannot be used to handle messages of type %q", handlerType, messageType)
}

func (s *subscriber) Publish(message interface{}) error {
	messageType := reflect.TypeOf(message)

	handler := s.handlers[messageType]
	if handler == nil {
		return fmt.Errorf("no handler found for message type %q", messageType)
	}

	//
	// dynamically call the method
	//
	// NOTE: this is "unsafe" since there is no guarantee that:
	// 1. the method takes 1 argument
	// 2. the type of the argument is the same as messageType
	// 3. the method returns an error
	//
	result := reflect.ValueOf(handler).Call([]reflect.Value{reflect.ValueOf(message)})

	// need to explicitly check for nil, otherwise type cast to error can fail
	if result[0].IsNil() {
		return nil
	}
	return result[0].Interface().(error)
}

func HandleMessage1(m *msg.Message1) error {
	fmt.Printf("HandleMessage1 received message %q\n", m.SomeData)
	return nil
}

func HandleMessage2(m *msg.Message2) error {
	fmt.Printf("HandleMessage2 received message %d\n", m.Other)
	return fmt.Errorf("example error")
}

func HandleMessageInvalidArgType(m string) error {
	return nil
}

func HandleMessageInvalidReturn(m *msg.Message2) {
	// no-op
}
