package main

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/ldez/grignotin/version"
)

const baseDownloadURL = "https://dl.google.com/go/"

type localInfo struct {
	Version string
	Path    string
}

type sem struct {
	client *http.Client
	goRoot string
	debug  bool
}

func (s *sem) getGo(root, targetedVersion string) error {
	info, err := s.getReleaseInfo(targetedVersion)
	if err != nil {
		return err
	}

	dest := filepath.Join(root, strings.TrimPrefix(info.Version, "go"))

	current, err := s.extractVersionFromGoRoot()
	if err != nil {
		return err
	}

	if info.Version == "go"+current.Version {
		fmt.Printf("Nothing to do: %s already installed.\n", info.Version)
		return nil
	}

	locals, err := s.getLocalVersions(root)
	if err != nil {
		return err
	}

	// removes current local version
	current, err = removeCurrent(locals, current)
	if err != nil {
		return err
	}

	err = createSymlink(dest, current)
	if err != nil {
		return err
	}

	for _, local := range locals {
		// the version already exits locally
		if info.Version == "go"+local.Version {
			fmt.Printf("[local] go%s has been replaced by %s.\n", current.Version, info.Version)

			return nil
		}
	}

	resp, err := s.client.Get(baseDownloadURL + info.Filename)
	if err != nil {
		return err
	}

	defer func() { _ = resp.Body.Close() }()

	err = s.extract(dest, resp.Body)
	if err != nil {
		return err
	}

	fmt.Printf("[remote] go%s has been replaced by %s.\n", current.Version, info.Version)

	return nil
}

func (s *sem) getReleaseInfo(v string) (*version.File, error) {
	stableReleases, err := version.GetReleases(false)
	if err != nil {
		return nil, err
	}

	info := s.findReleaseInfo(stableReleases, v)
	if info == nil {
		log.Printf("no stable release for %s, trying to find an unstable release", v)
		info, err = findLatestUnstable(v)
		if err != nil {
			return nil, fmt.Errorf("unsupported version: %w", err)
		}
	}

	s.logDebugf("find release: %+v", info)

	return info, nil
}

func (s *sem) findReleaseInfo(releases []version.Release, v string) *version.File {
	for _, release := range releases {
		if v != "" && strings.HasPrefix(release.Version, v) {
			f := findFile(release)
			if f != nil {
				return f
			}
		}
	}

	return nil
}

func (s *sem) extractVersionFromGoRoot() (*localInfo, error) {
	expr := regexp.MustCompile(`\d\.\d+(?:\.\d+)?`)

	goRoot := os.Getenv(s.goRoot)

	s.logDebugf("%s=%s", s.goRoot, goRoot)

	subMatch := expr.FindStringSubmatch(goRoot)

	if len(subMatch) != 1 {
		return nil, fmt.Errorf("unable to extract version from %s: %s", s.goRoot, goRoot)
	}

	return &localInfo{Version: subMatch[0], Path: goRoot}, nil
}

func (s *sem) getLocalVersions(dir string) (map[string]localInfo, error) {
	glob, err := filepath.Glob(filepath.Join(dir, "*"))
	if err != nil {
		s.logDebugf("glob error: %v", err)
		return nil, err
	}

	result := map[string]localInfo{}

	for _, folder := range glob {
		vPath := filepath.Base(folder)
		subMatch := regexp.MustCompile(`(\d\.\d+)(?:.\d+)?`).FindStringSubmatch(vPath)

		if len(subMatch) != 2 {
			s.logDebugf("subMatch: %v %s %s", subMatch, folder, vPath)
			continue
		}

		result["go"+subMatch[1]] = localInfo{Version: vPath, Path: folder}
	}

	return result, nil
}

func (s *sem) extract(dest string, stream io.Reader) error {
	err := os.MkdirAll(dest, 0o775)
	if err != nil {
		return err
	}

	s.logDebugf("Extracting the Go archive to %s", dest)

	uncompressed, err := gzip.NewReader(stream)
	if err != nil {
		return err
	}

	tr := tar.NewReader(uncompressed)

	for {
		header, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			return err
		}

		abs := filepath.Join(dest, filepath.FromSlash(header.Name))

		mode := header.FileInfo().Mode()

		switch {
		case mode.IsRegular():
			if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
				return err
			}

			out, err := os.OpenFile(abs, os.O_RDWR|os.O_CREATE|os.O_TRUNC, mode.Perm())
			if err != nil {
				return err
			}

			n, err := io.Copy(out, tr)
			if err != nil {
				return err
			}

			if err = out.Close(); err != nil {
				return err
			}

			if n != header.Size {
				return fmt.Errorf("TODO")
			}

		case mode.IsDir():
			if err := os.Mkdir(abs, 0o755); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unknown type: %s in %v", header.Name, mode)
		}
	}

	return nil
}

func (s *sem) logDebugf(format string, v ...interface{}) {
	if !s.debug {
		return
	}

	log.Printf(format, v...)
}

// removeCurrent removes current local version.
func removeCurrent(locals map[string]localInfo, current *localInfo) (*localInfo, error) {
	for _, local := range locals {
		local := local

		if local.Version == current.Version {
			err := os.RemoveAll(local.Path)
			if err != nil {
				return nil, err
			}

			return &local, nil
		}
	}

	return nil, fmt.Errorf("unable to find %s", current.Path)
}

func createSymlink(dest string, local *localInfo) error {
	err := os.MkdirAll(dest, 0o775)
	if err != nil {
		return err
	}

	err = os.RemoveAll(local.Path)
	if err != nil {
		return err
	}

	abs, err := filepath.Abs(dest)
	if err != nil {
		return err
	}

	return os.Symlink(abs, local.Path)
}

func findLatestUnstable(v string) (*version.File, error) {
	releases, err := version.GetReleases(true)
	if err != nil {
		return nil, err
	}

	exp, err := regexp.Compile(fmt.Sprintf(`%s(rc|beta).+`, v))
	if err != nil {
		return nil, err
	}

	var selected []version.Release
	for _, release := range releases {
		if !release.Stable && exp.MatchString(release.Version) {
			selected = append(selected, release)
		}
	}

	sort.Slice(selected, func(i, j int) bool {
		return selected[i].Version > selected[j].Version
	})

	if len(selected) > 0 {
		file := findFile(selected[0])
		if file == nil {
			return nil, fmt.Errorf("file not found: %s", v)
		}

		return file, nil
	}

	return nil, fmt.Errorf("version not found: %s", v)
}

func findFile(release version.Release) *version.File {
	for _, file := range release.Files {
		if file.OS == "linux" && file.Arch == "amd64" {
			return &file
		}
	}

	return nil
}
