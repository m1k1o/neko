package config

import (
	"os"
	"regexp"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type Desktop struct {
	Display string

	ScreenWidth  int
	ScreenHeight int
	ScreenRate   int16
}

func (Desktop) Init(cmd *cobra.Command) error {
	cmd.PersistentFlags().String("screen", "1280x720@30", "default screen resolution and framerate")
	if err := viper.BindPFlag("screen", cmd.PersistentFlags().Lookup("screen")); err != nil {
		return err
	}

	return nil
}

func (s *Desktop) Set() {
	// Display is provided by env variable
	s.Display = os.Getenv("DISPLAY")

	s.ScreenWidth = 1280
	s.ScreenHeight = 720
	s.ScreenRate = 30

	r := regexp.MustCompile(`([0-9]{1,4})x([0-9]{1,4})@([0-9]{1,3})`)
	res := r.FindStringSubmatch(viper.GetString("screen"))

	if len(res) > 0 {
		width, err1 := strconv.ParseInt(res[1], 10, 64)
		height, err2 := strconv.ParseInt(res[2], 10, 64)
		rate, err3 := strconv.ParseInt(res[3], 10, 64)

		if err1 == nil && err2 == nil && err3 == nil {
			s.ScreenWidth = int(width)
			s.ScreenHeight = int(height)
			s.ScreenRate = int16(rate)
		}
	}
}
