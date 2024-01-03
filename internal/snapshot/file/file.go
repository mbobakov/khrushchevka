package file

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/mbobakov/khrushchevka/internal"
	"github.com/mbobakov/khrushchevka/internal/lights"
	"github.com/spf13/afero"
)

type Options struct {
	Path string `long:"path" env:"PATH" default:"./snapshot.json" description:"path to snapshot file"`
}

type dto struct {
	Board uint8  `json:"board"`
	Pin   string `json:"pin"`
	IsOn  bool   `json:"is_on"`
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

	state := []dto{}

	for _, lvl := range j.mapping {
		for _, light := range lvl {
			isOn, err := j.lights.IsOn(light.Addr)
			if err != nil {
				return fmt.Errorf("couldn't get light state for '%v': %w", light.Addr, err)
			}

			state = append(state, dto{Board: light.Addr.Board, Pin: light.Addr.Pin, IsOn: isOn})
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

func (j *JSON) Replay(ctx context.Context, filpath string, delay time.Duration) error {
	j.mu.Lock()
	defer j.mu.Unlock()

	// Open the file
	file, err := j.fs.Open(filpath)
	if err != nil {
		return fmt.Errorf("couldn't open file '%s': %w", filpath, err)
	}
	defer file.Close()

	// Create a scanner to read the file line by line
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if ctx.Err() != nil {
			return nil
		}

		data := []dto{}
		err := json.Unmarshal(scanner.Bytes(), &data)
		if err != nil {
			return fmt.Errorf("couldn't unmarshal data: %w", err)
		}

		for _, d := range data {
			err := j.lights.Set(internal.LightAddress{Board: d.Board, Pin: d.Pin}, d.IsOn)
			if err != nil {
				return fmt.Errorf("couldn't set light '%v': %w", d, err)
			}
		}

		time.Sleep(delay)

		for _, d := range data {
			err := j.lights.Set(internal.LightAddress{Board: d.Board, Pin: d.Pin}, false)
			if err != nil {
				return fmt.Errorf("couldn't set light '%v': %w", d, err)
			}
		}

	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("couldn't read file: %w", err)
	}

	return nil

}
