package router

import (
	"gangway/src/settings"
	"os/exec"

	"github.com/rs/zerolog/log"
	"github.com/songgao/water"
)

func createTap() (*water.Interface, error) {
	config := water.Config{
		DeviceType: water.TAP,
	}
	config.Name = "gangway"
	ifce, err := water.New(config)
	if err != nil {
		return nil, err
	}
	log.Debug().Msgf("tap device created")

	cmd := exec.Command("ip", "addr", "add", *settings.CIDR, "dev", ifce.Name())
	err = cmd.Run()
	if err != nil {
		return nil, err
	}

	cmd = exec.Command("ip", "link", "set", "dev", ifce.Name(), "up")
	err = cmd.Run()
	if err != nil {
		return nil, err
	}
	return ifce, nil
}
