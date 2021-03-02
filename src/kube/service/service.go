package service

import "github.com/rs/zerolog/log"

type Service struct {
	localPort   int
	serviceName string
	namespace   string
	servicePort int
}

//todo
func loadServiceConfig() ([]Service, error) {
	return nil, nil
}

func Init() {
	svcs, err := loadServiceConfig()
	if err != nil {
		//todo do something
	}
	for _, svc := range svcs {
		err := createService(svc)
		if err != nil {
			log.Err(err)
		}
	}
}

// createService is an async operat, it may failed by duplicate name, error format
// and port occupied
func createService(svc Service) error {
	return nil
}
