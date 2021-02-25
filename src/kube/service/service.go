package service

import "github.com/rs/zerolog/log"

type Service struct {
	localPort   int
	serviceName string
	namespace   string
	servicePort int
}

var svcs []Service

//todo
func loadServiceConfig() ([]Service, error) {
	return nil, nil
}

func Collect() (int, error) {
	svcs, err := loadServiceConfig()
	if err != nil {
		return 0, err
	}
	return len(svcs), nil
}

//todo
func Register() {
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
