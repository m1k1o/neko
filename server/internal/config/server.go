package config

import (
	"path"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/m1k1o/neko/server/pkg/utils"
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

func (Server) InitV2(cmd *cobra.Command) error {
	cmd.PersistentFlags().String("bind", "", "V2: address/port/socket to serve neko")
	if err := viper.BindPFlag("bind", cmd.PersistentFlags().Lookup("bind")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("cert", "", "V2: path to the SSL cert used to secure the neko server")
	if err := viper.BindPFlag("cert", cmd.PersistentFlags().Lookup("cert")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("key", "", "V2: path to the SSL key used to secure the neko server")
	if err := viper.BindPFlag("key", cmd.PersistentFlags().Lookup("key")); err != nil {
		return err
	}

	cmd.PersistentFlags().Bool("proxy", false, "V2: enable reverse proxy mode")
	if err := viper.BindPFlag("proxy", cmd.PersistentFlags().Lookup("proxy")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("static", "", "V2: path to neko client files to serve")
	if err := viper.BindPFlag("static", cmd.PersistentFlags().Lookup("static")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("path_prefix", "", "V2: path prefix for HTTP requests")
	if err := viper.BindPFlag("path_prefix", cmd.PersistentFlags().Lookup("path_prefix")); err != nil {
		return err
	}

	cmd.PersistentFlags().StringSlice("cors", []string{}, "V2: list of allowed origins for CORS")
	if err := viper.BindPFlag("cors", cmd.PersistentFlags().Lookup("cors")); err != nil {
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
	in, _ := utils.ArrayIn("*", s.CORS)
	if len(s.CORS) == 0 || in {
		s.CORS = []string{"*"}
	}
}

func (s *Server) SetV2() {
	enableLegacy := false

	if viper.IsSet("cert") {
		s.Cert = viper.GetString("cert")
		log.Warn().Msg("you are using v2 configuration 'NEKO_CERT' which is deprecated, please use 'NEKO_SERVER_CERT' instead")
		enableLegacy = true
	}
	if viper.IsSet("key") {
		s.Key = viper.GetString("key")
		log.Warn().Msg("you are using v2 configuration 'NEKO_KEY' which is deprecated, please use 'NEKO_SERVER_KEY' instead")
		enableLegacy = true
	}
	if viper.IsSet("bind") {
		s.Bind = viper.GetString("bind")
		log.Warn().Msg("you are using v2 configuration 'NEKO_BIND' which is deprecated, please use 'NEKO_SERVER_BIND' instead")
		enableLegacy = true
	}
	if viper.IsSet("proxy") {
		s.Proxy = viper.GetBool("proxy")
		log.Warn().Msg("you are using v2 configuration 'NEKO_PROXY' which is deprecated, please use 'NEKO_SERVER_PROXY' instead")
		enableLegacy = true
	}
	if viper.IsSet("static") {
		s.Static = viper.GetString("static")
		log.Warn().Msg("you are using v2 configuration 'NEKO_STATIC' which is deprecated, please use 'NEKO_SERVER_STATIC' instead")
		enableLegacy = true
	}
	if viper.IsSet("path_prefix") {
		s.PathPrefix = path.Join("/", path.Clean(viper.GetString("path_prefix")))
		log.Warn().Msg("you are using v2 configuration 'NEKO_PATH_PREFIX' which is deprecated, please use 'NEKO_SERVER_PATH_PREFIX' instead")
		enableLegacy = true
	}
	if viper.IsSet("cors") {
		s.CORS = viper.GetStringSlice("cors")
		in, _ := utils.ArrayIn("*", s.CORS)
		if len(s.CORS) == 0 || in {
			s.CORS = []string{"*"}
		}
		log.Warn().Msg("you are using v2 configuration 'NEKO_CORS' which is deprecated, please use 'NEKO_SERVER_CORS' instead")
		enableLegacy = true
	}

	// set legacy flag if any V2 configuration was used
	if !viper.IsSet("legacy") && enableLegacy {
		log.Warn().Msg("legacy configuration is enabled because at least one V2 configuration was used, please migrate to V3 configuration, visit https://neko.m1k1o.net/docs/v3/migration-from-v2 for more details")
		viper.Set("legacy", true)
	}
}

func (s *Server) HasCors() bool {
	return len(s.CORS) > 0
}

func (s *Server) AllowOrigin(origin string) bool {
	// if CORS is disabled, allow all origins
	if len(s.CORS) == 0 {
		return true
	}

	// if CORS is enabled, allow only origins in the list
	in, _ := utils.ArrayIn(origin, s.CORS)
	return in || s.CORS[0] == "*"
}
