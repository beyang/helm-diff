package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/databus23/helm-diff/diff"
	"github.com/databus23/helm-diff/manifest"
)

type diff2Cmd struct {
	path0 string
	path1 string
}

func (d *diff2Cmd) run() error {
	specs0, err := generateChart(d.path0)
	if err != nil {
		return err
	}
	specs1, err := generateChart(d.path1)
	if err != nil {
		return err
	}

	diff.DiffManifests(specs0, specs1, nil, os.Stdout)
	return nil
}

func generateChart(chartPath string) (map[string]*manifest.MappingResult, error) {
	cmd := exec.Command("helm", "template", ".")
	cmd.Dir = chartPath
	out, err := cmd.Output()
	if err != nil {
		fmt.Fprintln(os.Stderr, string(err.(*exec.ExitError).Stderr))
		return nil, err
	}
	return manifest.Parse(string(out)), nil
}
