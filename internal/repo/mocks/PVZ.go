// Code generated by mockery v2.53.3. DO NOT EDIT.

package mocks

import (
	context "context"

	entity "github.com/GlebMoskalev/go-pickup-point-api/internal/entity"
	mock "github.com/stretchr/testify/mock"

	time "time"
)

// PVZ is an autogenerated mock type for the PVZ type
type PVZ struct {
	mock.Mock
}

// Create provides a mock function with given fields: ctx, city
func (_m *PVZ) Create(ctx context.Context, city string) (*entity.PVZ, error) {
	ret := _m.Called(ctx, city)

	if len(ret) == 0 {
		panic("no return value specified for Create")
	}

	var r0 *entity.PVZ
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (*entity.PVZ, error)); ok {
		return rf(ctx, city)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) *entity.PVZ); ok {
		r0 = rf(ctx, city)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*entity.PVZ)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, city)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Exists provides a mock function with given fields: ctx, pvzID
func (_m *PVZ) Exists(ctx context.Context, pvzID string) bool {
	ret := _m.Called(ctx, pvzID)

	if len(ret) == 0 {
		panic("no return value specified for Exists")
	}

	var r0 bool
	if rf, ok := ret.Get(0).(func(context.Context, string) bool); ok {
		r0 = rf(ctx, pvzID)
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// ListWithDetails provides a mock function with given fields: ctx, startDate, endDate, page, limit
func (_m *PVZ) ListWithDetails(ctx context.Context, startDate *time.Time, endDate *time.Time, page int, limit int) ([]entity.PVZWithDetails, error) {
	ret := _m.Called(ctx, startDate, endDate, page, limit)

	if len(ret) == 0 {
		panic("no return value specified for ListWithDetails")
	}

	var r0 []entity.PVZWithDetails
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *time.Time, *time.Time, int, int) ([]entity.PVZWithDetails, error)); ok {
		return rf(ctx, startDate, endDate, page, limit)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *time.Time, *time.Time, int, int) []entity.PVZWithDetails); ok {
		r0 = rf(ctx, startDate, endDate, page, limit)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]entity.PVZWithDetails)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *time.Time, *time.Time, int, int) error); ok {
		r1 = rf(ctx, startDate, endDate, page, limit)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewPVZ creates a new instance of PVZ. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewPVZ(t interface {
	mock.TestingT
	Cleanup(func())
}) *PVZ {
	mock := &PVZ{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
