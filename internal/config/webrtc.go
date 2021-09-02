package config

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"demodesk/neko/internal/types"
	"demodesk/neko/internal/utils"
)

// default port range - min
const defEprMin = 59000

// default port range - max
const defEprMax = 59100

// default stun server
const defStunSrv = "stun:stun.l.google.com:19302"

type WebRTC struct {
	ICELite      bool
	ICETrickle   bool
	ICEServers   []types.ICEServer
	EphemeralMin uint16
	EphemeralMax uint16

	NAT1To1IPs     []string
	IpRetrievalUrl string
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

	cmd.PersistentFlags().StringSlice("webrtc.nat1to1", []string{}, "sets a list of external IP addresses of 1:1 (D)NAT and a candidate type for which the external IP address is used")
	if err := viper.BindPFlag("webrtc.nat1to1", cmd.PersistentFlags().Lookup("webrtc.nat1to1")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("webrtc.ip_retrieval_url", "https://checkip.amazonaws.com", "URL address used for retrieval of the external IP address")
	if err := viper.BindPFlag("webrtc.ip_retrieval_url", cmd.PersistentFlags().Lookup("webrtc.ip_retrieval_url")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("webrtc.epr", fmt.Sprintf("%d-%d", defEprMin, defEprMax), "limits the pool of ephemeral ports that ICE UDP connections can allocate from")
	if err := viper.BindPFlag("webrtc.epr", cmd.PersistentFlags().Lookup("webrtc.epr")); err != nil {
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

	min := uint16(defEprMin)
	max := uint16(defEprMax)

	epr := viper.GetString("webrtc.epr")
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
