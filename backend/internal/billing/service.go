package billing

import (
	"context"
	"errors"
)

type Service struct{ store *Store }

func NewService(store *Store) *Service { return &Service{store: store} }

func (s *Service) CreatePackage(ctx context.Context, input CreatePackageInput) (Package, error) {
	if input.Type != PackageTime && input.Type != PackageCount && input.Type != PackageTimeCount {
		return Package{}, errors.New("invalid package type")
	}
	if input.Type != PackageCount && (input.DurationSeconds == nil || *input.DurationSeconds <= 0) {
		return Package{}, errors.New("durationSeconds is required for time packages")
	}
	if input.Type != PackageTime && input.TotalCount <= 0 {
		return Package{}, errors.New("totalCount is required for count packages")
	}
	return s.store.CreatePackage(ctx, input)
}

func (s *Service) GrantPackage(ctx context.Context, userID, packageID uint64) (PackageInstance, error) {
	return s.store.GrantPackage(ctx, userID, packageID)
}

func (s *Service) ListInstances(ctx context.Context, userID uint64) ([]PackageInstance, error) {
	return s.store.ListInstances(ctx, userID)
}

func (s *Service) Consume(ctx context.Context, userID, packageID uint64, kind, requestID, endpoint string) (PackageInstance, error) {
	return s.store.Consume(ctx, userID, packageID, kind, requestID, endpoint)
}
