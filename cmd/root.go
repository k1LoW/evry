// Copyright © 2019 Ken'ichiro Oyama <k1lowxb@gmail.com>
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
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/k1LoW/evry/splitter"
	"github.com/k1LoW/evry/version"
	"github.com/spf13/cobra"
)

var (
	line    int
	sec     int
	command string
	timeout int
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use: "[COMMAND] | evry [-l N or -s N] -c [COMMAND]",
	Example: `  Count number of requests every 10 seconds

    tail -f access.log | evry -s 10 -c 'wc -l'`,
	Short: "evry split STDIN stream and execute specified command every N lines/seconds",
	Long:  `evry split STDIN stream and execute specified command every N lines/seconds.`,
	Args: func(cmd *cobra.Command, args []string) error {
		// `--version` option
		versionVal, err := cmd.Flags().GetBool("version")
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "%s\n", err)
			os.Exit(1)
		}
		if versionVal {
			fmt.Println(version.Version)
			os.Exit(0)
		}
		fi, err := os.Stdin.Stat()
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "%s\n", err)
			os.Exit(1)
		}
		if (fi.Mode() & os.ModeCharDevice) != 0 {
			return errors.New("evry need STDIN. Please use pipe")
		}
		if (line == 0 && sec == 0) || (line > 0) && (sec > 0) {
			return errors.New("evry need `--line` OR `--sec` flag")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		var s splitter.Splitter
		var err error
		var c []string

		if len(args) > 0 {
			c = args
		} else {
			c = []string{"sh", "-c", command}
		}

		if line > 0 {
			s, err = splitter.NewLineSplitter(ctx, line, c, timeout)
		} else if sec > 0 {
			s, err = splitter.NewSecSplitter(ctx, sec, c, timeout)
		}

		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "%s\n", err)
			os.Exit(1)
		}

		go s.Start()
		defer s.Stop()

		r := bufio.NewReader(os.Stdin)
		for {
			b, err := r.ReadBytes('\n')
			if err == io.EOF {
				break
			} else if err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "%s\n", err)
				os.Exit(1)
			}
			s.In(b)
		}

		s.Close()
		select {
		case <-s.Done():
			break
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().IntVarP(&line, "line", "l", 0, "split stream every N lines")
	rootCmd.Flags().IntVarP(&sec, "sec", "s", 0, "split stream every N seconds")
	rootCmd.Flags().StringVarP(&command, "command", "c", "cat", "command to be executed")
	rootCmd.Flags().IntVarP(&timeout, "timeout", "", 600, "command timeout")
	rootCmd.Flags().BoolP("help", "h", false, "help for evry")
	rootCmd.Flags().BoolP("version", "v", false, "version for evry")
}
