package config

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/demodesk/neko/pkg/utils"
)

type Server struct {
	Cert    string
	Key     string
	Bind    string
	Static  string
	PProf   bool
	Metrics bool
	CORS    []string
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

	cmd.PersistentFlags().String("server.static", "", "path to neko client files to serve")
	if err := viper.BindPFlag("server.static", cmd.PersistentFlags().Lookup("server.static")); err != nil {
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

	cmd.PersistentFlags().StringSlice("server.cors", []string{"*"}, "list of allowed origins for CORS")
	if err := viper.BindPFlag("server.cors", cmd.PersistentFlags().Lookup("server.cors")); err != nil {
		return err
	}

	return nil
}

func (s *Server) Set() {
	s.Cert = viper.GetString("server.cert")
	s.Key = viper.GetString("server.key")
	s.Bind = viper.GetString("server.bind")
	s.Static = viper.GetString("server.static")
	s.PProf = viper.GetBool("server.pprof")
	s.Metrics = viper.GetBool("server.metrics")

	s.CORS = viper.GetStringSlice("server.cors")
	in, _ := utils.ArrayIn("*", s.CORS)
	if len(s.CORS) == 0 || in {
		s.CORS = []string{"*"}
	}
}

func (s *Server) AllowOrigin(origin string) bool {
	in, _ := utils.ArrayIn(origin, s.CORS)
	return in || s.CORS[0] == "*"
}
