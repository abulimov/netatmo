/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"io"
	"net"
	"os"

	netatmo2 "github.com/mariusbreivik/netatmo/api/netatmo"
	"github.com/mariusbreivik/netatmo/internal/netatmo"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var submitFlag bool
var graphiteFlag string

// graphiteCmd represents the graphite command
var graphiteCmd = &cobra.Command{
	Use:     "graphite",
	Short:   "read data from netatmo station",
	Long:    `read data from netatmo station`,
	Example: "netatmo graphite",
	RunE: func(cmd *cobra.Command, args []string) error {
		netatmoClient, err := netatmo.NewClient(netatmo.Config{
			ClientID:     viper.GetString("netatmo.clientID"),
			ClientSecret: viper.GetString("netatmo.clientSecret"),
			Username:     viper.GetString("netatmo.username"),
			Password:     viper.GetString("netatmo.password"),
		})

		if len(args) > 0 {
			fmt.Println(cmd.UsageString())
		}
		if err != nil {
			return err
		}

		return processData(netatmoClient.GetStationData(), graphiteFlag, submitFlag)
	},
}

func processData(s netatmo2.StationData, address string, submit bool) error {
	var out io.Writer
	udpAddr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		return err
	}
	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		return err
	}
	defer conn.Close()
	//now := time.Now().Unix()
	now := s.Body.Devices[0].DashboardData.TimeUtc
	if submit {
		out = conn
	} else {
		out = os.Stdout
	}
	// indoor
	fmt.Fprintf(out, "office.temperature %f %d\n", s.Body.Devices[0].DashboardData.Temperature, now)
	fmt.Fprintf(out, "office.humidity %d %d\n", s.Body.Devices[0].DashboardData.Humidity, now)
	fmt.Fprintf(out, "office.noise %d %d\n", s.Body.Devices[0].DashboardData.Noise, now)
	fmt.Fprintf(out, "office.co2 %d %d\n", s.Body.Devices[0].DashboardData.CO2, now)
	fmt.Fprintf(out, "office.pressure %f %d\n", s.Body.Devices[0].DashboardData.Pressure, now)
	// outdoor
	now = s.Body.Devices[0].Modules[0].DashboardData.TimeUtc
	fmt.Fprintf(out, "outdoor.temperature %f %d\n", s.Body.Devices[0].Modules[0].DashboardData.Temperature, now)
	fmt.Fprintf(out, "outdoor.humidity %d %d\n", s.Body.Devices[0].Modules[0].DashboardData.Humidity, now)
	return nil
}

func init() {
	rootCmd.AddCommand(graphiteCmd)
	graphiteCmd.Flags().BoolVarP(&submitFlag, "submit", "s", false, "submit data to Graphite server, or just print")
	graphiteCmd.Flags().StringVarP(&graphiteFlag, "server", "S", "127.0.0.1:2003", "graphite carbon UDP receiver address")
}
