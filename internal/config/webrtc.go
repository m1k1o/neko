package config

import (
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/demodesk/neko/pkg/types"
	"github.com/demodesk/neko/pkg/utils"
)

// default stun server
const defStunSrv = "stun:stun.l.google.com:19302"

type WebRTC struct {
	ICELite      bool
	ICETrickle   bool
	ICEServers   []types.ICEServer
	EphemeralMin uint16
	EphemeralMax uint16
	TCPMux       int
	UDPMux       int

	NAT1To1IPs     []string
	IpRetrievalUrl string

	EstimatorEnabled        bool
	EstimatorInitialBitrate int
}

func (WebRTC) Init(cmd *cobra.Command) error {
	cmd.PersistentFlags().Bool("webrtc.icelite", false, "configures whether or not the ICE agent should be a lite agent")
	if err := viper.BindPFlag("webrtc.icelite", cmd.PersistentFlags().Lookup("webrtc.icelite")); err != nil {
		return err
	}

	cmd.PersistentFlags().Bool("webrtc.icetrickle", true, "configures whether cadidates should be sent asynchronously using Trickle ICE")
	if err := viper.BindPFlag("webrtc.icetrickle", cmd.PersistentFlags().Lookup("webrtc.icetrickle")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("webrtc.iceservers", "[]", "STUN and TURN servers in JSON format with `urls`, `username`, `password` keys")
	if err := viper.BindPFlag("webrtc.iceservers", cmd.PersistentFlags().Lookup("webrtc.iceservers")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("webrtc.epr", "", "limits the pool of ephemeral ports that ICE UDP connections can allocate from")
	if err := viper.BindPFlag("webrtc.epr", cmd.PersistentFlags().Lookup("webrtc.epr")); err != nil {
		return err
	}

	cmd.PersistentFlags().Int("webrtc.tcpmux", 0, "single TCP mux port for all peers")
	if err := viper.BindPFlag("webrtc.tcpmux", cmd.PersistentFlags().Lookup("webrtc.tcpmux")); err != nil {
		return err
	}

	cmd.PersistentFlags().Int("webrtc.udpmux", 0, "single UDP mux port for all peers, replaces EPR")
	if err := viper.BindPFlag("webrtc.udpmux", cmd.PersistentFlags().Lookup("webrtc.udpmux")); err != nil {
		return err
	}

	cmd.PersistentFlags().StringSlice("webrtc.nat1to1", []string{}, "sets a list of external IP addresses of 1:1 (D)NAT and a candidate type for which the external IP address is used")
	if err := viper.BindPFlag("webrtc.nat1to1", cmd.PersistentFlags().Lookup("webrtc.nat1to1")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("webrtc.ip_retrieval_url", "https://checkip.amazonaws.com", "URL address used for retrieval of the external IP address")
	if err := viper.BindPFlag("webrtc.ip_retrieval_url", cmd.PersistentFlags().Lookup("webrtc.ip_retrieval_url")); err != nil {
		return err
	}

	// bandwidth estimator

	cmd.PersistentFlags().Bool("webrtc.estimator.enabled", false, "enables the bandwidth estimator")
	if err := viper.BindPFlag("webrtc.estimator.enabled", cmd.PersistentFlags().Lookup("webrtc.estimator.enabled")); err != nil {
		return err
	}

	cmd.PersistentFlags().Int("webrtc.estimator.initial_bitrate", 1_000_000, "initial bitrate for the bandwidth estimator")
	if err := viper.BindPFlag("webrtc.estimator.initial_bitrate", cmd.PersistentFlags().Lookup("webrtc.estimator.initial_bitrate")); err != nil {
		return err
	}

	return nil
}

func (s *WebRTC) Set() {
	s.ICELite = viper.GetBool("webrtc.icelite")
	s.ICETrickle = viper.GetBool("webrtc.icetrickle")

	if err := viper.UnmarshalKey("webrtc.iceservers", &s.ICEServers, viper.DecodeHook(
		utils.JsonStringAutoDecode(s.ICEServers),
	)); err != nil {
		log.Warn().Err(err).Msgf("unable to parse ICE servers")
	}

	if len(s.ICEServers) == 0 {
		s.ICEServers = append(s.ICEServers, types.ICEServer{
			URLs: []string{defStunSrv},
		})
	}

	s.TCPMux = viper.GetInt("webrtc.tcpmux")
	s.UDPMux = viper.GetInt("webrtc.udpmux")

	epr := viper.GetString("webrtc.epr")
	if epr != "" {
		ports := strings.SplitN(epr, "-", -1)
		if len(ports) > 1 {
			min, err := strconv.ParseUint(ports[0], 10, 16)
			if err != nil {
				log.Panic().Err(err).Msgf("unable to parse ephemeral min port")
			}

			max, err := strconv.ParseUint(ports[1], 10, 16)
			if err != nil {
				log.Panic().Err(err).Msgf("unable to parse ephemeral max port")
			}

			s.EphemeralMin = uint16(min)
			s.EphemeralMax = uint16(max)
		}

		if s.EphemeralMin > s.EphemeralMax {
			log.Panic().Msgf("ephemeral min port cannot be bigger than max")
		}
	}

	if epr == "" && s.TCPMux == 0 && s.UDPMux == 0 {
		// using default epr range
		s.EphemeralMin = 59000
		s.EphemeralMax = 59100

		log.Warn().
			Uint16("min", s.EphemeralMin).
			Uint16("max", s.EphemeralMax).
			Msgf("no TCP, UDP mux or epr specified, using default epr range")
	}

	s.NAT1To1IPs = viper.GetStringSlice("webrtc.nat1to1")
	s.IpRetrievalUrl = viper.GetString("webrtc.ip_retrieval_url")
	if s.IpRetrievalUrl != "" && len(s.NAT1To1IPs) == 0 {
		ip, err := utils.HttpRequestGET(s.IpRetrievalUrl)
		if err == nil {
			s.NAT1To1IPs = append(s.NAT1To1IPs, ip)
		} else {
			log.Warn().Err(err).Msgf("IP retrieval failed")
		}
	}

	// bandwidth estimator

	s.EstimatorEnabled = viper.GetBool("webrtc.estimator.enabled")
	s.EstimatorInitialBitrate = viper.GetInt("webrtc.estimator.initial_bitrate")
}
