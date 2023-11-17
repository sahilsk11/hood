package repository

type MdsRepository interface {
	GetAssetSplits(symbols []string)
}
