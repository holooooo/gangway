package service

import (
	"fmt"

	"github.com/rs/zerolog/log"
)

type Service struct {
	Type string // service type.accept listener or proxy

	LocalPort int

	ServiceName      string
	ServiceNamespace string
	ServicePort      int
}

func (s *Service) String() string {
	return fmt.Sprintf("[local:%v=>%v:%v]", s.LocalPort, s.ServiceName, s.ServicePort)
}

func (s *Service) GetLocalAddress() string {
	return fmt.Sprintf("0.0.0.0:%v", s.LocalPort)
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
