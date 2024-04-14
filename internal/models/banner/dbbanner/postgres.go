package dbbanner

import (
	"banner-service/internal/models/banner"
	"banner-service/pkg/db/postgresql"
	"banner-service/pkg/logging"
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
)

type bannerRepository struct {
	db     postgresql.Client
	logger *logging.Logger
}

func NewBannerRepository(db postgresql.Client, logger *logging.Logger) banner.Storage {
	return &bannerRepository{
		db:     db,
		logger: logger,
	}
}

func (b *bannerRepository) GetBannerFromDB(ctx context.Context, tagID, featureID int, useLastRevision bool) (*banner.Banner, error) {
	query := `
    SELECT b.id, b.title, b.text, b.url, b.is_active, b.created_at, b.updated_at, array_agg(t.id) as tag_ids, f.id as feature_id, f.name as feature_name
    FROM banners b
    JOIN features f ON b.feature_id = f.id
    JOIN banner_tags bt ON b.id = bt.banner_id
    JOIN tags t ON bt.tag_id = t.id
    WHERE f.id = $1 AND t.id = $2
    GROUP BY b.id, f.id
    ORDER BY
`

	var orderBy string
	if useLastRevision {
		orderBy = "b.updated_at DESC"
	} else {
		orderBy = "b.created_at DESC"
	}

	query += orderBy + " LIMIT 1"

	b.logger.Trace(fmt.Sprintf("SQL Query: %s", query))

	row := b.db.QueryRow(ctx, query, featureID, tagID)

	var bn banner.Banner
	var tagIDs []int
	var featureName string
	err := row.Scan(&bn.ID, &bn.Title, &bn.Text, &bn.URL, &bn.IsActive, &bn.CreatedAt, &bn.UpdatedAt, &tagIDs, &bn.FeatureID.ID, &featureName)
	if err != nil {
		return nil, err
	}

	for _, tagID := range tagIDs {
		var tag banner.Tag
		err := b.db.QueryRow(ctx, "SELECT id, name FROM tags WHERE id = $1", tagID).Scan(&tag.ID, &tag.Name)
		if err != nil {
			return nil, err
		}
		bn.Tags = append(bn.Tags, tag)
	}

	bn.FeatureID.Name = featureName

	return &bn, nil
}

func (b *bannerRepository) GetBannersByFiltering(ctx context.Context, featureID, tagID, limit, offset int) ([]*banner.Banner, error) {
	var banners []*banner.Banner

	query := `
       SELECT banners.id,
              banners.feature_id,
              features.name AS feature_name,
              banners.title,
              banners.text,
              banners.url,
              banners.is_active,
              banners.created_at,
              banners.updated_at,
              array_agg(tags.id) AS tag_ids,
              array_agg(tags.name) AS tag_names
       FROM banners
       LEFT JOIN banner_tags ON banners.id = banner_tags.banner_id
       LEFT JOIN tags ON banner_tags.tag_id = tags.id
       LEFT JOIN features ON banners.feature_id = features.id
       WHERE
           (banners.feature_id = $1 OR $1 IS NULL)
           AND (tags.id = $2 OR $2 IS NULL)
       GROUP BY
           banners.id, banners.title, banners.text, banners.url, banners.feature_id, banners.is_active,
           banners.created_at, banners.updated_at, features.name
       LIMIT $3 OFFSET $4;
   `

	b.logger.Trace(fmt.Sprintf("SQL Query: %s", query))

	rows, err := b.db.Query(ctx, query, featureID, tagID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var ban banner.Banner
		var tagIDs []int
		var tagNames []string

		err = rows.Scan(&ban.ID, &ban.FeatureID.ID, &ban.FeatureID.Name, &ban.Title, &ban.Text, &ban.URL, &ban.IsActive, &ban.CreatedAt, &ban.UpdatedAt, &tagIDs, &tagNames)
		if err != nil {
			return nil, err
		}

		var tags []banner.Tag
		for i, tagID := range tagIDs {
			tags = append(tags, banner.Tag{ID: tagID, Name: tagNames[i]})
		}
		ban.Tags = tags

		banners = append(banners, &ban)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return banners, nil
}

func (b *bannerRepository) CreateBanner(ctx context.Context, banner *banner.Banner) error {
	query := `INSERT INTO banners (title, text, url, is_active, feature_id) VALUES ($1, $2, $3, $4, $5) RETURNING id`

	b.logger.Trace(fmt.Sprintf("SQL Query: %s", query))

	featureID := banner.FeatureID.ID
	err := b.db.QueryRow(ctx, query, banner.Title, banner.Text, banner.URL, banner.IsActive, featureID).Scan(&banner.ID)
	if err != nil {
		return err
	}

	// Сохраняем связи баннера с тегами
	for _, tag := range banner.Tags {
		q := `INSERT INTO banner_tags (banner_id, tag_id) VALUES ($1, $2)`
		_, err := b.db.Exec(ctx, q, banner.ID, tag.ID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (b *bannerRepository) UpdateBanner(ctx context.Context, banner *banner.Banner) error {
	query := `UPDATE banners SET title=$1, text=$2, url=$3, is_active=$4, feature_id=$5 WHERE id=$6`
	tx, err := b.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	b.logger.Trace(fmt.Sprintf("SQL Query: %s", query))

	featureID := banner.FeatureID.ID
	_, err = b.db.Exec(ctx, query, banner.Title, banner.Text, banner.URL, banner.IsActive, featureID, banner.ID)
	if err != nil {
		return err
	}

	// Удаляем все связи баннера с тегами
	q := `DELETE FROM banner_tags WHERE banner_id=$1`
	_, err = b.db.Exec(ctx, q, banner.ID)
	if err != nil {
		return err
	}

	// Сохраняем новые связи баннера с тегами
	for _, tag := range banner.Tags {
		q := `INSERT INTO banner_tags (banner_id, tag_id) VALUES ($1, $2)`
		_, err := b.db.Exec(ctx, q, banner.ID, tag.ID)
		if err != nil {
			return err
		}
	}
	err = tx.Commit(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (b *bannerRepository) DeleteBanner(ctx context.Context, id int) error {
	// Удаляем все связи баннера с тегами
	q := `DELETE FROM banner_tags WHERE banner_id=$1`
	tx, err := b.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	b.logger.Trace(fmt.Sprintf("SQL Query: %s", q))

	_, err = b.db.Exec(ctx, q, id)
	if err != nil {
		return err
	}

	// Удаляем сам баннер
	query := `DELETE FROM banners WHERE id=$1`
	_, err = b.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	return nil
}
