package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func Execute() error {
	return newRootCommand().Execute()
}

func newRootCommand() *cobra.Command {
	options := defaultGetOptions()

	rootCmd := &cobra.Command{
		Use:           "crawldown [flags] <url>",
		Short:         "Download website content and convert it to Markdown",
		Long:          "CrawlDown downloads a single page or crawls a website and saves the content as Markdown files.",
		Version:       buildVersion(),
		SilenceUsage:  true,
		SilenceErrors: true,
		Args: func(cmd *cobra.Command, args []string) error {
			return validateGetInvocation(options, args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGet(options, args)
		},
	}

	rootCmd.SetVersionTemplate("{{printf \"%s\\n\" .Version}}")
	bindGetFlags(rootCmd, options)
	rootCmd.AddCommand(newGetCommand(), newAddSkillCommand())

	return rootCmd
}

func buildVersion() string {
	parts := []string{version}

	if commit != "" && commit != "none" {
		parts = append(parts, commit)
	}

	if date != "" && date != "unknown" {
		parts = append(parts, date)
	}

	return strings.Join(parts, " ")
}

func bindGetFlags(cmd *cobra.Command, options *getOptions) {
	flags := cmd.Flags()
	flags.StringVarP(&options.outputDir, "output", "o", "", "Directory where Markdown files will be saved")
	flags.StringVarP(&options.singleURL, "single", "s", "", "Download a single page instead of crawling from the positional URL")
	flags.IntVarP(&options.maxDepth, "depth", "d", 2, "Maximum crawl depth")
	flags.StringSliceVarP(&options.excludedPaths, "exclude", "e", nil, "URL path prefixes to exclude from crawling")
	flags.IntVarP(&options.requestTimeout, "timeout", "t", 60, "Request timeout in seconds")
	flags.IntVar(&options.requestDelay, "delay", 1, "Delay between requests in seconds")
	flags.BoolVar(&options.ignoreRobotsTxt, "ignore-robots-txt", false, "Ignore robots.txt while crawling")
	flags.BoolVar(&options.followExternalLinks, "follow-external-links", false, "Allow following external links")
	flags.StringVar(&options.userAgent, "user-agent", "CrawlDown/1.0", "HTTP user agent used for requests")
}

func newGetCommand() *cobra.Command {
	options := defaultGetOptions()

	getCmd := &cobra.Command{
		Use:           "get [flags] <url>",
		Short:         "Download website content as Markdown",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args: func(cmd *cobra.Command, args []string) error {
			return validateGetInvocation(options, args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGet(options, args)
		},
	}

	bindGetFlags(getCmd, options)

	return getCmd
}

func validateGetInvocation(options *getOptions, args []string) error {
	if options.outputDir == "" {
		return fmt.Errorf("required flag \"output\" not set")
	}

	if options.singleURL == "" {
		switch len(args) {
		case 0:
			return fmt.Errorf("requires a URL argument or --single")
		case 1:
		default:
			return fmt.Errorf("accepts at most 1 argument, received %d", len(args))
		}
		return nil
	}

	if len(args) > 1 {
		return fmt.Errorf("accepts at most 1 argument when --single is set, received %d", len(args))
	}

	return nil
}
