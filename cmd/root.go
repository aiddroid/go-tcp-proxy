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
  "fmt"
  "github.com/spf13/cobra"
  "log"
  "os"

  "github.com/mitchellh/go-homedir"
  "github.com/spf13/viper"
)


var cfgFile string
var logFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
  Use:   "tcp-proxy",
  Short: "Start TCP proxy",
  Long: `Start TCP proxy from a port to another.`,
  // Uncomment the following line if your bare application
  // has an action associated with it:
  // Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
  if err := rootCmd.Execute(); err != nil {
    fmt.Println(err)
    os.Exit(1)
  }
}

func init() {
  cobra.OnInitialize(initConfig)
  cobra.OnInitialize(initLogFile)

  // Here you will define your flags and configuration settings.
  // Cobra supports persistent flags, which, if defined here,
  // will be global for your application.
  rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is $HOME/.tcp-proxy.yaml)")
  rootCmd.PersistentFlags().StringVarP(&logFile, "logfile", "l", "", "log file path (default is STDOUT)")

  // Cobra also supports local flags, which will only run
  // when this action is called directly.

  // 绑定参数到viper,以便能从配置文件读取参数
  viper.BindPFlags(rootCmd.Flags())
  rootCmd.Flags().SortFlags = false
}


// initConfig reads in config file and ENV variables if set.
func initConfig() {
  if cfgFile != "" {
    // Use config file from the flag.
    viper.SetConfigFile(cfgFile)
  } else {
    // Find home directory.
    home, err := homedir.Dir()
    if err != nil {
      fmt.Println(err)
      os.Exit(1)
    }

    // Search config in home directory with name ".tcp-proxy" (without extension).
    viper.AddConfigPath(home)
    viper.SetConfigName(".tcp-proxy")
  }

  viper.AutomaticEnv() // read in environment variables that match

  // If a config file is found, read it in.
  if err := viper.ReadInConfig(); err == nil {
    fmt.Println("Using config file:", viper.ConfigFileUsed())
  }
}

func initLogFile() {
  if logFile != "" {
    logWriter, err := os.OpenFile(logFile, os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
    if err == nil {
      log.SetOutput(logWriter)
    }
  }
}

