package config

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"demodesk/neko/internal/utils"
)

type Server struct {
	Cert   string
	Key    string
	Bind   string
	Static string
	CORS   []string
	//Proxy  bool
}

func (Server) Init(cmd *cobra.Command) error {
	cmd.PersistentFlags().String("bind", "127.0.0.1:8080", "address/port/socket to serve neko")
	if err := viper.BindPFlag("bind", cmd.PersistentFlags().Lookup("bind")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("cert", "", "path to the SSL cert used to secure the neko server")
	if err := viper.BindPFlag("cert", cmd.PersistentFlags().Lookup("cert")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("key", "", "path to the SSL key used to secure the neko server")
	if err := viper.BindPFlag("key", cmd.PersistentFlags().Lookup("key")); err != nil {
		return err
	}

	cmd.PersistentFlags().String("static", "", "path to neko client files to serve")
	if err := viper.BindPFlag("static", cmd.PersistentFlags().Lookup("static")); err != nil {
		return err
	}

	cmd.PersistentFlags().StringSlice("cors", []string{"*"}, "list of allowed origins for CORS")
	if err := viper.BindPFlag("cors", cmd.PersistentFlags().Lookup("cors")); err != nil {
		return err
	}

	//cmd.PersistentFlags().Bool("proxy", false, "allow reverse proxies")
	//if err := viper.BindPFlag("proxy", cmd.PersistentFlags().Lookup("proxy")); err != nil {
	//	return err
	//}

	return nil
}

func (s *Server) Set() {
	s.Cert = viper.GetString("cert")
	s.Key = viper.GetString("key")
	s.Bind = viper.GetString("bind")
	s.Static = viper.GetString("static")

	s.CORS = viper.GetStringSlice("cors")
	in, _ := utils.ArrayIn("*", s.CORS)
	if len(s.CORS) == 0 || in {
		s.CORS = []string{"*"}
	}

	//s.Proxy = viper.GetBool("proxy")
}

func (s *Server) AllowOrigin(origin string) bool {
	in, _ := utils.ArrayIn(origin, s.CORS)
	return in || s.CORS[0] == "*"
}
