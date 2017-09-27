package repository

import "github.com/itachizhu/api-gateway-go/model"

type CacheRepository struct {
}

func NewCacheRepository() *CacheRepository {
	return new(CacheRepository)
}

func (p *CacheRepository) FindCache(appId int64, paths []string) *model.Cache {
	cache := new(model.Cache)
	db.Where("app_id = ? and path in (?)", appId, paths).Order("path desc", true).First(cache)
	if db.Error != nil {
		panic(db.Error)
	}
	return cache
}