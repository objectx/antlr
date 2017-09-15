package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"io/ioutil"
	"strings"
	"github.com/pkg/errors"
)

var (
	progPath  string
	progName  string
	antlrJar  string
	beVerbose = false
)

func init() {
	progPath = getProgramPath("antlr")
	progName = filepath.Base(progPath)
}

func main() {
	var err error
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n", progName)
		flag.PrintDefaults()
		os.Exit(1)
	}
	flag.BoolVar(&beVerbose, "v", false, "Be verbose")
	grun := flag.Bool("grun", false, "Run test-rig")
	flag.StringVar(&antlrJar, "antlr", "", "Path to ANTLR jar")
	flag.Parse()
	java, err := findJava()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: Failed to find java\n", progName)
		os.Exit(1)
	}
	verbose("%s: JAVA = %s\n", progName, java)
	cmdArgs, err := buildAntlrCommandArgs()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s:error: %v\n", progName, err)
		os.Exit(1)
	}
	if *grun {
		cmdArgs = append(cmdArgs, "org.antlr.v4.gui.TestRig")
	} else {
		cmdArgs = append(cmdArgs, "org.antlr.v4.Tool")
	}
	cmdArgs = append(cmdArgs, flag.Args()...)
	verbose("%s: cmd = %v\n", progName, cmdArgs)
	cmd := exec.Command(java, cmdArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s:error: %v\n", progName, err)
		os.Exit(1)
	}
	os.Exit(0)
}

func getProgramPath(def string) string {
	p, err := os.Executable()
	if err != nil {
		return def
	}
	return p
}

func verbose(format string, args ...interface{}) {
	if beVerbose {
		fmt.Fprintf(os.Stderr, format, args...)
	}
}

func findJava() (string, error) {
	p := os.ExpandEnv("${JAVA_HOME}/bin/java")
	p, err := exec.LookPath(p)
	if err != nil {
		return "", err
	}
	return filepath.Clean(p), nil
}

func buildAntlrCommandArgs() ([]string, error) {
	var err error
	exeDir := filepath.Dir(progPath)
	antlr := antlrJar
	if len(antlr) == 0 {
		antlr, err = findAntlr(exeDir)
		if err != nil {
			return nil, err
		}
	}
	orgClassPath, ok := os.LookupEnv("CLASSPATH")
	var classPath string
	if ok {
		classPath = fmt.Sprintf("%s%c%s", antlr, os.PathListSeparator, orgClassPath)
	} else {
		classPath = antlr
	}
	return []string{
		"-cp",
		classPath,
	}, nil
}

func findAntlr (dir string) (string, error) {
	items, err := ioutil.ReadDir(dir)
	if err != nil {
		return "", err
	}
	for _, f := range items {
		name := f.Name ()
		if strings.HasPrefix(name, "antlr-") && strings.HasSuffix(name, "-complete.jar") {
			return filepath.Clean (filepath.Join (dir, name)), nil
		}
	}
	return "", errors.Errorf("missing ANTLR")
}
