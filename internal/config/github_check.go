package config

import (
	"context"
	"fmt"
	"log"

	"github.com/Masterminds/semver"
	"github.com/fatih/color"

	"github.com/babarot/afx/internal/errors"
	"github.com/babarot/afx/internal/github"
	"github.com/babarot/afx/internal/runner"
)

func (c GitHub) Check(ctx context.Context, status chan<- runner.Status) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	select {
	case <-ctx.Done():
		log.Println("[DEBUG] canceled")
		return nil
	default:
		// go next
	}

	switch {
	case c.Release == nil:
		// TODO: Check git commit
		status <- runner.Status{Name: c.GetName(), Done: true, Err: false, Message: "(github)", NoColor: true}
		return nil
	case c.Release != nil:
		report, err := c.checkUpdates(ctx)
		if err != nil {
			err = errors.Wrapf(err, "%s: failed to check release version", c.Name)
		}
		status <- runner.Status{Name: c.GetName(), Done: true, Err: err != nil, Message: report.message}
		return err
	}

	status <- runner.Status{Name: c.GetName(), Done: true, Err: false}
	return nil
}

type report struct {
	message string
}

func (c GitHub) checkUpdates(ctx context.Context) (report, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	red := color.New(color.FgRed).SprintfFunc()
	yellow := color.New(color.FgYellow).SprintfFunc()

	tag := c.Release.Tag
	switch tag {
	case "latest", "stable", "nightly":
		return report{message: tag}, nil
	case "":
		return report{message: "(tag not set)"}, nil
	}

	release, err := github.NewRelease(
		ctx, c.Owner, c.Repo, "latest",
		github.WithWorkdir(c.GetHome()),
	)
	if err != nil {
		return report{
			message: fmt.Sprintf("%s %s", red("error!"), err),
		}, err
	}

	current, err := semver.NewVersion(tag)
	if err != nil {
		return report{}, nil
	}

	next, err := semver.NewVersion(release.Tag)
	if err != nil {
		return report{}, nil
	}

	switch current.Compare(next) {
	case -1:
		return report{
			message: fmt.Sprintf("%s v%s -> v%s",
				yellow("new!"), current, next),
		}, nil
	case 0:
		return report{message: "up-to-date"}, nil
	default:
		return report{}, errors.New("invalid version comparison")
	}
}
