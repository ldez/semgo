package main

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
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
	debug  bool
}

func (s *sem) getGo(root string, targetedVersion string) error {
	info, err := s.getReleaseInfo(targetedVersion)
	if err != nil {
		return err
	}

	dest := filepath.Join(root, strings.TrimPrefix(info.Version, "go"))

	locals, err := s.getLocalVersions(root + "/*")
	if err != nil {
		return err
	}

	local, err := s.findNearestLocalVersion(info, locals)
	if err != nil {
		return err
	}

	if local == nil {
		return nil
	}

	err = createSymlink(dest, local)
	if err != nil {
		return err
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

	fmt.Printf("%s has been replaced by %s.\n", local.Version, info.Version)

	return nil
}

func (s *sem) getReleaseInfo(v string) (*version.File, error) {
	releases, err := version.GetReleases(false)
	if err != nil {
		return nil, err
	}

	info := s.findReleaseInfo(releases, v)
	if info == nil {
		return nil, fmt.Errorf("unsupported version: %s", v)
	}

	if s.debug {
		log.Printf("find release: %+v", info)
	}

	return info, nil
}

func (s *sem) findReleaseInfo(releases []version.Release, v string) *version.File {
	for _, release := range releases {
		if strings.HasPrefix(release.Version, v) {
			for _, file := range release.Files {
				if file.OS == "linux" && file.Arch == "amd64" {
					return &file
				}
			}
		}
	}

	return nil
}

func (s *sem) getLocalVersions(dir string) (map[string]localInfo, error) {
	glob, err := filepath.Glob(dir)
	if err != nil {
		return nil, err
	}

	result := map[string]localInfo{}

	for _, s := range glob {
		vPath := filepath.Base(s)
		subMatch := regexp.MustCompile(`(\d\.\d+)(?:.\d+)?`).FindStringSubmatch(vPath)

		if len(subMatch) != 2 {
			continue
		}

		result["go"+subMatch[1]] = localInfo{Version: vPath, Path: s}
	}

	return result, nil
}

func (s *sem) findNearestLocalVersion(info *version.File, locals map[string]localInfo) (*localInfo, error) {
	subMatch := regexp.MustCompile(`(go\d\.\d+)(?:.\d+)?`).FindStringSubmatch(info.Version)
	if len(subMatch) != 2 {
		return nil, fmt.Errorf("invalid version: %s", info.Version)
	}

	truncVersion := subMatch[1]

	// if a version (goX.YZ) already exist
	if local, ok := locals[truncVersion]; ok {
		if info.Version == "go"+local.Version {
			return nil, nil
		}

		return &local, nil
	}

	parts := strings.Split(truncVersion, ".")

	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, err
	}

	for i := minor; i > 0; i-- {
		local, ok := locals[fmt.Sprintf("%s.%d", parts[0], i)]
		if ok {
			if s.debug {
				log.Printf("find nearest local version: %+v", local)
			}

			return &local, nil
		}
	}

	return nil, fmt.Errorf("unable to find the nearest version of %s", info.Version)
}

func (s *sem) extract(dest string, stream io.Reader) error {
	err := os.MkdirAll(dest, 0775)
	if err != nil {
		return err
	}

	if s.debug {
		log.Printf("Extracting the Go archive to %s", dest)
	}

	uncompressed, err := gzip.NewReader(stream)
	if err != nil {
		return err
	}

	tr := tar.NewReader(uncompressed)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}

		abs := filepath.Join(dest, filepath.FromSlash(header.Name))

		mode := header.FileInfo().Mode()

		switch {
		case mode.IsRegular():
			if err := os.MkdirAll(filepath.Dir(abs), 0755); err != nil {
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
			if err := os.Mkdir(abs, 0755); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unknown type: %s in %v", header.Name, mode)
		}
	}

	return nil
}

func createSymlink(dest string, local *localInfo) error {
	err := os.MkdirAll(dest, 0775)
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
