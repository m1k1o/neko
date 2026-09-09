package config

import (
	"path"
	"slices"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type Server struct {
	Cert       string
	Key        string
	Bind       string
	Proxy      bool
	Static     string
	PathPrefix string
	PProf      bool
	Metrics    bool
	CORS       []string
}

func (Server) Init(cmd *cobra.Command) error {
	cmd.PersistentFlags().String("server.bind", "127.0.0.1:8080", "address/port/socket to serve neko")
	if err := viper.BindPFlag("server.bind", cmd.PersistentFlags().Lookup("server.bind")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("server.cert", "", "path to the SSL cert used to secure the neko server")
	if err := viper.BindPFlag("server.cert", cmd.PersistentFlags().Lookup("server.cert")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("server.key", "", "path to the SSL key used to secure the neko server")
	if err := viper.BindPFlag("server.key", cmd.PersistentFlags().Lookup("server.key")); err != nil {
		return err
	}

	cmd.PersistentFlags().Bool("server.proxy", false, "trust reverse proxy headers")
	if err := viper.BindPFlag("server.proxy", cmd.PersistentFlags().Lookup("server.proxy")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("server.static", "", "path to neko client files to serve")
	if err := viper.BindPFlag("server.static", cmd.PersistentFlags().Lookup("server.static")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("server.path_prefix", "/", "path prefix for HTTP requests")
	if err := viper.BindPFlag("server.path_prefix", cmd.PersistentFlags().Lookup("server.path_prefix")); err != nil {
		return err
	}

	cmd.PersistentFlags().Bool("server.pprof", false, "enable pprof endpoint available at /debug/pprof")
	if err := viper.BindPFlag("server.pprof", cmd.PersistentFlags().Lookup("server.pprof")); err != nil {
		return err
	}

	cmd.PersistentFlags().Bool("server.metrics", true, "enable prometheus metrics available at /metrics")
	if err := viper.BindPFlag("server.metrics", cmd.PersistentFlags().Lookup("server.metrics")); err != nil {
		return err
	}

	cmd.PersistentFlags().StringSlice("server.cors", []string{}, "list of allowed origins for CORS, if empty CORS is disabled, if '*' is present all origins are allowed")
	if err := viper.BindPFlag("server.cors", cmd.PersistentFlags().Lookup("server.cors")); err != nil {
		return err
	}

	return nil
}

func (s *Server) Set() {
	s.Cert = viper.GetString("server.cert")
	s.Key = viper.GetString("server.key")
	s.Bind = viper.GetString("server.bind")
	s.Proxy = viper.GetBool("server.proxy")
	s.Static = viper.GetString("server.static")
	s.PathPrefix = path.Join("/", path.Clean(viper.GetString("server.path_prefix")))
	s.PProf = viper.GetBool("server.pprof")
	s.Metrics = viper.GetBool("server.metrics")

	s.CORS = viper.GetStringSlice("server.cors")
	if slices.Contains(s.CORS, "*") {
		s.CORS = []string{"*"}
		log.Warn().Msg("CORS is enabled for all origins, this is not recommended for production environments")
	}
}

func (s *Server) HasCors() bool {
	return len(s.CORS) > 0
}

func (s *Server) AllowOrigin(origin string) bool {
	// if CORS is not configured, or a wildcard is present,
	// allow WebSocket connections from any origin
	if len(s.CORS) == 0 || s.CORS[0] == "*" {
		return true
	}

	// allow only origins in the list
	return slices.Contains(s.CORS, origin)
}
