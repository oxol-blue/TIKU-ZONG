package billing

import "time"

const (
	PackageTime      = "time"
	PackageCount     = "count"
	PackageTimeCount = "time_count"
	UsageQuestions   = "questions"
	UsageAI          = "ai"
)

type Package struct {
	ID              uint64    `json:"id"`
	Name            string    `json:"name"`
	Type            string    `json:"type"`
	DurationSeconds *int64    `json:"durationSeconds,omitempty"`
	TotalCount      int       `json:"totalCount"`
	AICount         int       `json:"aiCount"`
	PriceCents      int       `json:"priceCents"`
	Status          int       `json:"status"`
	LimitCount      int       `json:"limitCount"`
	IsTrial         int       `json:"isTrial"`
	IsFree          int       `json:"isFree"`
	CreatedAt       time.Time `json:"createdAt"`
}

type CreatePackageInput struct {
	Name            string `json:"name" binding:"required"`
	Type            string `json:"type" binding:"required"`
	DurationSeconds *int64 `json:"durationSeconds"`
	TotalCount      int    `json:"totalCount"`
	AICount         int    `json:"aiCount"`
	PriceCents      int    `json:"priceCents"`
	LimitCount      int    `json:"limitCount"`
	IsTrial         int    `json:"isTrial"`
	IsFree          int    `json:"isFree"`
}

type Coupon struct {
	ID            uint64     `json:"id"`
	Code          string     `json:"code"`
	DiscountType  string     `json:"discountType"`
	DiscountValue int        `json:"discountValue"`
	TotalLimit    int        `json:"totalLimit"`
	UsedCount     int        `json:"usedCount"`
	ReservedCount int        `json:"reservedCount"`
	ExpiresAt     *time.Time `json:"expiresAt,omitempty"`
	Status        int        `json:"status"`
}

type CreateCouponInput struct {
	Code          string     `json:"code" binding:"required"`
	DiscountType  string     `json:"discountType" binding:"required"`
	DiscountValue int        `json:"discountValue" binding:"required"`
	TotalLimit    int        `json:"totalLimit"`
	ExpiresAt     *time.Time `json:"expiresAt"`
}

type PackageInstance struct {
	ID               uint64     `json:"id"`
	PackageID        uint64     `json:"packageId"`
	PackageName      string     `json:"packageName"`
	PackageType      string     `json:"packageType"`
	StartsAt         time.Time  `json:"startsAt"`
	ExpiresAt        *time.Time `json:"expiresAt,omitempty"`
	RemainingCount   int        `json:"remainingCount"`
	RemainingAICount int        `json:"remainingAiCount"`
	Status           int        `json:"status"`
}
