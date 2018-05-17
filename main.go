// main.go
package main

import (
	"bytes"
	"fmt"
	"os"
	"runtime"

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

func hello() (events.APIGatewayProxyResponse, error) {
	var resp string
	resp += fmt.Sprintln("Version:   ", runtime.Version())
	resp += fmt.Sprintln("GOARCH:    ", runtime.GOARCH)
	resp += fmt.Sprintln("GOOS:      ", runtime.GOOS)
	hostname, err := os.Hostname()
	if err != nil {
		resp += fmt.Sprintln("Hostname:  ", err)
	} else {
		resp += fmt.Sprintln("Hostname:  ", hostname)
	}
	resp += fmt.Sprintln("UID:       ", os.Getuid())
	err = unix.Uname(&utsname)
	if err != nil {
		resp += fmt.Sprintln("Uname:     ", err)
	} else {
		resp += fmt.Sprintln("Sysname:   ", toString(utsname.Sysname))
		resp += fmt.Sprintln("Nodename:  ", toString(utsname.Nodename))
		resp += fmt.Sprintln("Release:   ", toString(utsname.Release))
		resp += fmt.Sprintln("Version:   ", toString(utsname.Version))
		resp += fmt.Sprintln("Machine:   ", toString(utsname.Machine))
		resp += fmt.Sprintln("Domain:    ", toString(utsname.Domainname))
	}
	return events.APIGatewayProxyResponse{
		Body:       resp,
		StatusCode: 200,
	}, nil
}

func main() {
	lambda.Start(hello)
}
