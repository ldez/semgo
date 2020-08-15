package main

import (
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"time"

	"golang.org/x/mod/modfile"
)

const envGoRoot = "GOROOT"

const modPath = "./go.mod"

type config struct {
	dest       string
	version    string
	useModFile bool
}

func main() {
	cfg := config{}

	flag.StringVar(&cfg.dest, "dest", "/usr/local/golang/", "Path to the Go versions storage in SemaphoreCI.")
	flag.BoolVar(&cfg.useModFile, "mod", false, "")

	debug := flag.Bool("debug", false, "Debug mode.")
	help := flag.Bool("h", false, "Show this help.")

	flag.Usage = usage
	flag.Parse()

	if *help {
		usage()
	}

	nArgs := flag.NArg()
	if nArgs != 1 && !cfg.useModFile || cfg.useModFile && nArgs > 1 {
		log.Println("Error: missing go version.")
		usage()
	}

	cfg.version = flag.Arg(0)

	if cfg.version != "" {
		ok, _ := regexp.MatchString(`go(\d\.\d+)(?:.\d+)?`, cfg.version)
		if !ok {
			log.Fatalf("invalid version (expected gox.x+[.x+]): %s", cfg.version)
		}
	}

	if cfg.useModFile {
		mod, err := readGoMod()
		if err != nil {
			log.Fatal(err)
		}

		if mod != nil {
			cfg.version = "go" + mod.Go.Version
		}
	}

	if cfg.version == "" {
		log.Fatal("The version is missing.")
	}

	if *debug {
		log.Printf("Debug mode: %#v", cfg)
	}

	smg := sem{
		client: &http.Client{Timeout: 30 * time.Second},
		debug:  *debug,
		goRoot: envGoRoot,
	}

	err := smg.getGo(cfg.dest, cfg.version)
	if err != nil {
		log.Fatal(err)
	}
}

func readGoMod() (*modfile.File, error) {
	_, err := os.Stat(modPath)
	if err != nil {
		return nil, nil
	}

	file, err := ioutil.ReadFile(modPath)
	if err != nil {
		return nil, err
	}

	return modfile.Parse(modPath, file, nil)
}

func usage() {
	_, _ = os.Stderr.WriteString(`SemGo

semgo [flags] <version>

Flags:
`)
	flag.PrintDefaults()
	os.Exit(2)
}
