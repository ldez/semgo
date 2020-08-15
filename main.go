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
	if nArgs != 1 && !cfg.useModFile || cfg.useModFile && nArgs > 0 {
		log.Println("Error: missing go version.")
		usage()
	}

	if cfg.useModFile {
		file, err := ioutil.ReadFile("./go.mod")
		if err != nil {
			log.Fatal(err)
		}

		parse, err := modfile.Parse("./go.mod", file, nil)
		if err != nil {
			log.Fatal(err)
		}

		cfg.version = "go" + parse.Go.Version
	} else {
		cfg.version = flag.Arg(0)
	}

	if ok, _ := regexp.MatchString(`go(\d\.\d+)(?:.\d+)?`, cfg.version); !ok {
		log.Fatalf("invalid version (expected gox.x+[.x+]): %s", cfg.version)
	}

	smg := sem{
		client: &http.Client{Timeout: 30 * time.Second},
		debug:  *debug,
	}

	err := smg.getGo(cfg.dest, cfg.version)
	if err != nil {
		log.Fatal(err)
	}
}

func usage() {
	_, _ = os.Stderr.WriteString(`SemGo

semgo [flags] <version>

Flags:
`)
	flag.PrintDefaults()
	os.Exit(2)
}
