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
		switch frame.Ethertype() {
		case ethernet.ARP:
			handleARP(ifce, frame.Payload())
		case ethernet.IPv4:
			log.Info().Msg("!!!!!!!!!!!")
			log.Info().Msgf("Payload: % x", string(frame.Payload()))
		default:
			log.Debug().Msg("unsupported ethernet type")
		}

		log.Info().Msgf("Ethertype: % x", frame.Ethertype())
		log.Info().Msgf("Payload: % x", string(frame.Payload()))
	}
}
