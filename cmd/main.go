package main

import (
	"github.com/sirupsen/logrus"
)

func main() {
	rootCmd := rootCmd()
	if err := rootCmd.Execute(); err != nil {
		logrus.Fatal(err)
	}
}
