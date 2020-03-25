/*
Copyright © 2020 allen <aiddroid@gmail.com>

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
	"log"
	"os"
	"tcp-proxy/core"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var fromPort string
var toPort string
var whiteIpFile string
var isDump bool

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start TCP proxy",
	Long: `Start TCP proxy from a port to another.`,
	Run: func(cmd *cobra.Command, args []string) {
		if logFile != "" {
			logWriter, err := os.OpenFile(logFile, os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
			if err == nil {
				log.SetOutput(logWriter)
			}
		}

		proxyPort := fromPort
		targetPort := toPort

		if proxyPort == "" || targetPort == "" {
			log.Println("Usage example: tcp-proxy start 443 8379")
			return
		}

		log.Printf("Starting tcp-proxy from %s to %s", proxyPort, targetPort)
		core.StartServer(proxyPort, targetPort, whiteIpFile, isDump)
	},
}

func init() {
	rootCmd.AddCommand(startCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// startCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// startCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is $HOME/.tcp-proxy.yaml)")
	rootCmd.PersistentFlags().StringVarP(&logFile, "logfile", "l", "", "log file path (default is STDOUT)")

	// 定义一系列命令行参数
	startCmd.Flags().StringVarP(&fromPort, "from", "f", "", "From port, ig:7777")
	startCmd.MarkFlagRequired("from")

	startCmd.Flags().StringVarP(&toPort, "to", "t", "", "To port, ig:8080")
	startCmd.MarkFlagRequired("to")

	startCmd.Flags().StringVarP(&whiteIpFile, "whiteip", "w", "whiteip.txt", "White ip list file path")

	startCmd.Flags().BoolVarP(&isDump, "dump", "D", false, "Dump all data")

	// 绑定参数到viper,以便能从配置文件读取参数
	viper.BindPFlags(startCmd.Flags())
	startCmd.Flags().SortFlags = false
}