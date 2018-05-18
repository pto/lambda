// Lambda is a demonstration Amazon Web Service Lambda funcation.
package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"golang.org/x/sys/unix"
)

// handler is the Lambda handler function.
func handler(context context.Context) (events.APIGatewayProxyResponse, error) {
	var resp strings.Builder
	fmt.Fprintln(&resp, "Hello from Amazon λ!")
	fmt.Fprintln(&resp)
	fmt.Fprintln(&resp, "Version:      ", runtime.Version())
	fmt.Fprintln(&resp, "GOARCH:       ", runtime.GOARCH)
	fmt.Fprintln(&resp, "GOOS:         ", runtime.GOOS)
	fmt.Fprintln(&resp, "GOROOT:       ", runtime.GOROOT())
	fmt.Fprintln(&resp, "GOMAXPROCS:   ", runtime.GOMAXPROCS(0))
	fmt.Fprintln(&resp, "NumCPU:       ", runtime.NumCPU())
	fmt.Fprintln(&resp, "NumGoroutine: ", runtime.NumGoroutine())
	fmt.Fprintln(&resp, "Pagesize:     ", os.Getpagesize())
	fmt.Fprintln(&resp, "PID:          ", os.Getpid())
	fmt.Fprintln(&resp, "PPID:         ", os.Getppid())

	uid := strconv.Itoa(os.Getuid())
	user, err := user.LookupId(uid)
	var username string
	if err != nil {
		username = err.Error()
	} else {
		username = user.Name
	}
	fmt.Fprintf(&resp, "UID:           %d (%q)\n", os.Getuid(), username)

	fmt.Fprintln(&resp, "Time:         ", time.Now())
	deadline, ok := context.Deadline()
	if !ok {
		fmt.Fprintln(&resp, "Deadline:    <none>")
	} else {
		fmt.Fprintln(&resp, "Deadline:     ", deadline)
	}

	var utsname unix.Utsname
	err = unix.Uname(&utsname)
	if err != nil {
		fmt.Fprintln(&resp, "Uname:        ", err)
	} else {
		fmt.Fprintln(&resp, "Sysname:      ", toString(utsname.Sysname))
		fmt.Fprintln(&resp, "Nodename:     ", toString(utsname.Nodename))
		fmt.Fprintln(&resp, "Release:      ", toString(utsname.Release))
		fmt.Fprintln(&resp, "Version:      ", toString(utsname.Version))
		fmt.Fprintln(&resp, "Machine:      ", toString(utsname.Machine))
		fmt.Fprintln(&resp, "Domain:       ", toString(utsname.Domainname))
	}

	fmt.Fprintln(&resp, "Environment:")
	env := os.Environ()
	for _, e := range env {
		if strings.Contains(e, "_KEY") || strings.Contains(e, "_TOKEN") {
			continue
		}
		fmt.Fprintln(&resp, "  ", e)
	}
	fmt.Fprintln(&resp)

	wd, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(&resp, "Cannot read working directory: ", err)
	} else {
		fmt.Fprint(&resp, "Working directory ")
		listDir(&resp, wd)
	}

	listDir(&resp, "/usr/local/bin")
	listDir(&resp, "/usr/bin")
	listDir(&resp, "/bin")
	runtimeDir := os.Getenv("LAMBDA_RUNTIME_DIR")
	if runtimeDir == "" {
		fmt.Fprintln(&resp, "No runtime directory")
	} else {
		listDir(&resp, runtimeDir)
	}

	runCommand(&resp, "top", "-b", "-n", "1")
	runCommand(&resp, "df")
	listFile(&resp, "/proc/cpuinfo")

	return events.APIGatewayProxyResponse{
		Body: resp.String(),
		Headers: map[string]string{
			"content-type": "text/plain; charset=utf-8",
			"x-pto":        "hello, Lambda",
		},
		StatusCode: 200,
	}, nil
}

func main() {
	lambda.Start(handler)
}

// toString converts fields in the Utsname structure to strings.
func toString(buf [65]byte) string {
	b := buf[:]
	i := bytes.IndexByte(b, 0)
	if i < 0 {
		return string(b)
	}
	return string(b[:i])
}

// listDir lists the files in directory d
func listDir(w io.Writer, d string) {
	fmt.Fprintf(w, "%s:\n", d)

	// Read directory
	dir, err := os.Open(d)
	if err != nil {
		fmt.Fprintln(w, "   Cannot open:", err)
		return
	}
	defer dir.Close()
	fileinfos, err := dir.Readdir(0)
	if err != nil {
		fmt.Fprintln(w, "   Cannot read directory:", err)
		return
	}

	// Get filenames
	var files []string
	for _, fi := range fileinfos {
		files = append(files, fi.Name())
	}
	sort.Strings(files)
	var maxWidth int
	for _, f := range files {
		if len(f) > maxWidth {
			maxWidth = len(f)
		}
	}

	if len(files) == 0 {
		fmt.Fprintln(w, "(none)")
		fmt.Fprintln(w)
		return
	}

	// Write filenames in column major order
	columns := 5
	rows := (len(files) + columns - 1) / columns // round up
	tabw := tabwriter.NewWriter(w, maxWidth, 0, 1, ' ', 0)
	for r := 0; r < rows; r++ {
		for c := 0; c < columns; c++ {
			if c*rows+r < len(files) {
				fmt.Fprintf(tabw, "%s\t", files[c*rows+r])
			} else {
				fmt.Fprintf(tabw, "\t")
			}
		}
		fmt.Fprintln(tabw)
	}
	fmt.Fprintln(tabw)
	tabw.Flush()
}

func runCommand(w io.Writer, cmd string, args ...string) {
	fmt.Fprintf(w, "%s:\n", cmd)
	out, err := exec.Command(cmd, args...).Output()
	if err != nil {
		fmt.Fprintf(w, "Cannot run %s: %s\n", cmd, err)
		return
	}
	fmt.Fprintln(w, string(out))
}

func listFile(w io.Writer, filename string) {
	f, err := os.Open(filename)
	if err != nil {
		fmt.Fprintf(w, "Cannot open %s: %s\n", filename, err)
		return
	}
	defer f.Close()
	buf, err := ioutil.ReadAll(f)
	if err != nil {
		fmt.Fprintf(w, "Cannot read %s: %s\n", filename, err)
		return
	}
	fmt.Fprintf(w, "%s:\n", filename)
	fmt.Fprintln(w, string(buf))
}
