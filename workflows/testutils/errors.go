package testutils

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/kr/pretty"
	"reflect"

	"go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/temporal"
)

const (
	WorkflowCanceledErrType = "Canceled"
	WorkflowTimeoutErrType  = "Timeout"
	WorkflowPanicErrType    = "Panic"
	WorkflowUnknownErrType  = "Unknown"
)

func GetErrorType(err error) string {
	if err == nil {
		return ""
	}
	var appErr *temporal.ApplicationError
	if errors.As(err, &appErr) {
		return appErr.Type()
	}

	//var actErr *temporal.ActivityError
	//if errors.As(err, &actErr) {
	//	return actErr.Unwrap()
	//}

	var panicErr *temporal.PanicError
	if errors.As(err, &panicErr) {
		return WorkflowPanicErrType
	}

	var timeoutErr *temporal.TimeoutError
	if errors.As(err, &timeoutErr) {
		// handle timeout, could check timeout type by timeoutErr.TimeoutType()
		switch timeoutErr.TimeoutType() {
		case enums.TIMEOUT_TYPE_SCHEDULE_TO_START:
			// Handle ScheduleToStart timeout.
		case enums.TIMEOUT_TYPE_START_TO_CLOSE:
			// Handle StartToClose timeout.
		case enums.TIMEOUT_TYPE_HEARTBEAT:
			// Handle heartbeat timeout.
		default:
		}
		return WorkflowTimeoutErrType
	}

	var cancelErr *temporal.CanceledError
	if errors.As(err, &cancelErr) {
		return WorkflowCanceledErrType
	}

	return WorkflowUnknownErrType
}

type arrayHaving struct {
	expected interface{}
}

func (a *arrayHaving) Matches(x interface{}) bool {
	match, err := elementsMatch(a.expected, x)
	if err != nil {
		return false
	}
	return match
}

func (a *arrayHaving) String() string {
	return "array having elements: " + pretty.Sprint(a.expected)
}

func ArrayHaving(expected interface{}) gomock.Matcher {
	return &arrayHaving{expected: expected}
}

func elementsMatch(listA, listB interface{}) (ok bool, err error) {
	if isEmpty(listA) && isEmpty(listB) {
		return true, nil
	}

	aKind := reflect.TypeOf(listA).Kind()
	bKind := reflect.TypeOf(listB).Kind()

	if aKind != reflect.Array && aKind != reflect.Slice {
		return false, fmt.Errorf("%q has an unsupported type %s", listA, aKind)
	}

	if bKind != reflect.Array && bKind != reflect.Slice {
		return false, fmt.Errorf("%q has an unsupported type %s", listB, bKind)
	}

	aValue := reflect.ValueOf(listA)
	bValue := reflect.ValueOf(listB)

	aLen := aValue.Len()
	bLen := bValue.Len()

	if aLen != bLen {
		return false, nil
	}

	// Mark indexes in bValue that we already used
	visited := make([]bool, bLen)
	for i := 0; i < aLen; i++ {
		element := aValue.Index(i).Interface()
		found := false
		for j := 0; j < bLen; j++ {
			if visited[j] {
				continue
			}
			if ObjectsAreEqual(bValue.Index(j).Interface(), element) {
				visited[j] = true
				found = true
				break
			}
		}
		if !found {
			return false, nil
		}
	}

	return true, nil
}

// isEmpty gets whether the specified object is considered empty or not.
func isEmpty(object interface{}) bool {

	// get nil case out of the way
	if object == nil {
		return true
	}

	objValue := reflect.ValueOf(object)

	switch objValue.Kind() {
	// collection types are empty when they have no element
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Slice:
		return objValue.Len() == 0
		// pointers are empty if nil or if the value they point to is empty
	case reflect.Ptr:
		if objValue.IsNil() {
			return true
		}
		deref := objValue.Elem().Interface()
		return isEmpty(deref)
		// for all other types, compare against the zero value
	default:
		zero := reflect.Zero(objValue.Type())
		return reflect.DeepEqual(object, zero.Interface())
	}
}

// ObjectsAreEqual determines if two objects are considered equal.
//
// This function does no assertion of any kind.
func ObjectsAreEqual(expected, actual interface{}) bool {
	if expected == nil || actual == nil {
		return expected == actual
	}

	exp, ok := expected.([]byte)
	if !ok {
		return reflect.DeepEqual(expected, actual)
	}

	act, ok := actual.([]byte)
	if !ok {
		return false
	}
	if exp == nil || act == nil {
		return exp == nil && act == nil
	}
	return bytes.Equal(exp, act)
}
