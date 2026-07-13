package billing

import (
	"context"
	"errors"
	"strings"
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
	if (input.IsTrial != 0 && input.IsTrial != 1) || (input.IsFree != 0 && input.IsFree != 1) {
		return Package{}, errors.New("isTrial and isFree must be 0 or 1")
	}
	return s.store.CreatePackage(ctx, input)
}

func (s *Service) CreateCoupon(ctx context.Context, input CreateCouponInput) (Coupon, error) {
	input.Code = strings.ToUpper(strings.TrimSpace(input.Code))
	if input.Code == "" || (input.DiscountType != "fixed" && input.DiscountType != "percent") || input.DiscountValue <= 0 || input.TotalLimit < 0 {
		return Coupon{}, errors.New("invalid coupon")
	}
	if input.DiscountType == "percent" && input.DiscountValue > 100 {
		return Coupon{}, errors.New("percent coupon cannot exceed 100")
	}
	return s.store.CreateCoupon(ctx, input)
}

func (s *Service) ListCoupons(ctx context.Context) ([]Coupon, error) {
	return s.store.ListCoupons(ctx)
}

func (s *Service) GrantPackage(ctx context.Context, userID, packageID uint64) (PackageInstance, error) {
	return s.store.GrantPackage(ctx, userID, packageID)
}

func (s *Service) ListInstances(ctx context.Context, userID uint64) ([]PackageInstance, error) {
	return s.store.ListInstances(ctx, userID)
}

func (s *Service) ListAvailablePackages(ctx context.Context) ([]Package, error) {
	return s.store.ListAvailablePackages(ctx)
}

func (s *Service) Consume(ctx context.Context, userID, packageID uint64, kind, requestID, endpoint string, amount int) (PackageInstance, error) {
	return s.store.Consume(ctx, userID, packageID, kind, requestID, endpoint, amount)
}
