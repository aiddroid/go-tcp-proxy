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
	"github.com/thoas/go-funk"
	"go-tcp-proxy/core"
	"log"
	"math/rand"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var pCfg core.ProxyCfg

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start TCP proxy",
	Long: `Start TCP proxy from a port to another.`,
	Run: func(cmd *cobra.Command, args []string) {
		if pCfg.AuthUri == "" {
			rand.Seed(time.Now().UnixNano())
			pCfg.AuthUri = "/auth/" + funk.RandomString(16)
		}

		log.Println("Starting TCP proxy with config:", pCfg)
		core.StartServer(pCfg)
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

	// 定义一系列命令行参数
	// 对外暴露的端口
	startCmd.Flags().StringVarP(&pCfg.FromPort, "from", "f", "", "Exposed port, E.g:443")
	startCmd.MarkFlagRequired("from")

	// 要转发的端口
	startCmd.Flags().StringVarP(&pCfg.ToPort, "to", "t", "", "Port needs to be proxied, E.g:8379")
	startCmd.MarkFlagRequired("to")

	// 白名单文件路径
	startCmd.Flags().StringVarP(&pCfg.WhiteIpFile, "whiteip", "w", "whiteip.json", "White IP list file path")

	// 新IP认证URI
	startCmd.Flags().StringVarP(&pCfg.AuthUri, "auth", "a", "", "New IP auth URI")

	// 默认html页面文件路径
	startCmd.Flags().StringVarP(&pCfg.HtmlFile, "html", "H", "", "HTML file path for filtered IPs")

	// 是否dump全部数据
	startCmd.Flags().BoolVar(&pCfg.IsDump, "dump", false, "If all data dumped as logs")

	// 绑定参数到viper,以便能从配置文件读取参数
	viper.BindPFlags(startCmd.Flags())
	startCmd.Flags().SortFlags = false
}