package model

type Service struct {
	Id             int64      `json:"id",gorm:"AUTO_INCREMENT;primary_key"`
	AppName        string     `json:"appName",gorm:"column:app_name"`
	Level          string     `json:"level"`
	Encryption     int32      `json:"encryption",gorm:"column:is_encrypt"`
	NeedHeaders    int32      `json:"needHeaders",gorm:"column:is_header"`
	Verification   int32      `json:"verification",gorm:"column:is_verify"`
	Type           string     `json:"type",gorm:"column:service_type"`
	ReadTimeOut    int32      `json:"readTimeOut",gorm:"column:read_timeout"`
	ConnectTimeOut int32      `json:"connectTimeOut",gorm:"column:connect_timeout"`
	RequestTimeOut int32      `json:"requestTimeOut",gorm:"column:request_timeout"`
	Url            string     `json:"url",gorm:"column:url"`
	Remark         string     `json:"remark",gorm:"column:remark"`
	Token          string     `json:"token",gorm:"-"`
	ServiceUri     string     `json:"serviceUri",gorm:"-"`
}

func (Service) TableName() string  {
	return "app_service"
}