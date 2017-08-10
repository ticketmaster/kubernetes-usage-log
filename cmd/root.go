// Copyright © 2017 Kraig Amador <kraig@bigkraig.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/ticketmaster/kubernetes-usage-log/core/pkg/catalog"

	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
	restapi "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type RootConfigT struct {
	internal        bool
	clusterID       string
	usagePeriod     int64
	kubeconfig      string
	destinationPath string
	config          *restapi.Config
}

var RootConfig *RootConfigT

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "usage-log",
	Short: "Kubernetes usage logger",
	Long:  `This application will log Kubernetes usage metrics to a json log file to be used for accounting and billing processes.`,
	Run:   run,
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	RootConfig = new(RootConfigT)

	kubeconfigDefault := filepath.Join(homeDir(), ".kube", "config")

	RootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	RootCmd.Flags().BoolVar(&RootConfig.internal, "internal", true, "Running internal to cluster")
	RootCmd.Flags().StringVar(&RootConfig.clusterID, "id", "", "Unique ID to represent the cluster")
	RootCmd.Flags().Int64Var(&RootConfig.usagePeriod, "usagePeriod", 60, "Number of seconds per collection interval")
	RootCmd.Flags().StringVar(&RootConfig.kubeconfig, "kubeconfig", kubeconfigDefault, "absolute path to the kubeconfig file")
	RootCmd.Flags().StringVarP(&RootConfig.destinationPath, "destinationPath", "d", "logs/", "Destination path for usage logs")

	cobra.OnInitialize(initConfig)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if RootConfig.clusterID == "" {
		log.Fatal("A cluster id must be provided.")
	}
	if RootConfig.internal {
		RootConfig.config = insideCluster()
	} else {
		RootConfig.config = outsideCluster()
	}
}

func run(cmd *cobra.Command, args []string) {
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(RootConfig.config)
	if err != nil {
		panic(err.Error())
	}

	// Start the cataloging service
	stopCatalog := make(chan struct{})
	defer close(stopCatalog)
	go catalog.GenerateCatalog(clientset, stopCatalog, RootConfig.clusterID, RootConfig.usagePeriod, RootConfig.destinationPath)

	// Wait forever
	select {}
}

func insideCluster() (config *restapi.Config) {
	// creates the in-cluster config
	config, err := restapi.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	return
}

func outsideCluster() (config *restapi.Config) {
	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", RootConfig.kubeconfig)
	if err != nil {
		panic(err.Error())
	}
	return
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}
