package main

import (
	"os"

	"github.com/sirupsen/logrus"

	"laptudirm.com/x/arbiter/internal/arbiter/cmd"
)

func main() {
	logrus.SetFormatter(&logrus.TextFormatter{
		DisableTimestamp: true,
		PadLevelText:     true,
	})
	logrus.SetLevel(logrus.InfoLevel)

	if err := arbiter(); err != nil {
		logrus.Fatal(err)
	}
}

func arbiter() error {
	root := cmd.Root()
	root.SetArgs(os.Args[1:])
	return root.Execute()
}
