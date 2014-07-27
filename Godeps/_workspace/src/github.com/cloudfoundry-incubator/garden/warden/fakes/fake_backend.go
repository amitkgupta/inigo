// This file was generated by counterfeiter
package fakes

import (
	"sync"
	"time"

	"github.com/cloudfoundry-incubator/garden/warden"
)

type FakeBackend struct {
	PingStub        func() error
	pingMutex       sync.RWMutex
	pingArgsForCall []struct{}
	pingReturns struct {
		result1 error
	}
	CapacityStub        func() (warden.Capacity, error)
	capacityMutex       sync.RWMutex
	capacityArgsForCall []struct{}
	capacityReturns struct {
		result1 warden.Capacity
		result2 error
	}
	CreateStub        func(warden.ContainerSpec) (warden.Container, error)
	createMutex       sync.RWMutex
	createArgsForCall []struct {
		arg1 warden.ContainerSpec
	}
	createReturns struct {
		result1 warden.Container
		result2 error
	}
	DestroyStub        func(handle string) error
	destroyMutex       sync.RWMutex
	destroyArgsForCall []struct {
		handle string
	}
	destroyReturns struct {
		result1 error
	}
	ContainersStub        func(warden.Properties) ([]warden.Container, error)
	containersMutex       sync.RWMutex
	containersArgsForCall []struct {
		arg1 warden.Properties
	}
	containersReturns struct {
		result1 []warden.Container
		result2 error
	}
	LookupStub        func(handle string) (warden.Container, error)
	lookupMutex       sync.RWMutex
	lookupArgsForCall []struct {
		handle string
	}
	lookupReturns struct {
		result1 warden.Container
		result2 error
	}
	StartStub        func() error
	startMutex       sync.RWMutex
	startArgsForCall []struct{}
	startReturns struct {
		result1 error
	}
	StopStub        func()
	stopMutex       sync.RWMutex
	stopArgsForCall []struct{}
	GraceTimeStub        func(warden.Container) time.Duration
	graceTimeMutex       sync.RWMutex
	graceTimeArgsForCall []struct {
		arg1 warden.Container
	}
	graceTimeReturns struct {
		result1 time.Duration
	}
}

func (fake *FakeBackend) Ping() error {
	fake.pingMutex.Lock()
	defer fake.pingMutex.Unlock()
	fake.pingArgsForCall = append(fake.pingArgsForCall, struct{}{})
	if fake.PingStub != nil {
		return fake.PingStub()
	} else {
		return fake.pingReturns.result1
	}
}

func (fake *FakeBackend) PingCallCount() int {
	fake.pingMutex.RLock()
	defer fake.pingMutex.RUnlock()
	return len(fake.pingArgsForCall)
}

func (fake *FakeBackend) PingReturns(result1 error) {
	fake.pingReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeBackend) Capacity() (warden.Capacity, error) {
	fake.capacityMutex.Lock()
	defer fake.capacityMutex.Unlock()
	fake.capacityArgsForCall = append(fake.capacityArgsForCall, struct{}{})
	if fake.CapacityStub != nil {
		return fake.CapacityStub()
	} else {
		return fake.capacityReturns.result1, fake.capacityReturns.result2
	}
}

func (fake *FakeBackend) CapacityCallCount() int {
	fake.capacityMutex.RLock()
	defer fake.capacityMutex.RUnlock()
	return len(fake.capacityArgsForCall)
}

func (fake *FakeBackend) CapacityReturns(result1 warden.Capacity, result2 error) {
	fake.capacityReturns = struct {
		result1 warden.Capacity
		result2 error
	}{result1, result2}
}

func (fake *FakeBackend) Create(arg1 warden.ContainerSpec) (warden.Container, error) {
	fake.createMutex.Lock()
	defer fake.createMutex.Unlock()
	fake.createArgsForCall = append(fake.createArgsForCall, struct {
		arg1 warden.ContainerSpec
	}{arg1})
	if fake.CreateStub != nil {
		return fake.CreateStub(arg1)
	} else {
		return fake.createReturns.result1, fake.createReturns.result2
	}
}

func (fake *FakeBackend) CreateCallCount() int {
	fake.createMutex.RLock()
	defer fake.createMutex.RUnlock()
	return len(fake.createArgsForCall)
}

func (fake *FakeBackend) CreateArgsForCall(i int) warden.ContainerSpec {
	fake.createMutex.RLock()
	defer fake.createMutex.RUnlock()
	return fake.createArgsForCall[i].arg1
}

func (fake *FakeBackend) CreateReturns(result1 warden.Container, result2 error) {
	fake.createReturns = struct {
		result1 warden.Container
		result2 error
	}{result1, result2}
}

func (fake *FakeBackend) Destroy(handle string) error {
	fake.destroyMutex.Lock()
	defer fake.destroyMutex.Unlock()
	fake.destroyArgsForCall = append(fake.destroyArgsForCall, struct {
		handle string
	}{handle})
	if fake.DestroyStub != nil {
		return fake.DestroyStub(handle)
	} else {
		return fake.destroyReturns.result1
	}
}

