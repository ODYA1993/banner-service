package banner

import (
	"context"
)

type Storage interface {
	GetBannerFromDB(ctx context.Context, tagID, featureID int, useLastRevision bool) (*Banner, error)
	GetBannersByFiltering(ctx context.Context, featureID, tagID, limit, offset int) ([]*Banner, error)
	CreateBanner(ctx context.Context, banner *Banner) error
	UpdateBanner(ctx context.Context, banner *Banner) error
	DeleteBanner(ctx context.Context, id int) error
}
