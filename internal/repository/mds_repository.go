package repository

import "hood/internal/domain"

type MdsRepository interface {
	GetAssetSplits(symbols []string) ([]domain.AssetSplit, error)
}
