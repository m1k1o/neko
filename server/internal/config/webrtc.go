package config

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/m1k1o/neko/server/pkg/types"
	"github.com/m1k1o/neko/server/pkg/utils"
)

// default stun server
const defStunSrv = "stun:stun.l.google.com:19302"

type WebRTCEstimator struct {
	Enabled        bool
	Passive        bool
	Debug          bool
	InitialBitrate int

	// how often to read and process bandwidth estimation reports
	ReadInterval time.Duration
	// how long to wait for stable connection (only neutral or upward trend) before upgrading
	StableDuration time.Duration
	// how long to wait for unstable connection (downward trend) before downgrading
	UnstableDuration time.Duration
	// how long to wait for stalled connection (neutral trend with low bandwidth) before downgrading
	StalledDuration time.Duration
	// how long to wait before downgrading again after previous downgrade
	DowngradeBackoff time.Duration
	// how long to wait before upgrading again after previous upgrade
	UpgradeBackoff time.Duration
	// how bigger the difference between estimated and stream bitrate must be to trigger upgrade/downgrade
	DiffThreshold float64
}

type WebRTC struct {
	ICELite            bool
	ICETrickle         bool
	ICEServersFrontend []types.ICEServer
	ICEServersBackend  []types.ICEServer
	EphemeralMin       uint16
	EphemeralMax       uint16
	TCPMux             int
	UDPMux             int

	NAT1To1IPs     []string
	IpRetrievalUrl string

	Estimator WebRTCEstimator
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

	// Looks like this is conflicting with the frontend and backend ICE servers since latest versions
	//cmd.PersistentFlags().String("webrtc.iceservers", "[]", "STUN and TURN servers used by the ICE agent")
	//if err := viper.BindPFlag("webrtc.iceservers", cmd.PersistentFlags().Lookup("webrtc.iceservers")); err != nil {
	//	return err
	//}

	cmd.PersistentFlags().String("webrtc.iceservers.frontend", "[]", "STUN and TURN servers used by the frontend")
	if err := viper.BindPFlag("webrtc.iceservers.frontend", cmd.PersistentFlags().Lookup("webrtc.iceservers.frontend")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("webrtc.iceservers.backend", "[]", "STUN and TURN servers used by the backend")
	if err := viper.BindPFlag("webrtc.iceservers.backend", cmd.PersistentFlags().Lookup("webrtc.iceservers.backend")); err != nil {
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

	cmd.PersistentFlags().Bool("webrtc.estimator.passive", false, "passive estimator mode, when it does not switch pipelines, only estimates")
	if err := viper.BindPFlag("webrtc.estimator.passive", cmd.PersistentFlags().Lookup("webrtc.estimator.passive")); err != nil {
		return err
	}

	cmd.PersistentFlags().Bool("webrtc.estimator.debug", false, "enables debug logging for the bandwidth estimator")
	if err := viper.BindPFlag("webrtc.estimator.debug", cmd.PersistentFlags().Lookup("webrtc.estimator.debug")); err != nil {
		return err
	}

	cmd.PersistentFlags().Int("webrtc.estimator.initial_bitrate", 1_000_000, "initial bitrate for the bandwidth estimator")
	if err := viper.BindPFlag("webrtc.estimator.initial_bitrate", cmd.PersistentFlags().Lookup("webrtc.estimator.initial_bitrate")); err != nil {
		return err
	}

	cmd.PersistentFlags().Duration("webrtc.estimator.read_interval", 2*time.Second, "how often to read and process bandwidth estimation reports")
	if err := viper.BindPFlag("webrtc.estimator.read_interval", cmd.PersistentFlags().Lookup("webrtc.estimator.read_interval")); err != nil {
		return err
	}

	cmd.PersistentFlags().Duration("webrtc.estimator.stable_duration", 12*time.Second, "how long to wait for stable connection (upward or neutral trend) before upgrading")
	if err := viper.BindPFlag("webrtc.estimator.stable_duration", cmd.PersistentFlags().Lookup("webrtc.estimator.stable_duration")); err != nil {
		return err
	}

	cmd.PersistentFlags().Duration("webrtc.estimator.unstable_duration", 6*time.Second, "how long to wait for stalled connection (neutral trend with low bandwidth) before downgrading")
	if err := viper.BindPFlag("webrtc.estimator.unstable_duration", cmd.PersistentFlags().Lookup("webrtc.estimator.unstable_duration")); err != nil {
		return err
	}

	cmd.PersistentFlags().Duration("webrtc.estimator.stalled_duration", 24*time.Second, "how long to wait for stalled bandwidth estimation before downgrading")
	if err := viper.BindPFlag("webrtc.estimator.stalled_duration", cmd.PersistentFlags().Lookup("webrtc.estimator.stalled_duration")); err != nil {
		return err
	}

	cmd.PersistentFlags().Duration("webrtc.estimator.downgrade_backoff", 10*time.Second, "how long to wait before downgrading again after previous downgrade")
	if err := viper.BindPFlag("webrtc.estimator.downgrade_backoff", cmd.PersistentFlags().Lookup("webrtc.estimator.downgrade_backoff")); err != nil {
		return err
	}

	cmd.PersistentFlags().Duration("webrtc.estimator.upgrade_backoff", 5*time.Second, "how long to wait before upgrading again after previous upgrade")
	if err := viper.BindPFlag("webrtc.estimator.upgrade_backoff", cmd.PersistentFlags().Lookup("webrtc.estimator.upgrade_backoff")); err != nil {
		return err
	}

	cmd.PersistentFlags().Float64("webrtc.estimator.diff_threshold", 0.15, "how bigger the difference between estimated and stream bitrate must be to trigger upgrade/downgrade")
	if err := viper.BindPFlag("webrtc.estimator.diff_threshold", cmd.PersistentFlags().Lookup("webrtc.estimator.diff_threshold")); err != nil {
		return err
	}

	return nil
}

func (WebRTC) InitV2(cmd *cobra.Command) error {
	cmd.PersistentFlags().String("epr", "", "V2: limits the pool of ephemeral ports that ICE UDP connections can allocate from")
	if err := viper.BindPFlag("epr", cmd.PersistentFlags().Lookup("epr")); err != nil {
		return err
	}

	cmd.PersistentFlags().StringSlice("nat1to1", []string{}, "V2: sets a list of external IP addresses of 1:1 (D)NAT and a candidate type for which the external IP address is used")
	if err := viper.BindPFlag("nat1to1", cmd.PersistentFlags().Lookup("nat1to1")); err != nil {
		return err
	}

	cmd.PersistentFlags().Int("tcpmux", 0, "V2: single TCP mux port for all peers")
	if err := viper.BindPFlag("tcpmux", cmd.PersistentFlags().Lookup("tcpmux")); err != nil {
		return err
	}

	cmd.PersistentFlags().Int("udpmux", 0, "V2: single UDP mux port for all peers")
	if err := viper.BindPFlag("udpmux", cmd.PersistentFlags().Lookup("udpmux")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("ipfetch", "", "V2: automatically fetch IP address from given URL when nat1to1 is not present")
	if err := viper.BindPFlag("ipfetch", cmd.PersistentFlags().Lookup("ipfetch")); err != nil {
		return err
	}

	cmd.PersistentFlags().Bool("icelite", false, "V2: configures whether or not the ice agent should be a lite agent")
	if err := viper.BindPFlag("icelite", cmd.PersistentFlags().Lookup("icelite")); err != nil {
		return err
	}

	cmd.PersistentFlags().StringSlice("iceserver", []string{}, "V2: describes a single STUN and TURN server that can be used by the ICEAgent to establish a connection with a peer")
	if err := viper.BindPFlag("iceserver", cmd.PersistentFlags().Lookup("iceserver")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("iceservers", "", "V2: describes a single STUN and TURN server that can be used by the ICEAgent to establish a connection with a peer")
	if err := viper.BindPFlag("iceservers", cmd.PersistentFlags().Lookup("iceservers")); err != nil {
		return err
	}

	return nil
}

func (s *WebRTC) Set() {
	s.ICELite = viper.GetBool("webrtc.icelite")
	s.ICETrickle = viper.GetBool("webrtc.icetrickle")

	// parse frontend ice servers
	if err := viper.UnmarshalKey("webrtc.iceservers.frontend", &s.ICEServersFrontend, viper.DecodeHook(
		utils.JsonStringAutoDecode(s.ICEServersFrontend),
	)); err != nil {
		log.Warn().Err(err).Msgf("unable to parse frontend ICE servers")
	}

	// parse backend ice servers
	if err := viper.UnmarshalKey("webrtc.iceservers.backend", &s.ICEServersBackend, viper.DecodeHook(
		utils.JsonStringAutoDecode(s.ICEServersBackend),
	)); err != nil {
		log.Warn().Err(err).Msgf("unable to parse backend ICE servers")
	}

	if s.ICELite && len(s.ICEServersBackend) > 0 {
		log.Warn().Msgf("ICE Lite is enabled, but backend ICE servers are configured. Backend ICE servers will be ignored.")
	}

	// if no frontend or backend ice servers are configured
	if len(s.ICEServersFrontend) == 0 && len(s.ICEServersBackend) == 0 {
		// parse global ice servers
		var iceServers []types.ICEServer
		if err := viper.UnmarshalKey("webrtc.iceservers", &iceServers, viper.DecodeHook(
			utils.JsonStringAutoDecode(iceServers),
		)); err != nil {
			log.Warn().Err(err).Msgf("unable to parse global ICE servers")
		}

		// add default stun server if none are configured
		if len(iceServers) == 0 {
			iceServers = append(iceServers, types.ICEServer{
				URLs: []string{defStunSrv},
			})
		}

		s.ICEServersFrontend = append(s.ICEServersFrontend, iceServers...)
		s.ICEServersBackend = append(s.ICEServersBackend, iceServers...)
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

	s.Estimator.Enabled = viper.GetBool("webrtc.estimator.enabled")
	s.Estimator.Passive = viper.GetBool("webrtc.estimator.passive")
	s.Estimator.Debug = viper.GetBool("webrtc.estimator.debug")
	s.Estimator.InitialBitrate = viper.GetInt("webrtc.estimator.initial_bitrate")
	s.Estimator.ReadInterval = viper.GetDuration("webrtc.estimator.read_interval")
	s.Estimator.StableDuration = viper.GetDuration("webrtc.estimator.stable_duration")
	s.Estimator.UnstableDuration = viper.GetDuration("webrtc.estimator.unstable_duration")
	s.Estimator.StalledDuration = viper.GetDuration("webrtc.estimator.stalled_duration")
	s.Estimator.DowngradeBackoff = viper.GetDuration("webrtc.estimator.downgrade_backoff")
	s.Estimator.UpgradeBackoff = viper.GetDuration("webrtc.estimator.upgrade_backoff")
	s.Estimator.DiffThreshold = viper.GetFloat64("webrtc.estimator.diff_threshold")
}

func (s *WebRTC) SetV2() {
	enableLegacy := false

	if viper.IsSet("nat1to1") {
		s.NAT1To1IPs = viper.GetStringSlice("nat1to1")
		log.Warn().Msg("you are using v2 configuration 'NEKO_NAT1TO1' which is deprecated, please use 'NEKO_WEBRTC_NAT1TO1' instead")
		enableLegacy = true
	}
	if viper.IsSet("tcpmux") {
		s.TCPMux = viper.GetInt("tcpmux")
		log.Warn().Msg("you are using v2 configuration 'NEKO_TCPMUX' which is deprecated, please use 'NEKO_WEBRTC_TCPMUX' instead")
		enableLegacy = true
	}
	if viper.IsSet("udpmux") {
		s.UDPMux = viper.GetInt("udpmux")
		log.Warn().Msg("you are using v2 configuration 'NEKO_UDPMUX' which is deprecated, please use 'NEKO_WEBRTC_UDPMUX' instead")
		enableLegacy = true
	}
	if viper.IsSet("icelite") {
		s.ICELite = viper.GetBool("icelite")
		log.Warn().Msg("you are using v2 configuration 'NEKO_ICELITE' which is deprecated, please use 'NEKO_WEBRTC_ICELITE' instead")
		enableLegacy = true
	}

	if viper.IsSet("iceservers") {
		iceServers := []types.ICEServer{}
		iceServersJson := viper.GetString("iceservers")
		if iceServersJson != "" {
			err := json.Unmarshal([]byte(iceServersJson), &iceServers)
			if err != nil {
				log.Panic().Err(err).Msg("failed to process iceservers")
			}
		}
		s.ICEServersFrontend = iceServers
		s.ICEServersBackend = iceServers
		log.Warn().Msg("you are using v2 configuration 'NEKO_ICESERVERS' which is deprecated, please use 'NEKO_WEBRTC_ICESERVERS_FRONTEND' and/or 'NEKO_WEBRTC_ICESERVERS_BACKEND' instead")
		enableLegacy = true
	}

	if viper.IsSet("iceserver") {
		iceServerSlice := viper.GetStringSlice("iceserver")
		if len(iceServerSlice) > 0 {
			s.ICEServersFrontend = append(s.ICEServersFrontend, types.ICEServer{URLs: iceServerSlice})
			s.ICEServersBackend = append(s.ICEServersBackend, types.ICEServer{URLs: iceServerSlice})
		}
		log.Warn().Msg("you are using v2 configuration 'NEKO_ICESERVER' which is deprecated, please use 'NEKO_WEBRTC_ICESERVERS_FRONTEND' and/or 'NEKO_WEBRTC_ICESERVERS_BACKEND' instead")
		enableLegacy = true
	}

	if viper.IsSet("ipfetch") {
		if len(s.NAT1To1IPs) == 0 {
			ipfetch := viper.GetString("ipfetch")
			ip, err := utils.HttpRequestGET(ipfetch)
			if err != nil {
				log.Panic().Err(err).Str("ipfetch", ipfetch).Msg("failed to fetch ip address")
			}
			s.NAT1To1IPs = append(s.NAT1To1IPs, ip)
		}
		log.Warn().Msg("you are using v2 configuration 'NEKO_IPFETCH' which is deprecated, please use 'NEKO_WEBRTC_IP_RETRIEVAL_URL' instead")
		enableLegacy = true
	}

	if viper.IsSet("epr") {
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
		log.Warn().Msg("you are using v2 configuration 'NEKO_EPR' which is deprecated, please use 'NEKO_WEBRTC_EPR' instead")
		enableLegacy = true
	}

	// set legacy flag if any V2 configuration was used
	if !viper.IsSet("legacy") && enableLegacy {
		log.Warn().Msg("legacy configuration is enabled because at least one V2 configuration was used, please migrate to V3 configuration, visit https://neko.m1k1o.net/docs/v3/migration-from-v2 for more details")
		viper.Set("legacy", true)
	}
}
