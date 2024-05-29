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

package tournament

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

func NewEngine(config EngineConfig) (*Player, error) {
	var engine Player
	engine.protocol = config.Protocol
	process := exec.Command(config.Cmd, strings.Fields(config.Arg)...)

	engine.config = config

	process.Dir = config.Dir

	stdin, _ := process.StdinPipe()
	stdout, _ := process.StdoutPipe()

	engine.writer = bufio.NewWriter(stdin)
	engine.reader = bufio.NewReader(stdout)
	engine.lines = make(chan string)

	engine.logger = io.Discard

	engine.Cmd = process

	if err := process.Start(); err != nil {
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

			engine.logf("info: (engine)> %s\n", line)
			engine.lines <- line
		}
	}()

	if config.InitStr != "" {
		if err := engine.Write(config.InitStr); err != nil {
			return nil, err
		}
	}

	engine.Initialize()

	return &engine, nil
}

type Player struct {
	config EngineConfig

	*exec.Cmd

	protocol string

	writer *bufio.Writer
	reader *bufio.Reader

	lines chan string

	logger io.Writer
	err    error
}

// conf=NAME		Use an engine with the name NAME from Cute Chess'
// 		engines.json configuration file.
// name=NAME		Set the name to NAME
// cmd=COMMAND		Set the command to COMMAND
// dir=DIR		Set the working directory to DIR
// arg=ARG		Pass ARG to the engine as a command line argument
// initstr=TEXT		Send TEXT to the engine's standard input at startup.
// 		TEXT may contain multiple lines seprated by '\n'.
// stderr=FILE		Redirect standard error output to FILE
// restart=MODE		Set the restart mode to MODE which can be:
// 		'auto': the engine decides whether to restart (default)
// 		'on': the engine is always restarted between games
// 		'off': the engine is never restarted between games
// 		Setting this option does not prevent engines from being
// 		restarted between rounds in a tournament featuring more
// 		than two engines.
// trust			Trust result claims from the engine without validation.
// 		By default all claims are validated.
// proto=PROTOCOL	Set the chess protocol to PROTOCOL, which can be one of:
// 		'xboard': The Xboard/Winboard/CECP protocol
// 		'uci': The Universal Chess Interface
// tc=TIMECONTROL	Set the time control to TIMECONTROL. The format is
// 		moves/time+increment, where 'moves' is the number of
// 		moves per tc, 'time' is time per tc (either seconds or
// 		minutes:seconds), and 'increment' is time increment
// 		per move in seconds.
// 		Infinite time control can be set with 'tc=inf'.
// st=N			Set the time limit for each move to N seconds.
// 		This option can't be used in combination with "tc".
// timemargin=N		Let engines go N milliseconds over the time limit.
// book=FILE		Use FILE (Polyglot book file) as the opening book
// bookdepth=N		Set the maximum book depth (in fullmoves) to N
// whitepov		Invert the engine's scores when it plays black. This
// 		option should be used with engines that always report
// 		scores from white's perspective.
// depth=N		Set the search depth limit to N plies
// nodes=N		Set the node count limit to N nodes
// ponder		Enable pondering if the engine supports it. By default
// 		pondering is disabled.
// option.OPTION=VALUE	Set custom option OPTION to value VALUE

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

// NewGame prepares the engine for a new game of chess.
func (engine *Player) NewGame() error {
	if err := engine.Write(engine.protocol + "newgame"); err != nil {
		return err
	}

	return engine.Synchronize()
}

// Initialize initializes the engine on startup.
func (engine *Player) Initialize() error {
	if err := engine.Write(engine.protocol); err != nil {
		return err
	}

	_, err := engine.Await(engine.protocol+"ok", 5*time.Second)
	return err
}

// Synchronize waits for the engine to complete some time consuming task
// and synchronizes the interface with it.
func (engine *Player) Synchronize() error {
	if err := engine.Write("isready"); err != nil {
		return err
	}

	_, err := engine.Await("readyok", 5*time.Second)
	return err
}

// Kill kills the engine.
func (engine *Player) Kill() error {
	if err := engine.Write("quit"); err != nil {
		return err
	}

	return engine.Process.Kill()
}

var ErrReadTimeout = errors.New("engine: read i/o timeout")

// Await is a utility function which waits for a particular string from
// the engine with a fixed timeout.
func (engine *Player) Await(pattern string, timeout time.Duration) (string, error) {
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

func (engine *Player) Write(format string, a ...any) error {
	engine.logf("info: ("+engine.config.Name+")< "+format+"\n", a...)

	if _, err := fmt.Fprintf(engine.writer, format+"\n", a...); err != nil {
		return err
	}

	return engine.writer.Flush()
}

func (engine *Player) logf(format string, a ...any) {
	fmt.Fprintf(engine.logger, format, a...)
}
