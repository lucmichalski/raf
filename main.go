package main

import (
	"errors"
	"fmt"
	"log"
	"os"

	cli "github.com/urfave/cli/v2"
)

type output struct {
	Raw            string
	VarCount       int
	CustomVarCount int
	Tokens         TokenStream
}

// Opts contains the global settings for the library and it is passed to nearly
// all methods.
type Opts struct {
	DryRun  bool
	Verbose bool
}

func main() {
	cli.VersionFlag = &cli.BoolFlag{
		Name:    "version",
		Aliases: []string{"V"},
		Usage:   "print raf version",
	}
	app := &cli.App{
		Name:        "raf",
		Usage:       "raf -p \"title=Video\\ \\d+\\ \\-\\ ([A-Za-z0-9\\ ]+)_\" -d -o 'UnionStudio - $cnt - $title.mkv' *",
		Description: cliDescription,
		Version:     "v0.1",
		Flags: []cli.Flag{
			&cli.StringSliceFlag{
				Name:    "prop",
				Aliases: []string{"p"},
				//Usage:   "-p \"title=Video\\ \\d+\\ \\-\\ ([A-Za-z0-9\\ ]+)_\"",
				Usage: propFlagDescription,
			},
			&cli.StringFlag{
				Name:     "output",
				Aliases:  []string{"o"},
				Usage:    outputFlagDescription,
				Required: true,
			},
			&cli.BoolFlag{
				Name:    "dryrun",
				Aliases: []string{"d"},
				Usage:   dryRunFlagDescription,
			},
			&cli.BoolFlag{
				Name:    "verbose",
				Aliases: []string{"v"},
				Usage:   "Prints verbose output",
			},
		},
		Action: rename,
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func rename(c *cli.Context) error {
	matches, err := validateMatcher(c)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	props, err := validateProps(c)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	out, err := validateOutput(c)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if len(props) < out.CustomVarCount {
		fmt.Printf("The command declared %d properties but uses %d in the output formatter\n", len(props), out.CustomVarCount)
		os.Exit(1)
	}
	opts := Opts{
		DryRun:  c.Bool("dryrun"),
		Verbose: c.Bool("verbose"),
	}
	err = RenameAllFiles(props, out.Tokens, matches, opts)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return nil
}

func validateMatcher(c *cli.Context) ([]string, error) {
	files := c.Args().Slice()
	if files == nil || len(files) == 0 {
		return nil, errors.New("No input files")
	}

	return files, nil
}

func validateProps(c *cli.Context) ([]Prop, error) {
	args := c.StringSlice("prop")
	props := make([]Prop, len(args))

	for idx, v := range args {
		prop, err := ParseProp(v)
		if err != nil {
			return nil, err
		}
		props[idx] = prop
	}

	return props, nil
}

func validateOutput(c *cli.Context) (*output, error) {
	rawOutput := c.String("output")
	if rawOutput == "" {
		return nil, errors.New("Output formatter must be a valid string and cannot be empty")
	}

	tokens, err := ParseOutput(rawOutput)
	if err != nil {
		return nil, err
	}

	varCount := 0
	customVarCount := 0
	for _, t := range tokens {
		if t.Type != TokenTypeProperty {
			continue
		}
		varCount++
		if _, ok := ReservedVarNames[t.Value]; !ok {
			customVarCount++
		}
	}

	return &output{
		Raw:            rawOutput,
		VarCount:       varCount,
		CustomVarCount: customVarCount,
		Tokens:         tokens,
	}, nil
}
