package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"regexp"
	"time"
)

type config struct {
	dest    string
	version string
}

func main() {
	cfg := config{}

	flag.StringVar(&cfg.dest, "dest", "/usr/local/golang/", "Path to the Go versions storage in SemaphoreCI.")

	debug := flag.Bool("debug", true, "Debug mode.")
	help := flag.Bool("h", false, "Show this help.")

	flag.Usage = usage
	flag.Parse()

	nArgs := flag.NArg()
	if nArgs != 1 {
		usage()
	}

	cfg.version = flag.Arg(0)

	if *help {
		usage()
	}

	expr := regexp.MustCompile(`go(\d\.\d+)(?:.\d+)?`)

	if ok := expr.MatchString(cfg.version); !ok {
		log.Fatalf("invalid version (expected gox.x+[.x+]): %s", cfg.version)
	}

	smg := sem{
		client: &http.Client{Timeout: 2 * time.Second},
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
