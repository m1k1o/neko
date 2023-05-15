package config

import (
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/demodesk/neko/pkg/types"
	"github.com/demodesk/neko/pkg/utils"
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

	cmd.PersistentFlags().String("webrtc.iceservers", "[]", "Global STUN and TURN servers in JSON format with `urls`, `username` and `credential` keys")
	if err := viper.BindPFlag("webrtc.iceservers", cmd.PersistentFlags().Lookup("webrtc.iceservers")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("webrtc.iceservers.frontend", "[]", "Frontend only STUN and TURN servers in JSON format with `urls`, `username` and `credential` keys")
	if err := viper.BindPFlag("webrtc.iceservers.frontend", cmd.PersistentFlags().Lookup("webrtc.iceservers.frontend")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("webrtc.iceservers.backend", "[]", "Backend only STUN and TURN servers in JSON format with `urls`, `username` and `credential` keys")
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

func (s *WebRTC) Set() {
	s.ICELite = viper.GetBool("webrtc.icelite")
	s.ICETrickle = viper.GetBool("webrtc.icetrickle")

	// parse frontend ice servers
	if err := viper.UnmarshalKey("webrtc.iceservers.frontend", &s.ICEServersFrontend, viper.DecodeHook(
		utils.JsonStringAutoDecode([]types.ICEServer{}),
	)); err != nil {
		log.Warn().Err(err).Msgf("unable to parse frontend ICE servers")
	}

	// parse backend ice servers
	if err := viper.UnmarshalKey("webrtc.iceservers.backend", &s.ICEServersBackend, viper.DecodeHook(
		utils.JsonStringAutoDecode([]types.ICEServer{}),
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
			utils.JsonStringAutoDecode([]types.ICEServer{}),
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
