package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/lindell/multi-gitter/internal/multigitter"
	"github.com/spf13/cobra"
)

const cloneHelp = `
This command will clone down multiple repositories. The output is the base clone directory.

The environment variable REPOSITORY will be set to the name of the repository currently being executed by the script.
`

// CloneCmd clones multiple repositories
func CloneCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "clone",
		Short:   "Clones multiple repositories",
		Long:    cloneHelp,
		Args:    cobra.ExactArgs(0),
		PreRunE: logFlagInit,
		RunE:    clone,
	}

	cmd.Flags().IntP("concurrent", "C", 1, "The maximum number of concurrent runs.")
	cmd.Flags().StringSliceP("skip-repo", "s", nil, "Skip changes on specified repositories, the name is including the owner of repository in the format \"ownerName/repoName\".")
	configureGit(cmd)
	configurePlatform(cmd)
	configureRunPlatform(cmd, false)
	configureLogging(cmd, "-")
	configureConfig(cmd)
	cmd.Flags().AddFlagSet(outputFlag())

	return cmd
}

func clone(cmd *cobra.Command, args []string) error {
	flag := cmd.Flags()

	concurrent, _ := flag.GetInt("concurrent")
	skipRepository, _ := flag.GetStringSlice("skip-repo")
	strOutput, _ := flag.GetString("output")

	if concurrent < 1 {
		return errors.New("concurrent runs can't be less than one")
	}

	output, err := fileOutput(strOutput, os.Stdout)
	if err != nil {
		return err
	}

	vc, err := getVersionController(flag, true)
	if err != nil {
		return err
	}

	gitCreator, err := getGitCreator(flag)
	if err != nil {
		return err
	}

	// Set up signal listening to cancel the context and let started runs finish gracefully
	ctx, cancel := context.WithCancel(context.Background())
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("Finishing up ongoing runs. Press CTRL+C again to abort now.")
		cancel()
		<-c
		os.Exit(1)
	}()

	runner := &multigitter.Runner{
		Output: output,

		VersionController: vc,

		SkipRepository:   skipRepository,
		Concurrent: concurrent,

		CreateGit: gitCreator,
	}

	err = runner.Clone(ctx)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	return nil
}
