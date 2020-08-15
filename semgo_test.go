package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"

	"github.com/ldez/grignotin/version"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_getLocalVersions(t *testing.T) {
	smg := sem{
		debug: true,
	}

	testCases := []struct {
		desc     string
		dir      string
		expected map[string]localInfo
	}{
		{
			desc:     "no version",
			dir:      "./fixtures/usr/local/",
			expected: map[string]localInfo{},
		},
		{
			desc: "with versions",
			dir:  "./fixtures/usr/local/golang/",
			expected: map[string]localInfo{
				"go1.10": {Version: "1.10.8", Path: "fixtures/usr/local/golang/1.10.8"},
				"go1.11": {Version: "1.11.13", Path: "fixtures/usr/local/golang/1.11.13"},
				"go1.12": {Version: "1.12.17", Path: "fixtures/usr/local/golang/1.12.17"},
				"go1.13": {Version: "1.13.14", Path: "fixtures/usr/local/golang/1.13.14"},
				"go1.14": {Version: "1.14.6", Path: "fixtures/usr/local/golang/1.14.6"},
			},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			versions, err := smg.getLocalVersions(test.dir)
			require.NoError(t, err)

			assert.Equal(t, test.expected, versions)
		})
	}
}

func Test_extractVersionFromGoRoot(t *testing.T) {
	smg := sem{
		debug:  true,
		goRoot: "GROOT",
	}

	testCases := []struct {
		desc     string
		val      string
		expected *localInfo
	}{
		{
			desc:     "M.m.p",
			val:      "/usr/local/golang/1.10.8/go",
			expected: &localInfo{Version: "1.10.8", Path: "/usr/local/golang/1.10.8/go"},
		},
		{
			desc:     "M.m",
			val:      "/usr/local/golang/1.15/go",
			expected: &localInfo{Version: "1.15", Path: "/usr/local/golang/1.15/go"},
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			require.NoError(t, os.Setenv("GROOT", test.val))
			t.Cleanup(func() { _ = os.Unsetenv("GROOT") })

			root, err := smg.extractVersionFromGoRoot()
			require.NoError(t, err)

			assert.Equal(t, test.expected, root)
		})
	}
}

func Test_findReleaseInfo(t *testing.T) {
	smg := sem{
		debug: true,
	}

	file, err := ioutil.ReadFile("./fixtures/releases.json")
	require.NoError(t, err)

	var releases []version.Release
	err = json.Unmarshal(file, &releases)
	require.NoError(t, err)

	testCases := []struct {
		desc     string
		version  string
		expected *version.File
	}{
		{
			desc: "empty version",
		},
		{
			desc:    "go1.14",
			version: "go1.14",
			expected: &version.File{
				Filename: "go1.14.7.linux-amd64.tar.gz",
				OS:       "linux",
				Arch:     "amd64",
				Version:  "go1.14.7",
				SHA256:   "4a7fa60f323ee1416a4b1425aefc37ea359e9d64df19c326a58953a97ad41ea5",
				Size:     123747311,
				Kind:     "archive",
			},
		},
		{
			desc:    "go1.14.7",
			version: "go1.14.7",
			expected: &version.File{
				Filename: "go1.14.7.linux-amd64.tar.gz",
				OS:       "linux",
				Arch:     "amd64",
				Version:  "go1.14.7",
				SHA256:   "4a7fa60f323ee1416a4b1425aefc37ea359e9d64df19c326a58953a97ad41ea5",
				Size:     123747311,
				Kind:     "archive",
			},
		},
		{
			desc:    "go1.15",
			version: "go1.15",
			expected: &version.File{
				Filename: "go1.15.linux-amd64.tar.gz",
				OS:       "linux",
				Arch:     "amd64",
				Version:  "go1.15",
				SHA256:   "2d75848ac606061efe52a8068d0e647b35ce487a15bb52272c427df485193602",
				Size:     121136135,
				Kind:     "archive",
			},
		},
		{
			desc:     "non existing version",
			version:  "go1.15.1",
			expected: nil,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			info := smg.findReleaseInfo(releases, test.version)

			assert.Equal(t, test.expected, info)
		})
	}
}
