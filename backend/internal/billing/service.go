package billing

import (
	"context"
	"errors"
	"strings"
)

type Service struct{ store *Store }

func NewService(store *Store) *Service { return &Service{store: store} }

func (s *Service) CreatePackage(ctx context.Context, input CreatePackageInput) (Package, error) {
	if err := validatePackage(input.Type, input.DurationSeconds, input.TotalCount, input.AICount, input.PriceCents, input.LimitCount, input.IsTrial, input.IsFree); err != nil {
		return Package{}, err
	}
	return s.store.CreatePackage(ctx, input)
}

func validatePackage(packageType string, durationSeconds *int64, totalCount, aiCount, priceCents, limitCount, isTrial, isFree int) error {
	if packageType != PackageTime && packageType != PackageCount && packageType != PackageTimeCount {
		return errors.New("invalid package type")
	}
	if packageType != PackageCount && (durationSeconds == nil || *durationSeconds <= 0) {
		return errors.New("durationSeconds is required for time packages")
	}
	if packageType != PackageTime && totalCount <= 0 {
		return errors.New("totalCount is required for count packages")
	}
	if aiCount < 0 || priceCents < 0 || limitCount < 0 {
		return errors.New("package numeric values cannot be negative")
	}
	if (isTrial != 0 && isTrial != 1) || (isFree != 0 && isFree != 1) {
		return errors.New("isTrial and isFree must be 0 or 1")
	}
	return nil
}

func (s *Service) CreateCoupon(ctx context.Context, input CreateCouponInput) (Coupon, error) {
	input.Code = strings.ToUpper(strings.TrimSpace(input.Code))
	if err := validateCoupon(input); err != nil {
		return Coupon{}, err
	}
	return s.store.CreateCoupon(ctx, input)
}

func validateCoupon(input CreateCouponInput) error {
	if input.Code == "" || (input.DiscountType != "fixed" && input.DiscountType != "percent") || input.DiscountValue <= 0 || input.TotalLimit < 0 {
		return errors.New("invalid coupon")
	}
	if input.DiscountType == "percent" && input.DiscountValue > 100 {
		return errors.New("percent coupon cannot exceed 100")
	}
	return nil
}

func validateStatus(status int) error {
	if status != 0 && status != 1 {
		return errors.New("invalid status")
	}
	return nil
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

func (s *Service) ListPackages(ctx context.Context) ([]Package, error) {
	return s.store.ListPackages(ctx)
}

func (s *Service) UpdatePackageStatus(ctx context.Context, id uint64, status int) error {
	if err := validateStatus(status); err != nil {
		return errors.New("invalid package status")
	}
	return s.store.UpdatePackageStatus(ctx, id, status)
}

func (s *Service) UpdatePackage(ctx context.Context, id uint64, input UpdatePackageInput) (Package, error) {
	if err := validatePackage(input.Type, input.DurationSeconds, input.TotalCount, input.AICount, input.PriceCents, input.LimitCount, input.IsTrial, input.IsFree); err != nil {
		return Package{}, err
	}
	return s.store.UpdatePackage(ctx, id, input)
}

func (s *Service) UpdateCouponStatus(ctx context.Context, id uint64, status int) error {
	if err := validateStatus(status); err != nil {
		return errors.New("invalid coupon status")
	}
	return s.store.UpdateCouponStatus(ctx, id, status)
}

func (s *Service) Consume(ctx context.Context, userID, packageID uint64, kind, requestID, endpoint string, amount int) (PackageInstance, error) {
	return s.store.Consume(ctx, userID, packageID, kind, requestID, endpoint, amount)
}

func (s *Service) HasAIQuota(ctx context.Context, userID, packageID uint64) (bool, error) {
	return s.store.HasAIQuota(ctx, userID, packageID)
}
