// Lambda is a demonstration Amazon Web Service Lambda Funcation.
package main

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/user"
	"runtime"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"golang.org/x/sys/unix"
)

var utsname unix.Utsname

func toString(buf [65]byte) string {
	b := buf[:]
	i := bytes.IndexByte(b, 0)
	if i < 0 {
		return string(b)
	} else {
		return string(b[:i])
	}
}

func listDir(resp *strings.Builder, d string) {
	// Read directory
	dir, err := os.Open(d)
	if err != nil {
		fmt.Fprintln(resp, "   Cannot open:", err)
		return
	}
	defer dir.Close()
	fileinfos, err := dir.Readdir(0)
	if err != nil {
		fmt.Fprintln(resp, "   Cannot read directory:", err)
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

	log.Print("Number of files in ", d, " is ", len(files))
	if len(files) == 0 {
		fmt.Fprintln(resp, "(none)")
		return
	}

	// Write filenames
	w := tabwriter.NewWriter(resp, maxWidth, maxWidth, 1, ' ', 0)
	const columns = 5
loop:
	for r := 0; r < len(files)/columns+1; r += columns {
		for c := 0; c < columns; c++ {
			if r*columns+c >= len(files) {
				if c > 0 {
					fmt.Println(w)
				}
				break loop
			}
			if c > 0 {
				fmt.Fprint(w, "\t")
			}
			fmt.Fprint(w, files[r*columns+c])
		}
		fmt.Fprintln(w)
	}
	w.Flush()
	fmt.Fprintln(w)
}

func hello(context context.Context) (events.APIGatewayProxyResponse, error) {
	var resp strings.Builder
	fmt.Fprintln(&resp, "Hello from Amazon λ!")
	fmt.Fprintln(&resp, "Version:   ", runtime.Version())
	fmt.Fprintln(&resp, "GOARCH:    ", runtime.GOARCH)
	fmt.Fprintln(&resp, "GOOS:      ", runtime.GOOS)
	fmt.Fprintln(&resp, "Pagesize:  ", os.Getpagesize())
	fmt.Fprintln(&resp, "PID:       ", os.Getpid())
	fmt.Fprintln(&resp, "PPID:      ", os.Getppid())

	user, err := user.Current()
	var username string
	if err != nil {
		fmt.Fprintln(&resp, "Cannot get current user: ", err.Error())
	} else {
		username = user.Name
	}
	fmt.Fprintf(&resp, "UID:        %d (%q)\n", os.Getuid(), username)

	wd, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(&resp, "Cannot read working directory: ", err)
	} else {
		fmt.Fprintln(&resp, "Directory: ", wd)
	}
	listDir(&resp, wd)

	fmt.Fprintln(&resp, "Time:      ", time.Now())
	deadline, ok := context.Deadline()
	if !ok {
		fmt.Fprintln(&resp, "Deadline: <none>")
	} else {
		fmt.Fprintln(&resp, "Deadline:  ", deadline)
	}

	err = unix.Uname(&utsname)
	if err != nil {
		fmt.Fprintln(&resp, "Uname:     ", err)
	} else {
		fmt.Fprintln(&resp, "Sysname:   ", toString(utsname.Sysname))
		fmt.Fprintln(&resp, "Nodename:  ", toString(utsname.Nodename))
		fmt.Fprintln(&resp, "Release:   ", toString(utsname.Release))
		fmt.Fprintln(&resp, "Version:   ", toString(utsname.Version))
		fmt.Fprintln(&resp, "Machine:   ", toString(utsname.Machine))
		fmt.Fprintln(&resp, "Domain:    ", toString(utsname.Domainname))
	}

	resp.WriteString("Environment:\n")
	env := os.Environ()
	for _, e := range env {
		if strings.Contains(e, "_KEY") || strings.Contains(e, "_TOKEN") {
			continue
		}
		fmt.Fprintln(&resp, "  ", e)
	}

	fmt.Fprintln(&resp, "/usr/local/bin:")
	listDir(&resp, "/usr/local/bin")
	fmt.Fprintln(&resp, "/usr/bin:")
	listDir(&resp, "/usr/bin")
	fmt.Fprintln(&resp, "/bin:")
	listDir(&resp, "/bin")
	runtime := os.Getenv("LAMBDA_RUNTIME_DIR")
	if runtime == "" {
		fmt.Fprintln(&resp, "No runtime directory")
	} else {
		fmt.Fprintln(&resp, runtime)
		listDir(&resp, runtime)
	}

	cmd := exec.Command("df")
	out, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Fprintln(&resp, "Cannot run df:", err)
	} else {
		if err := cmd.Start(); err != nil {
			fmt.Fprintln(&resp, "Cannot start df:", err)
		} else {
			output, err := ioutil.ReadAll(out)
			if err != nil {
				fmt.Fprintln(&resp, "Cannot read df output:", err)
			} else {
				fmt.Fprintln(&resp, string(output))
			}
		}
	}

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
	lambda.Start(hello)
}
