package file

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/mbobakov/khrushchevka/internal"
	"github.com/mbobakov/khrushchevka/internal/lights"
	"github.com/mbobakov/khrushchevka/internal/snapshot"
	"github.com/spf13/afero"
)

type Options struct {
	Path string `long:"path" env:"PATH" default:"./snapshot.json" description:"path to snapshot file"`
}

type JSON struct {
	mu      sync.Mutex
	fs      afero.Fs
	path    string
	lights  lights.ControllerI
	mapping [][]internal.Light
}

func New(opts Options, fs afero.Fs, l lights.ControllerI, mapping [][]internal.Light) *JSON {
	return &JSON{
		fs:      fs,
		path:    opts.Path,
		mapping: mapping,
		lights:  l,
	}
}

func (j *JSON) Snapshot() error {
	j.mu.Lock()
	defer j.mu.Unlock()

	f, err := j.fs.OpenFile(j.path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("couldn't open file '%s': %w", j.path, err)
	}
	defer f.Close()

	state := []snapshot.LightDTO{}

	for _, lvl := range j.mapping {
		for _, light := range lvl {
			if light.Addr.Pin == "" {
				continue
			}
			isOn, err := j.lights.IsOn(light.Addr)
			if err != nil {
				return fmt.Errorf("couldn't get light state for '%v': %w", light.Addr, err)
			}

			state = append(state, snapshot.LightDTO{Board: light.Addr.Board, Pin: light.Addr.Pin, IsOn: isOn})
		}
	}

	jsBuf, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("couldn't marshal state: %w", err)
	}

	_, err = f.Write(jsBuf)
	if err != nil {
		return fmt.Errorf("couldn't write to file: %w", err)
	}

	_, err = f.WriteString("\n")
	if err != nil {
		return fmt.Errorf("couldn't write endline to file: %w", err)
	}

	return nil
}
