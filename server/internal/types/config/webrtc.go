package config

import (
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"n.eko.moe/neko/internal/utils"
)

type WebRTC struct {
	ICELite      bool
	ICEServers   []string
	EphemeralMin uint16
	EphemeralMax uint16
	NAT1To1IPs   []string
}

func (WebRTC) Init(cmd *cobra.Command) error {
	cmd.PersistentFlags().String("epr", "59000-59100", "limits the pool of ephemeral ports that ICE UDP connections can allocate from")
	if err := viper.BindPFlag("epr", cmd.PersistentFlags().Lookup("epr")); err != nil {
		return err
	}

	cmd.PersistentFlags().StringSlice("nat1to1", []string{}, "sets a list of external IP addresses of 1:1 (D)NAT and a candidate type for which the external IP address is used")
	if err := viper.BindPFlag("nat1to1", cmd.PersistentFlags().Lookup("nat1to1")); err != nil {
		return err
	}

	cmd.PersistentFlags().Bool("icelite", false, "configures whether or not the ice agent should be a lite agent")
	if err := viper.BindPFlag("icelite", cmd.PersistentFlags().Lookup("icelite")); err != nil {
		return err
	}

	cmd.PersistentFlags().StringSlice("iceserver", []string{"stun:stun.l.google.com:19302"}, "describes a single STUN and TURN server that can be used by the ICEAgent to establish a connection with a peer")
	if err := viper.BindPFlag("iceserver", cmd.PersistentFlags().Lookup("iceserver")); err != nil {
		return err
	}

	return nil
}

func (s *WebRTC) Set() {
	s.ICELite = viper.GetBool("icelite")
	s.ICEServers = viper.GetStringSlice("iceserver")
	s.NAT1To1IPs = viper.GetStringSlice("nat1to1")

	if len(s.NAT1To1IPs) == 0 {
		ip, err := utils.GetIP()
		if err == nil {
			s.NAT1To1IPs = append(s.NAT1To1IPs, ip)
		}
	}

	min := uint16(59000)
	max := uint16(59100)
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
