package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var (
	source = flag.String("source", "vendor", "source directory")
	target = flag.String("target", "$GOPATH/bin", "target directory")
	quiet  = flag.Bool("quiet", false, "disable output")
)

func main() {
	flag.Parse()

	packages := flag.Args()
	if len(packages) < 1 {
		fail(errors.New("no packages: specify a package"))
	}
	print(fmt.Sprintf("packages: %s", strings.Join(packages, " ")))

	gopath, err := ioutil.TempDir("", "go-vendorinstall-gopath")
	if err != nil {
		fail(err)
	}
	defer os.RemoveAll(gopath)
	print(fmt.Sprintf("gopath: %s", gopath))

	if err := link(gopath, *source); err != nil {
		fail(err)
	}

	if out, err := install(gopath, *target, packages); err != nil {
		print(string(out))
		fail(err)
	}
}

func print(msg string) {
	if !*quiet {
		fmt.Println(msg)
	}
}

func fail(err error) {
	fmt.Printf("error: %s", err.Error())
	os.Exit(1)
}

func link(gopath, source string) error {
	srcdir, err := filepath.Abs(source)
	if err != nil {
		return err
	}

	linkto := filepath.Join(gopath, "src")
	if err := os.MkdirAll(linkto, 0777); err != nil {
		return err
	}

	files, err := ioutil.ReadDir(srcdir)
	if err != nil {
		return err
	}

	for _, file := range files {
		real := filepath.Join(srcdir, file.Name())
		link := filepath.Join(linkto, file.Name())
		if err := os.Symlink(real, link); err != nil {
			return err
		}
	}

	return nil
}

func install(gopath, target string, packages []string) ([]byte, error) {
	exe, err := exec.LookPath("go")
	if err != nil {
		return nil, err
	}

	gobin, err := filepath.Abs(target)
	if err != nil {
		return nil, err
	}

	args := append([]string{"install"}, packages...)
	env := []string{fmt.Sprintf("GOPATH=%s", gopath), fmt.Sprintf("GOBIN=%s", gobin)}
	print(fmt.Sprintf("%s %s", exe, strings.Join(args, " ")))

	cmd := exec.Command(exe, args...)
	cmd.Env = env

	return cmd.CombinedOutput()
}
