package config

import (
	"strconv"
	"strings"

	"github.com/pion/webrtc/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"n.eko.moe/neko/internal/utils"
)

type WebRTC struct {
	Device       string
	AudioCodec   string
	AudioParams  string
	Display      string
	VideoCodec   string
	VideoParams  string
	EphemeralMin uint16
	EphemeralMax uint16
	NAT1To1IPs   []string
}

func (WebRTC) Init(cmd *cobra.Command) error {
	cmd.PersistentFlags().String("device", "auto_null.monitor", "audio device to capture")
	if err := viper.BindPFlag("device", cmd.PersistentFlags().Lookup("device")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("aduio", "", "audio codec parameters to use for streaming (unused)")
	if err := viper.BindPFlag("aparams", cmd.PersistentFlags().Lookup("aduio")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("display", ":99.0", "XDisplay to capture")
	if err := viper.BindPFlag("display", cmd.PersistentFlags().Lookup("display")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("video", "", "video codec parameters to use for streaming (unused)")
	if err := viper.BindPFlag("vparams", cmd.PersistentFlags().Lookup("video")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("epr", "59000-59100", "limits the pool of ephemeral ports that ICE UDP connections can allocate from")
	if err := viper.BindPFlag("epr", cmd.PersistentFlags().Lookup("epr")); err != nil {
		return err
	}

	// video codecs
	cmd.PersistentFlags().Bool("vp8", false, "use VP8 video codec")
	if err := viper.BindPFlag("vp8", cmd.PersistentFlags().Lookup("vp8")); err != nil {
		return err
	}

	cmd.PersistentFlags().Bool("vp9", false, "use VP9 video codec")
	if err := viper.BindPFlag("vp9", cmd.PersistentFlags().Lookup("vp9")); err != nil {
		return err
	}

	cmd.PersistentFlags().Bool("h264", false, "use H264 video codec")
	if err := viper.BindPFlag("h264", cmd.PersistentFlags().Lookup("h264")); err != nil {
		return err
	}

	// audio codecs
	cmd.PersistentFlags().Bool("opus", false, "use Opus audio codec")
	if err := viper.BindPFlag("opus", cmd.PersistentFlags().Lookup("opus")); err != nil {
		return err
	}

	cmd.PersistentFlags().Bool("g722", false, "use G722 audio codec")
	if err := viper.BindPFlag("g722", cmd.PersistentFlags().Lookup("g722")); err != nil {
		return err
	}

	cmd.PersistentFlags().Bool("pcmu", false, "use PCMU audio codec")
	if err := viper.BindPFlag("pcmu", cmd.PersistentFlags().Lookup("pcmu")); err != nil {
		return err
	}

	cmd.PersistentFlags().Bool("pcma", false, "use PCMA audio codec")
	if err := viper.BindPFlag("pcma", cmd.PersistentFlags().Lookup("pcma")); err != nil {
		return err
	}

	cmd.PersistentFlags().StringSlice("ip", []string{}, "sets a list of external IP addresses of 1:1 (D)NAT and a candidate type for which the external IP address is used")
	if err := viper.BindPFlag("ip", cmd.PersistentFlags().Lookup("ip")); err != nil {
		return err
	}

	return nil
}

func (s *WebRTC) Set() {
	videoCodec := webrtc.VP8
	if viper.GetBool("vp8") {
		videoCodec = webrtc.VP8
	} else if viper.GetBool("vp9") {
		videoCodec = webrtc.VP9
	} else if viper.GetBool("h264") {
		videoCodec = webrtc.H264
	}

	audioCodec := webrtc.Opus
	if viper.GetBool("opus") {
		audioCodec = webrtc.Opus
	} else if viper.GetBool("g722") {
		audioCodec = webrtc.G722
	} else if viper.GetBool("pcmu") {
		audioCodec = webrtc.PCMU
	} else if viper.GetBool("pcma") {
		audioCodec = webrtc.PCMA
	}

	s.Device = viper.GetString("device")
	s.AudioCodec = audioCodec
	s.AudioParams = viper.GetString("aparams")
	s.Display = viper.GetString("display")
	s.VideoCodec = videoCodec
	s.VideoParams = viper.GetString("vparams")
	s.NAT1To1IPs = viper.GetStringSlice("ip")

	ip, err := utils.GetIP()
	if err == nil {
		s.NAT1To1IPs = append(s.NAT1To1IPs, ip)
	}

	min := uint16(59000)
	max := uint16(59100)
	epr := viper.GetString("epr")
	ports := strings.SplitN(epr, "-", -1)
	if len(ports[0]) > 1 {
		start, err := strconv.ParseUint(ports[0], 16, 16)
		if err == nil {
			min = uint16(start)
		}

		end, err := strconv.ParseUint(ports[1], 16, 16)
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
