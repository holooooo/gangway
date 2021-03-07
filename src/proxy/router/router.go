package router

import (
	"github.com/rs/zerolog/log"
	"github.com/songgao/packets/ethernet"
)

func ListenTap() error {
	ifce, err := createTap()
	if err != nil {
		return err
	}

	var frame ethernet.Frame
	for {
		frame.Resize(1500)
		n, err := ifce.Read([]byte(frame))
		if err != nil {
			log.Warn().Err(err).Msg("error on read tap device")
		}
		frame = frame[:n]
		log.Info().Msgf("Dst: %s", frame.Destination())
		log.Info().Msgf("Src: %s", frame.Source())
		log.Info().Msgf("Ethertype: % x", frame.Ethertype())
		log.Info().Msgf("Payload: % x", string(frame.Payload()))
	}
}
