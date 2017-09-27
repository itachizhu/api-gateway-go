package model

type Cache struct {
	Id             int64      `json:"id",gorm:"AUTO_INCREMENT;primary_key"`
	ServiceId      int64      `json:"serviceId",gorm:"column:app_id"`
	AppName        string     `json:"appName",gorm:"-"`
	ByUser         int32      `json:"byUser",gorm:"column:by_user"`
	CacheTime      int32      `json:"cacheTime",gorm:"column:cache_time"`
	Uri            string     `json:"uri",gorm:"column:url"`
	Path           string     `json:"path",gorm:"column:path"`
}

func (Cache) TableName() string  {
	return "app_cache"
}