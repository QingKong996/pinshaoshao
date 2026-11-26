package service

import (
	"context"
	"errors"
	"testing"
)

type expiryServiceStub struct {
	due       []int
	dueErr    error
	expireErr map[int]error
	expired   []int
	limit     int
}

func (s *expiryServiceStub) DueOrderIDs(_ context.Context, limit int) ([]int, error) {
	s.limit = limit
	if limit < len(s.due) {
		return s.due[:limit], s.dueErr
	}
	return s.due, s.dueErr
}

func (s *expiryServiceStub) Expire(_ context.Context, orderID int) error {
	s.expired = append(s.expired, orderID)
	return s.expireErr[orderID]
}

func TestExpiryWorkerProcessesBatchAndContinuesAfterFailure(t *testing.T) {
	service := &expiryServiceStub{
		due:       []int{3, 4, 5},
		expireErr: map[int]error{4: errors.New("temporary failure")},
	}
	worker := NewExpiryWorker(service, 0, 2)

	worker.process(context.Background())

	if service.limit != 2 {
		t.Fatalf("due limit = %d, want 2", service.limit)
	}
	if len(service.expired) != 2 ||
		service.expired[0] != 3 ||
		service.expired[1] != 4 {
		t.Fatalf("expired orders = %v, want [3 4]", service.expired)
	}
}

func TestExpiryWorkerStopsWhenDueQueryFails(t *testing.T) {
	service := &expiryServiceStub{
		dueErr: errors.New("database unavailable"),
	}
	worker := NewExpiryWorker(service, 0, 0)

	worker.process(context.Background())

	if len(service.expired) != 0 {
		t.Fatalf("expired orders = %v, want none", service.expired)
	}
}
