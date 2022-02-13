package multigitter

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/lindell/multi-gitter/internal/scm"
	"github.com/pkg/errors"
	"github.com/whilp/git-urls"
	log "github.com/sirupsen/logrus"

	"github.com/lindell/multi-gitter/internal/multigitter/repocounter"
)

// Cloner contains fields to be able to do the run
type Cloner struct {
	VersionController VersionController

	Arguments     []string

	Output io.Writer

	DryRun          bool
	Concurrent      int
	SkipRepository  []string // A list of repositories that run will skip

	CreateGit func(dir string) Git
}

// Run runs a script for multiple repositories and creates PRs with the changes made
func (r *Runner) Clone(ctx context.Context) error {
	// Fetch all repositories that are are going to be used in the run
	repos, err := r.VersionController.GetRepositories(ctx)
	if err != nil {
		return errors.Wrap(err, "could not fetch repositories")
	}

	repos = filterRepositories(repos, r.SkipRepository)

	if len(repos) == 0 {
		log.Infof("No repositories found. Please make sure the user of the token has the correct access to the repos you want to change.")
		return nil
	}

	// Setting up a "counter" that keeps track of successful and failed runs
	rc := repocounter.NewCounter()
	defer func() {
		if info := rc.Info(); info != "" {
			fmt.Fprint(r.Output, info)
		}
	}()

	log.Infof("Running on %d repositories", len(repos))

	if r.DryRun {
		log.Info("Skipping cloning repos because of dry run")
		return nil
	}

	runInParallel(func(i int) {
		logger := log.WithField("repo", repos[i].FullName())

		defer func() {
			if r := recover(); r != nil {
				log.Error(r)
				rc.AddError(errors.New("run paniced"), repos[i])
			}
		}()

		err := r.cloneSingleRepo(ctx, repos[i])
		if err != nil {
			if err != errAborted {
				logger.Info(err)
			}
			rc.AddError(err, repos[i])

			if log.IsLevelEnabled(log.TraceLevel) {
				if stackTrace := getStackTrace(err); stackTrace != "" {
					log.Trace(stackTrace)
				}
			}

			return
		}

		rc.AddSuccessRepositories(repos[i])
	}, len(repos), r.Concurrent)

	return nil
}

func (r *Runner) cloneSingleRepo(ctx context.Context, repo scm.Repository) (error) {
	if ctx.Err() != nil {
		return errAborted
	}

	log := log.WithField("repo", repo.FullName())
	log.Info("Cloning repo")

	directoryName, err := getCloneDirectory(repo.CloneURL())
	directoryString := fmt.Sprintf("%s/multi-gitter/%s", os.TempDir(), directoryName)
	err = os.Mkdir(directoryString, 100)
	sourceController := r.CreateGit(directoryName)

	err = sourceController.Clone(repo.CloneURL(), repo.DefaultBranch())
	if err != nil {
		return err
	}

	return nil
}

func getCloneDirectory(repoUrl string) (string, error) {
	url, err := giturls.Parse(repoUrl)
	urlPath := url.Path

	urlPath = strings.TrimLeft(url.Path, "/")
	dir := strings.ReplaceAll(urlPath, ".git", "")

	return dir, err
}
