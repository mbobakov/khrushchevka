// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package live

import (
	"github.com/mbobakov/khrushchevka/internal"
	"sync"
)

// Ensure, that LightsControllerMock does implement LightsController.
// If this is not the case, regenerate this file with moq.
var _ LightsController = &LightsControllerMock{}

// LightsControllerMock is a mock implementation of LightsController.
//
//	func TestSomethingThatUsesLightsController(t *testing.T) {
//
//		// make and configure a mocked LightsController
//		mockedLightsController := &LightsControllerMock{
//			SetFunc: func(addr internal.LightAddress, isON bool) error {
//				panic("mock out the Set method")
//			},
//		}
//
//		// use mockedLightsController in code that requires LightsController
//		// and then make assertions.
//
//	}
type LightsControllerMock struct {
	// SetFunc mocks the Set method.
	SetFunc func(addr internal.LightAddress, isON bool) error

	// calls tracks calls to the methods.
	calls struct {
		// Set holds details about calls to the Set method.
		Set []struct {
			// Addr is the addr argument value.
			Addr internal.LightAddress
			// IsON is the isON argument value.
			IsON bool
		}
	}
	lockSet sync.RWMutex
}

// Set calls SetFunc.
func (mock *LightsControllerMock) Set(addr internal.LightAddress, isON bool) error {
	if mock.SetFunc == nil {
		panic("LightsControllerMock.SetFunc: method is nil but LightsController.Set was just called")
	}
	callInfo := struct {
		Addr internal.LightAddress
		IsON bool
	}{
		Addr: addr,
		IsON: isON,
	}
	mock.lockSet.Lock()
	mock.calls.Set = append(mock.calls.Set, callInfo)
	mock.lockSet.Unlock()
	return mock.SetFunc(addr, isON)
}

// SetCalls gets all the calls that were made to Set.
// Check the length with:
//
//	len(mockedLightsController.SetCalls())
func (mock *LightsControllerMock) SetCalls() []struct {
	Addr internal.LightAddress
	IsON bool
} {
	var calls []struct {
		Addr internal.LightAddress
		IsON bool
	}
	mock.lockSet.RLock()
	calls = mock.calls.Set
	mock.lockSet.RUnlock()
	return calls
}
