package main

// dp(1) - directory pipe

import (
	"archive/tar"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"syscall"
)

var inp  = "%i"
var outp = "%o"

var keep = false

var cmd []string

var ind  string
var outd string
var tmpd = os.TempDir()

func help(n int) {
	argv0 := path.Base(os.Args[0])
	fmt.Fprintf(os.Stderr, "%s [-k] [-i %%i] [-o %%o] [-t /tmp] <cmd> [args ...]\n", argv0)
	fmt.Fprintf(os.Stderr, "%s [-h]\n", argv0)
	os.Exit(n)
}

func fails(err error) {
	argv0 := path.Base(os.Args[0])
	var eerr *exec.ExitError
	if errors.As(err, &eerr) {
		log.Println(argv0, ": ", err)
		os.Exit(eerr.ProcessState.ExitCode())
	} else {
		log.Fatal(argv0, ": ", err)
	}
}

// args is altered
func initCmd(args []string) {
	var err error
	for i := 0; i < len(args); i++ {
		if args[i] == inp {
			if ind == "" {
				ind, err = os.MkdirTemp(tmpd, "dp-")
				if err != nil {
					fails(err)
				}
			}
			args[i] = ind
		} else if args[i] == outp {
			if outd == "" {
				outd, err = os.MkdirTemp(tmpd, "dp-")
				if err != nil {
					fails(err)
				}
			}
			args[i] = outd
		}
	}

	cmd = args
}

func init() {
	i := 1
	for ; i < len(os.Args); i++ {
		if os.Args[i] == "-h" {
			help(0)
		} else if os.Args[i] == "-k" {
			keep = true
		} else if os.Args[i] == "-i" || os.Args[i] == "-o" || os.Args[i] == "-t" {
			if i == len(os.Args)-1 {
				help(1)
			}
			i++
			if os.Args[i-1] == "-i" {
				inp = os.Args[i]
			} else if os.Args[i-1] == "-o" {
				outp = os.Args[i]
			} else {
				tmpd = os.Args[i]
			}
		} else {
			break
		}
	}

	// missing command
	if i == len(os.Args) {
		help(2)
	}

	initCmd(os.Args[i:])

	if keep {
		if ind != "" {
			fmt.Fprintf(os.Stderr, "Input directory (%s): %s\n", cmd[0], ind)
		}
		if outd != "" {
			fmt.Fprintf(os.Stderr, "Output directory (%s): %s\n", cmd[0], outd)
		}
	}
}

// extract stdin to ind
func getIn() error {
	tr := tar.NewReader(os.Stdin)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if hdr.Typeflag == tar.TypeDir {
			if err := os.MkdirAll(hdr.Name, os.ModePerm); err != nil {
				return err
			}
			continue
		}

		fn := filepath.Join(ind, hdr.Name)

		if err := os.MkdirAll(filepath.Dir(fn), os.ModePerm); err != nil {
			return err
		}

		fh, err := os.OpenFile(fn, os.O_RDWR|os.O_CREATE, fs.FileMode(hdr.Mode))
		if err != nil {
			return err
		}
		defer fh.Close()

		if _, err := io.Copy(fh, tr); err != nil {
			return err
		}
	}

	return nil
}

// NOTE: should work on any UNIX, but no more.
func getUidGid(path string) (int, int, error) {
	// reasonable default
	uid := os.Getuid()
	gid := os.Getgid()

	stat, err := os.Stat(path)
	if err != nil {
		return 0, 0, err
	}

	if x, ok := stat.Sys().(*syscall.Stat_t); ok {
	    uid = int(x.Uid)
	    gid = int(x.Gid)
	}

	return uid, gid, nil
}

// compress outd to stdout
func setOut() error {
	tw := tar.NewWriter(os.Stdout)

	err := filepath.Walk(outd, func(path string, info fs.FileInfo, err error) error {
		if path == outd {
			return nil
		}

		if err != nil {
			return err
		}

		// XXX IIRC, filepath.WalkDir() doesn't fill a fs.FileInfo, so it
		// should be more efficient, as we call os.Stat() here anyway.
		uid, gid, err := getUidGid(path)
		if err != nil {
			return err
		}

		// outd is generated from os.MkdirTemp(), which I guess
		// we can expect to keep never returning a "/"-terminated path.
		hdr := &tar.Header{
			Name    : strings.TrimPrefix(path, outd+string(os.PathSeparator)),
			Mode    : int64(info.Mode()),
			Size    : info.Size(),
			ModTime : info.ModTime(),
			Uid     : uid,
			Gid     : gid,
		}

		// NOTE: if path had been "/"-terminated, (.Walk trims it),
		// Typeflag would have been automatically be promoted to a
		// TypeDir.
		if info.IsDir() {
			hdr.Typeflag = tar.TypeDir
		}

		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}

		if !info.IsDir() {
			xs, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			if _, err := tw.Write(xs); err != nil {
				return err
			}
		}

		return nil
	})

	if err == nil {
		err = tw.Close()
	}

	return err
}

func cleanup(fail bool) {
	if keep {
		return
	}

	// always prints error
	f := log.Print
	if fail {
		f = log.Fatal
	}

	if ind != "" {
		if err := os.RemoveAll(ind); err != nil {
			f(err)
		}
	}

	if outd != "" {
		if err := os.RemoveAll(outd); err != nil {
			f(err)
		}
	}
}

func runCmd() error {
	com := exec.Command(cmd[0], cmd[1:]...)

	if ind == "" {
		com.Stdin = os.Stdin
	}

	// We'll be sending a .tar on os.Stdout; so anything
	// from com.Stdout is sent to os.Stderr instead
	// of being lost, or potentially interfering with our
	// tar(1) output.
	com.Stdout = os.Stderr

	com.Stderr = os.Stderr

	if err := com.Run(); err != nil {
		return err
	}

	return nil
}

func main() {
	if ind != "" {
		if err := getIn(); err != nil {
			cleanup(false)
			fails(err)
		}
	}

	if err := runCmd(); err != nil {
		cleanup(false)
		fails(err)
	}

	if outd != "" {
		if err := setOut(); err != nil {
			cleanup(false)
			fails(err)
		}
	}

	cleanup(true)
}
