// Copyright © 2023 Rak Laptudirm <rak@laptudirm.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package match

import (
	"bufio"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

type EngineConfig struct {
	Name string `yaml:"name"`
	Cmd  string `yaml:"cmd"`
	Dir  string `yaml:"dir"`
	Arg  string `yaml:"arg"`

	Protocol string `yaml:"protocol"`

	Stderr string `yaml:"stderr"`

	InitStr string `yaml:"init-string"`

	Options map[string]string `yaml:"options"`

	TimeC string `yaml:"tc"`
	Depth int    `yaml:"depth"`
	Nodes int    `yaml:"nodes"`
}

func StartEngine(config EngineConfig) (*Engine, error) {
	var engine Engine
	engine.protocol = config.Protocol
	process := exec.Command(config.Cmd, strings.Fields(config.Arg)...)

	engine.config = config

	process.Dir = config.Dir

	stdin, _ := process.StdinPipe()
	stdout, _ := process.StdoutPipe()

	engine.writer = bufio.NewWriter(stdin)
	engine.reader = bufio.NewReader(stdout)
	engine.lines = make(chan string)

	engine.Cmd = process

	if err := engine.Cmd.Start(); err != nil {
		return nil, err
	}

	go func() {
		for {
			line, err := engine.reader.ReadString('\n')
			if err != nil {
				engine.err = err
				close(engine.lines)
				return
			}

			line = strings.Trim(line, " \n\t\r")

			logrus.Debugf("info: ("+engine.config.Name+")> %s\n", line)
			engine.lines <- line
		}
	}()

	if engine.config.InitStr != "" {
		if err := engine.Write(engine.config.InitStr); err != nil {
			return nil, err
		}
	}

	if err := engine.Initialize(); err != nil {
		return nil, err
	}

	if err := engine.NewGame(); err != nil {
		return nil, err
	}

	return &engine, nil
}

type Engine struct {
	config EngineConfig

	*exec.Cmd

	protocol string

	writer *bufio.Writer
	reader *bufio.Reader

	lines chan string

	err error
}

// NewGame prepares the engine for a new game of chess.
func (engine *Engine) NewGame() error {
	if err := engine.Write(engine.protocol + "newgame"); err != nil {
		return err
	}

	return engine.Synchronize()
}

// Initialize initializes the engine on startup.
func (engine *Engine) Initialize() error {
	if err := engine.Write(engine.protocol); err != nil {
		return err
	}

	_, err := engine.Await(engine.protocol+"ok", 5*time.Second)
	return err
}

// Synchronize waits for the engine to complete some time consuming task
// and synchronizes the interface with it.
func (engine *Engine) Synchronize() error {
	if err := engine.Write("isready"); err != nil {
		return err
	}

	_, err := engine.Await("readyok", 5*time.Second)
	return err
}

// Kill kills the engine.
func (engine *Engine) Kill() error {
	if err := engine.Write("quit"); err != nil {
		return err
	}

	return engine.Process.Kill()
}

var ErrReadTimeout = errors.New("engine: read i/o timeout")

// Await is a utility function which waits for a particular string from
// the engine with a fixed timeout.
func (engine *Engine) Await(pattern string, timeout time.Duration) (string, error) {
	regex := regexp.MustCompile(pattern)
	timer := time.NewTimer(timeout)

	for {
		select {
		case <-timer.C:
			// timer ran out: wait timeout

			if engine.err != nil {
				return "", engine.err
			}

			return "", ErrReadTimeout

		case line := <-engine.lines:
			if regex.MatchString(line) {
				// line is the expected line
				return line, nil
			}
		}
	}
}

func (engine *Engine) Write(format string, a ...any) error {
	logrus.Debugf("info: ("+engine.config.Name+")< "+format+"\n", a...)

	if _, err := fmt.Fprintf(engine.writer, format+"\n", a...); err != nil {
		return err
	}

	return engine.writer.Flush()
}
