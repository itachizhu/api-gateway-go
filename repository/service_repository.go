package repository

import "github.com/itachizhu/api-gateway-go/model"

type ServiceRepository struct {
}

func NewServiceRepository() *ServiceRepository {
	return new(ServiceRepository)
}

func (s *ServiceRepository) FindService(appName string) *model.Service {
	service := new(model.Service)
	db.Where("app_name=?", appName).First(service)
	if db.Error != nil {
		panic(db.Error)
	}
	return service
}