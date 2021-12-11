package config

import (
	"encoding/json"
	"strconv"
	"strings"

	"m1k1o/neko/internal/utils"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/pion/webrtc/v3"
)

type WebRTC struct {
	ICELite      bool
	ICEServers   []webrtc.ICEServer
	EphemeralMin uint16
	EphemeralMax uint16
	NAT1To1IPs   []string
	TCPMUX       int
	UDPMUX       int

	ImplicitControl bool
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

	cmd.PersistentFlags().Int("tcpmux", 0, "single TCP mux port for all peers")
	if err := viper.BindPFlag("tcpmux", cmd.PersistentFlags().Lookup("tcpmux")); err != nil {
		return err
	}

	cmd.PersistentFlags().Int("udpmux", 0, "single UDP mux port for all peers")
	if err := viper.BindPFlag("udpmux", cmd.PersistentFlags().Lookup("udpmux")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("ipfetch", "http://checkip.amazonaws.com", "automatically fetch IP address from given URL when nat1to1 is not present")
	if err := viper.BindPFlag("ipfetch", cmd.PersistentFlags().Lookup("ipfetch")); err != nil {
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

	cmd.PersistentFlags().String("iceservers", "", "describes a single STUN and TURN server that can be used by the ICEAgent to establish a connection with a peer")
	if err := viper.BindPFlag("iceservers", cmd.PersistentFlags().Lookup("iceservers")); err != nil {
		return err
	}

	// TODO: Should be moved to session config.
	cmd.PersistentFlags().Bool("implicit_control", false, "if enabled members can gain control implicitly")
	if err := viper.BindPFlag("implicit_control", cmd.PersistentFlags().Lookup("implicit_control")); err != nil {
		return err
	}

	return nil
}

func (s *WebRTC) Set() {
	s.NAT1To1IPs = viper.GetStringSlice("nat1to1")
	s.TCPMUX = viper.GetInt("tcpmux")
	s.UDPMUX = viper.GetInt("udpmux")
	s.ICELite = viper.GetBool("icelite")
	s.ICEServers = []webrtc.ICEServer{}

	iceServersJson := viper.GetString("iceservers")
	if iceServersJson != "" {
		err := json.Unmarshal([]byte(iceServersJson), &s.ICEServers)
		if err != nil {
			log.Panic().Err(err).Msg("failed to process iceservers")
		}
	}

	iceServerSlice := viper.GetStringSlice("iceserver")
	if len(iceServerSlice) > 0 {
		s.ICEServers = append(s.ICEServers, webrtc.ICEServer{URLs: iceServerSlice})
	}

	if len(s.NAT1To1IPs) == 0 {
		ipfetch := viper.GetString("ipfetch")
		ip, err := utils.GetIP(ipfetch)
		if err != nil {
			log.Panic().Err(err).Str("ipfetch", ipfetch).Msg("failed to fetch ip address")
		}
		s.NAT1To1IPs = append(s.NAT1To1IPs, ip)
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

	// TODO: Should be moved to session config.
	s.ImplicitControl = viper.GetBool("implicit_control")
}
