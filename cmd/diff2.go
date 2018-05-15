package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	yaml "gopkg.in/yaml.v2"
	"k8s.io/helm/pkg/strvals"

	"github.com/databus23/helm-diff/diff"
	"github.com/databus23/helm-diff/manifest"
)

type diff2Cmd struct {
	path0      string
	path1      string
	valueFiles valueFiles
	values     []string
}

func (d *diff2Cmd) run() error {
	for i, vf := range d.valueFiles {
		var err error
		d.valueFiles[i], err = filepath.Abs(vf)
		if err != nil {
			return err
		}
	}

	specs0, err := d.generateChart(d.path0)
	if err != nil {
		return err
	}
	specs1, err := d.generateChart(d.path1)
	if err != nil {
		return err
	}

	diff.DiffManifests(specs0, specs1, nil, os.Stdout)
	return nil
}

func (d *diff2Cmd) generateChart(chartPath string) (map[string]*manifest.MappingResult, error) {
	args := []string{"template", "."}
	for _, vf := range d.valueFiles {
		args = append(args, "-f", vf)
	}
	for _, v := range d.values {
		args = append(args, "--set", v)
	}

	cmd := exec.Command("helm", args...)
	cmd.Dir = chartPath
	out, err := cmd.Output()
	if err != nil {
		fmt.Fprintln(os.Stderr, string(err.(*exec.ExitError).Stderr))
		return nil, err
	}
	return manifest.Parse(string(out)), nil
}

func (d *diff2Cmd) vals() ([]byte, error) {
	base := map[string]interface{}{}

	// User specified a values files via -f/--values
	for _, filePath := range d.valueFiles {
		currentMap := map[string]interface{}{}
		bytes, err := ioutil.ReadFile(filePath)
		if err != nil {
			return []byte{}, err
		}

		if err := yaml.Unmarshal(bytes, &currentMap); err != nil {
			return []byte{}, fmt.Errorf("failed to parse %s: %s", filePath, err)
		}
		// Merge with the previous map
		base = mergeValues(base, currentMap)
	}

	// User specified a value via --set
	for _, value := range d.values {
		if err := strvals.ParseInto(value, base); err != nil {
			return []byte{}, fmt.Errorf("failed parsing --set data: %s", err)
		}
	}

	return yaml.Marshal(base)
}
