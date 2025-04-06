package config

import (
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type WebSocket struct {
	Password      string
	AdminPassword string
	Locks         []string

	ControlProtection bool

	HeartbeatInterval int

	FileTransferEnabled bool
	FileTransferPath    string

	// WebSocket Streaming Config
	StreamingEnabled bool
	VideoCodec       string // Preferred video codec (e.g., "h264")
	AudioCodec       string // Preferred audio codec (e.g., "aac")
	VideoBitrate     uint   // Video bitrate in kbps for WebSocket streaming
	AudioBitrate     uint   // Audio bitrate in kbps for WebSocket streaming
	FragmentDuration uint   // fMP4 fragment duration in ms
}

func (WebSocket) Init(cmd *cobra.Command) error {
	cmd.PersistentFlags().String("password", "neko", "password for connecting to stream")
	if err := viper.BindPFlag("password", cmd.PersistentFlags().Lookup("password")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("password_admin", "admin", "admin password for connecting to stream")
	if err := viper.BindPFlag("password_admin", cmd.PersistentFlags().Lookup("password_admin")); err != nil {
		return err
	}

	cmd.PersistentFlags().StringSlice("locks", []string{}, "resources, that will be locked when starting (control, login)")
	if err := viper.BindPFlag("locks", cmd.PersistentFlags().Lookup("locks")); err != nil {
		return err
	}

	cmd.PersistentFlags().Bool("control_protection", false, "control protection means, users can gain control only if at least one admin is in the room")
	if err := viper.BindPFlag("control_protection", cmd.PersistentFlags().Lookup("control_protection")); err != nil {
		return err
	}

	cmd.PersistentFlags().Int("heartbeat_interval", 120, "heartbeat interval in seconds")
	if err := viper.BindPFlag("heartbeat_interval", cmd.PersistentFlags().Lookup("heartbeat_interval")); err != nil {
		return err
	}

	// File transfer

	cmd.PersistentFlags().Bool("file_transfer_enabled", false, "enable file transfer feature")
	if err := viper.BindPFlag("file_transfer_enabled", cmd.PersistentFlags().Lookup("file_transfer_enabled")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("file_transfer_path", "/home/neko/Downloads", "path to use for file transfer")
	if err := viper.BindPFlag("file_transfer_path", cmd.PersistentFlags().Lookup("file_transfer_path")); err != nil {
		return err
	}

	// WebSocket Streaming Flags
	cmd.PersistentFlags().Bool("ws_streaming", false, "enable video/audio streaming over WebSocket (fallback)")
	if err := viper.BindPFlag("ws_streaming", cmd.PersistentFlags().Lookup("ws_streaming")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("ws_video_codec", "h264", "video codec for WebSocket streaming (h264)")
	if err := viper.BindPFlag("ws_video_codec", cmd.PersistentFlags().Lookup("ws_video_codec")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("ws_audio_codec", "aac", "audio codec for WebSocket streaming (aac)")
	if err := viper.BindPFlag("ws_audio_codec", cmd.PersistentFlags().Lookup("ws_audio_codec")); err != nil {
		return err
	}

	cmd.PersistentFlags().Uint("ws_video_bitrate", 2048, "video bitrate in kbps for WebSocket streaming")
	if err := viper.BindPFlag("ws_video_bitrate", cmd.PersistentFlags().Lookup("ws_video_bitrate")); err != nil {
		return err
	}

	cmd.PersistentFlags().Uint("ws_audio_bitrate", 128, "audio bitrate in kbps for WebSocket streaming")
	if err := viper.BindPFlag("ws_audio_bitrate", cmd.PersistentFlags().Lookup("ws_audio_bitrate")); err != nil {
		return err
	}

	cmd.PersistentFlags().Uint("ws_fragment_duration", 100, "fMP4 fragment duration in milliseconds for WebSocket streaming")
	if err := viper.BindPFlag("ws_fragment_duration", cmd.PersistentFlags().Lookup("ws_fragment_duration")); err != nil {
		return err
	}

	return nil
}

func (s *WebSocket) Set() {
	s.Password = viper.GetString("password")
	s.AdminPassword = viper.GetString("password_admin")
	s.Locks = viper.GetStringSlice("locks")

	s.ControlProtection = viper.GetBool("control_protection")

	s.HeartbeatInterval = viper.GetInt("heartbeat_interval")

	s.FileTransferEnabled = viper.GetBool("file_transfer_enabled")
	s.FileTransferPath = viper.GetString("file_transfer_path")
	s.FileTransferPath = filepath.Clean(s.FileTransferPath)

	// WebSocket Streaming Config
	s.StreamingEnabled = viper.GetBool("ws_streaming")
	s.VideoCodec = viper.GetString("ws_video_codec")
	s.AudioCodec = viper.GetString("ws_audio_codec")
	s.VideoBitrate = viper.GetUint("ws_video_bitrate")
	s.AudioBitrate = viper.GetUint("ws_audio_bitrate")
	s.FragmentDuration = viper.GetUint("ws_fragment_duration")
}
