package config

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"demodesk/neko/internal/utils"
)

type WebRTC struct {
	ICELite      bool
	ICETrickle   bool
	ICEServers   []string
	EphemeralMin uint16
	EphemeralMax uint16

	NAT1To1IPs     []string
	IpRetrievalUrl string
}

const (
	defEprMin = 59000
	defEprMax = 59100
)

func (WebRTC) Init(cmd *cobra.Command) error {
	cmd.PersistentFlags().Bool("icelite", false, "configures whether or not the ice agent should be a lite agent")
	if err := viper.BindPFlag("icelite", cmd.PersistentFlags().Lookup("icelite")); err != nil {
		return err
	}

	cmd.PersistentFlags().Bool("icetrickle", true, "configures whether cadidates should be sent asynchronously using Trickle ICE")
	if err := viper.BindPFlag("icetrickle", cmd.PersistentFlags().Lookup("icetrickle")); err != nil {
		return err
	}

	cmd.PersistentFlags().StringSlice("iceserver", []string{"stun:stun.l.google.com:19302"}, "describes a single STUN and TURN server that can be used by the ICEAgent to establish a connection with a peer")
	if err := viper.BindPFlag("iceserver", cmd.PersistentFlags().Lookup("iceserver")); err != nil {
		return err
	}

	cmd.PersistentFlags().StringSlice("nat1to1", []string{}, "sets a list of external IP addresses of 1:1 (D)NAT and a candidate type for which the external IP address is used")
	if err := viper.BindPFlag("nat1to1", cmd.PersistentFlags().Lookup("nat1to1")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("ip_retrieval_url", "https://checkip.amazonaws.com", "URL address used for retrieval of the external IP address")
	if err := viper.BindPFlag("ip_retrieval_url", cmd.PersistentFlags().Lookup("ip_retrieval_url")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("epr", fmt.Sprintf("%d-%d", defEprMin, defEprMax), "limits the pool of ephemeral ports that ICE UDP connections can allocate from")
	if err := viper.BindPFlag("epr", cmd.PersistentFlags().Lookup("epr")); err != nil {
		return err
	}

	return nil
}

func (s *WebRTC) Set() {
	s.ICELite = viper.GetBool("icelite")
	s.ICETrickle = viper.GetBool("icetrickle")
	s.ICEServers = viper.GetStringSlice("iceserver")

	s.NAT1To1IPs = viper.GetStringSlice("nat1to1")
	s.IpRetrievalUrl = viper.GetString("ip_retrieval_url")
	if s.IpRetrievalUrl != "" && len(s.NAT1To1IPs) == 0 {
		ip, err := utils.HttpRequestGET(s.IpRetrievalUrl)
		if err == nil {
			s.NAT1To1IPs = append(s.NAT1To1IPs, ip)
		} else {
			log.Warn().Err(err).Msgf("IP retrieval failed")
		}
	}

	min := uint16(defEprMin)
	max := uint16(defEprMax)

	epr := viper.GetString("epr")
	ports := strings.SplitN(epr, "-", -1)
	if len(ports) > 1 {
		start, err := strconv.ParseUint(ports[0], 10, 16)
		if err == nil {
			min = uint16(start)
		}

		end, err := strconv.ParseUint(ports[1], 10, 16)
		if err == nil {
			max = uint16(end)
		}
	}

	if min > max {
		s.EphemeralMin = max
		s.EphemeralMax = min
	} else {
		s.EphemeralMin = min
		s.EphemeralMax = max
	}
}