func (fake *FakeBackend) DestroyCallCount() int {
	fake.destroyMutex.RLock()
	defer fake.destroyMutex.RUnlock()
	return len(fake.destroyArgsForCall)
}

func (fake *FakeBackend) DestroyArgsForCall(i int) string {
	fake.destroyMutex.RLock()
	defer fake.destroyMutex.RUnlock()
	return fake.destroyArgsForCall[i].handle
}

func (fake *FakeBackend) DestroyReturns(result1 error) {
	fake.destroyReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeBackend) Containers(arg1 warden.Properties) ([]warden.Container, error) {
	fake.containersMutex.Lock()
	defer fake.containersMutex.Unlock()
	fake.containersArgsForCall = append(fake.containersArgsForCall, struct {
		arg1 warden.Properties
	}{arg1})
	if fake.ContainersStub != nil {
		return fake.ContainersStub(arg1)
	} else {
		return fake.containersReturns.result1, fake.containersReturns.result2
	}
}

func (fake *FakeBackend) ContainersCallCount() int {
	fake.containersMutex.RLock()
	defer fake.containersMutex.RUnlock()
	return len(fake.containersArgsForCall)
}

func (fake *FakeBackend) ContainersArgsForCall(i int) warden.Properties {
	fake.containersMutex.RLock()
	defer fake.containersMutex.RUnlock()
	return fake.containersArgsForCall[i].arg1
}

func (fake *FakeBackend) ContainersReturns(result1 []warden.Container, result2 error) {
	fake.containersReturns = struct {
		result1 []warden.Container
		result2 error
	}{result1, result2}
}

func (fake *FakeBackend) Lookup(handle string) (warden.Container, error) {
	fake.lookupMutex.Lock()
	defer fake.lookupMutex.Unlock()
	fake.lookupArgsForCall = append(fake.lookupArgsForCall, struct {
		handle string
	}{handle})
	if fake.LookupStub != nil {
		return fake.LookupStub(handle)
	} else {
		return fake.lookupReturns.result1, fake.lookupReturns.result2
	}
}

func (fake *FakeBackend) LookupCallCount() int {
	fake.lookupMutex.RLock()
	defer fake.lookupMutex.RUnlock()
	return len(fake.lookupArgsForCall)
}

func (fake *FakeBackend) LookupArgsForCall(i int) string {
	fake.lookupMutex.RLock()
	defer fake.lookupMutex.RUnlock()
	return fake.lookupArgsForCall[i].handle
}

func (fake *FakeBackend) LookupReturns(result1 warden.Container, result2 error) {
	fake.lookupReturns = struct {
		result1 warden.Container
		result2 error
	}{result1, result2}
}

func (fake *FakeBackend) Start() error {
	fake.startMutex.Lock()
	defer fake.startMutex.Unlock()
	fake.startArgsForCall = append(fake.startArgsForCall, struct{}{})
	if fake.StartStub != nil {
		return fake.StartStub()
	} else {
		return fake.startReturns.result1
	}
}

func (fake *FakeBackend) StartCallCount() int {
	fake.startMutex.RLock()
	defer fake.startMutex.RUnlock()
	return len(fake.startArgsForCall)
}

func (fake *FakeBackend) StartReturns(result1 error) {
	fake.startReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeBackend) Stop() {
	fake.stopMutex.Lock()
	defer fake.stopMutex.Unlock()
	fake.stopArgsForCall = append(fake.stopArgsForCall, struct{}{})
	if fake.StopStub != nil {
		fake.StopStub()
	}
}

func (fake *FakeBackend) StopCallCount() int {
	fake.stopMutex.RLock()
	defer fake.stopMutex.RUnlock()
	return len(fake.stopArgsForCall)
}

func (fake *FakeBackend) GraceTime(arg1 warden.Container) time.Duration {
	fake.graceTimeMutex.Lock()
	defer fake.graceTimeMutex.Unlock()
	fake.graceTimeArgsForCall = append(fake.graceTimeArgsForCall, struct {
		arg1 warden.Container
	}{arg1})
	if fake.GraceTimeStub != nil {
		return fake.GraceTimeStub(arg1)
	} else {
		return fake.graceTimeReturns.result1
	}
}

func (fake *FakeBackend) GraceTimeCallCount() int {
	fake.graceTimeMutex.RLock()
	defer fake.graceTimeMutex.RUnlock()
	return len(fake.graceTimeArgsForCall)
}

func (fake *FakeBackend) GraceTimeArgsForCall(i int) warden.Container {
	fake.graceTimeMutex.RLock()
	defer fake.graceTimeMutex.RUnlock()
	return fake.graceTimeArgsForCall[i].arg1
}

func (fake *FakeBackend) GraceTimeReturns(result1 time.Duration) {
	fake.graceTimeReturns = struct {
		result1 time.Duration
	}{result1}
}

var _ warden.Backend = new(FakeBackend)