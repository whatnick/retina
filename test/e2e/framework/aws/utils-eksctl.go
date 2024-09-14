package aws

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/fatih/color"
	"github.com/kris-nova/logger"
	lol "github.com/kris-nova/lolgopher"
	"github.com/spf13/cobra"
	"github.com/weaveworks/eksctl/pkg/ctl/cmdutils"
	"github.com/weaveworks/eksctl/pkg/ctl/create"
	"github.com/weaveworks/eksctl/pkg/ctl/delete"
)

func initLogger(level int, colorValue string, logBuffer *bytes.Buffer, dumpLogsValue bool) {
	logger.Layout = "2006-01-02 15:04:05"

	var bitwiseLevel int
	switch level {
	case 4:
		bitwiseLevel = logger.LogDeprecated | logger.LogAlways | logger.LogSuccess | logger.LogCritical | logger.LogWarning | logger.LogInfo | logger.LogDebug
	case 3:
		bitwiseLevel = logger.LogDeprecated | logger.LogAlways | logger.LogSuccess | logger.LogCritical | logger.LogWarning | logger.LogInfo
	case 2:
		bitwiseLevel = logger.LogDeprecated | logger.LogAlways | logger.LogSuccess | logger.LogCritical | logger.LogWarning
	case 1:
		bitwiseLevel = logger.LogDeprecated | logger.LogAlways | logger.LogSuccess | logger.LogCritical
	case 0:
		bitwiseLevel = logger.LogDeprecated | logger.LogAlways | logger.LogSuccess
	default:
		bitwiseLevel = logger.LogDeprecated | logger.LogEverything
	}
	logger.BitwiseLevel = bitwiseLevel

	if dumpLogsValue {
		switch colorValue {
		case "fabulous":
			logger.Writer = io.MultiWriter(lol.NewLolWriter(), logBuffer)
		case "true":
			logger.Writer = io.MultiWriter(color.Output, logBuffer)
		default:
			logger.Writer = io.MultiWriter(os.Stdout, logBuffer)
		}

	} else {
		switch colorValue {
		case "fabulous":
			logger.Writer = lol.NewLolWriter()
		case "true":
			logger.Writer = color.Output
		default:
			logger.Writer = os.Stdout
		}
	}

	logger.Line = func(prefix, format string, a ...interface{}) string {
		if !strings.Contains(format, "\n") {
			format = fmt.Sprintf("%s%s", format, "\n")
		}
		now := time.Now()
		fNow := now.Format(logger.Layout)
		var colorize func(format string, a ...interface{}) string
		var icon string
		switch prefix {
		case logger.PreAlways:
			icon = "✿"
			colorize = color.GreenString
		case logger.PreCritical:
			icon = "✖"
			colorize = color.RedString
		case logger.PreInfo:
			icon = "ℹ"
			colorize = color.CyanString
		case logger.PreDebug:
			icon = "▶"
			colorize = color.GreenString
		case logger.PreSuccess:
			icon = "✔"
			colorize = color.CyanString
		case logger.PreWarning:
			icon = "!"
			colorize = color.GreenString
		default:
			icon = "ℹ"
			colorize = color.CyanString
		}

		out := fmt.Sprintf(format, a...)
		out = fmt.Sprintf("%s [%s]  %s", fNow, icon, out)
		if colorValue == "true" {
			out = colorize(out)
		}

		return out
	}
}

func checkCommand(rootCmd *cobra.Command) {
	for _, cmd := range rootCmd.Commands() {
		// just a precaution as the verb command didn't have runE
		if cmd.RunE != nil {
			continue
		}
		cmd.RunE = func(c *cobra.Command, args []string) error {
			var e error
			if len(args) == 0 {
				e = fmt.Errorf("please provide a valid resource for \"%s\"", c.Name())
			} else {
				e = fmt.Errorf("unknown resource type \"%s\"", args[0])
			}
			fmt.Printf("Error: %s\n\n", e.Error())

			if err := c.Help(); err != nil {
				logger.Debug("ignoring cobra error %q", err.Error())
			}
			return e
		}
	}
}

func templateCluster(templateFile, outputFile string, c *CreateCluster) error {
	// Read the template file
	tmpl, err := template.ParseFiles(templateFile)
	if err != nil {
		return err
	}

	// Create the output file
	outFile, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer outFile.Close()

	// Execute the template with the data
	err = tmpl.Execute(outFile, c)
	if err != nil {
		return err
	}

	return nil
}

func CreateEKSCtlCmd() *cobra.Command {
	// Create the ekctl root cmd to execute
	rootCmd := &cobra.Command{
		Use:   "eksctl [command]",
		Short: "The official CLI for Amazon EKS",
		Run: func(c *cobra.Command, _ []string) {
			if err := c.Help(); err != nil {
				logger.Debug("ignoring cobra error %q", err.Error())
			}
		},
		SilenceUsage: true,
	}

	loggerLevel := rootCmd.PersistentFlags().IntP("verbose", "v", 3, "set log level, use 0 to silence, 4 for debugging and 5 for debugging with AWS debug logging")
	colorValue := rootCmd.PersistentFlags().StringP("color", "C", "true", "toggle colorized logs (valid options: true, false, fabulous)")
	dumpLogsValue := rootCmd.PersistentFlags().BoolP("dumpLogs", "d", false, "dump logs to disk on failure if set to true")

	logBuffer := new(bytes.Buffer)

	cobra.OnInitialize(func() {
		initLogger(*loggerLevel, *colorValue, logBuffer, *dumpLogsValue)
	})

	flagGrouping := cmdutils.NewGrouping()
	createCmd := create.Command(flagGrouping)
	deleteCmd := delete.Command(flagGrouping)
	rootCmd.AddCommand(createCmd)
	rootCmd.AddCommand(deleteCmd)

	checkCommand(rootCmd)

	return rootCmd
}
