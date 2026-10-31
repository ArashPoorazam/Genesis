package experiment

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ArashPoorazam/Genesis/internal/simulation"
)

type Preset struct {
	Name   string            `json:"name"`
	Config simulation.Config `json:"config"`
}

type Store struct {
	Directory string
}

func DefaultStore() (Store, error) {
	directory, err := os.UserConfigDir()
	if err != nil {
		return Store{}, err
	}
	return Store{Directory: filepath.Join(directory, "genesis", "presets")}, nil
}

func validName(name string) error {
	if name != strings.TrimSpace(name) || len([]rune(name)) < 1 || len([]rune(name)) > 64 || name == "." || name == ".." || strings.ContainsAny(name, "/\\\x00<>:\"|?*\n\r\t") {
		return errors.New("use a name of 1–64 characters without path separators or special filename characters")
	}
	return nil
}

func (store Store) path(name string) (string, error) {
	if err := validName(name); err != nil {
		return "", err
	}
	return filepath.Join(store.Directory, name+".json"), nil
}

func (store Store) List() ([]string, error) {
	entries, err := os.ReadDir(store.Directory)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	names := []string{}
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
			names = append(names, strings.TrimSuffix(entry.Name(), ".json"))
		}
	}
	sort.Strings(names)
	return names, nil
}

func ReadPreset(path string) (Preset, error) {
	file, err := os.Open(path)
	if err != nil {
		return Preset{}, err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return Preset{}, err
	}
	if info.Size() > 1<<20 {
		return Preset{}, errors.New("preset exceeds 1 MB")
	}

	decoder := json.NewDecoder(io.LimitReader(file, 1<<20))
	decoder.DisallowUnknownFields()
	var preset Preset
	if err = decoder.Decode(&preset); err != nil {
		return preset, fmt.Errorf("invalid preset: %w", err)
	}

	var extra any
	if err = decoder.Decode(&extra); err != io.EOF {
		return preset, errors.New("unexpected data after preset")
	}
	if err = validName(preset.Name); err != nil {
		return preset, err
	}
	if err = preset.Config.Validate(); err != nil {
		return preset, err
	}

	return preset, nil
}

// WritePreset replaces the destination only after a complete, synced temporary write.
func WritePreset(path string, preset Preset) error {
	if err := validName(preset.Name); err != nil {
		return err
	}
	if err := preset.Config.Validate(); err != nil {
		return err
	}

	data, err := json.MarshalIndent(preset, "", "  ")
	if err != nil {
		return err
	}

	directory := filepath.Dir(path)
	if err = os.MkdirAll(directory, 0700); err != nil {
		return err
	}

	file, err := os.CreateTemp(directory, ".genesis-*.tmp")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())

	if _, err = file.Write(append(data, '\n')); err != nil {
		file.Close()
		return err
	}
	if err = file.Sync(); err != nil {
		file.Close()
		return err
	}
	if err = file.Close(); err != nil {
		return err
	}
	return os.Rename(file.Name(), path)
}

func (store Store) Save(name string, config simulation.Config) error {
	path, err := store.path(name)
	if err != nil {
		return err
	}
	return WritePreset(path, Preset{Name: name, Config: config})
}

func (store Store) Load(name string) (Preset, error) {
	path, err := store.path(name)
	if err != nil {
		return Preset{}, err
	}
	return ReadPreset(path)
}

func (store Store) Exists(name string) bool {
	path, err := store.path(name)
	if err != nil {
		return false
	}
	_, err = os.Stat(path)
	return err == nil
}

func (store Store) Delete(name string) error {
	path, err := store.path(name)
	if err != nil {
		return err
	}
	return os.Remove(path)
}

func (store Store) Rename(old, name string) error {
	if old == name {
		return nil
	}
	if store.Exists(name) {
		return errors.New("a preset with that name already exists")
	}
	preset, err := store.Load(old)
	if err != nil {
		return err
	}
	if err = store.Save(name, preset.Config); err != nil {
		return err
	}
	if err = store.Delete(old); err != nil {
		return fmt.Errorf("new copy saved, but old preset could not be removed: %w", err)
	}
	return nil
}
