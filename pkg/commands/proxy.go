package commands

import (
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"strings"

	shlex "github.com/flynn/go-shlex"
	"github.com/moul/advanced-ssh-config/vendor/github.com/Sirupsen/logrus"
	"github.com/moul/advanced-ssh-config/vendor/github.com/codegangsta/cli"

	"github.com/moul/advanced-ssh-config/pkg/config"
)

func cmdProxy(c *cli.Context) {
	if len(c.Args()) < 1 {
		logrus.Fatalf("assh: \"proxy\" requires 1 argument. See 'assh proxy --help'.")
	}

	host, port, err := configGetHostPort(c.Args()[0], c.Int("port"))
	if err != nil {
		logrus.Fatalf("Cannot get host '%s': %v", c.Args()[0], err)
	}

	err = proxyGo(host, port)
	//err = proxyCommand("nc -v -w 180 -G 5 {host} {port}", host, port)
	if err != nil {
		logrus.Fatalf("Proxy error: %v", err)
	}
}

func configGetHostPort(dest string, portFlag int) (string, uint, error) {
	conf, err := config.Open()
	if err != nil {
		return "", 0, err
	}

	// Get host configuration
	host := conf.GetHostSafe(dest)

	// Dial
	var port uint
	if portFlag > 0 {
		port = uint(portFlag)
	} else {
		port = host.Port
	}

	return host.Host, port, nil
}

func proxyCommand(command string, host string, port uint) error {
	command = strings.Replace(command, "{host}", host, -1)
	command = strings.Replace(command, "{port}", fmt.Sprintf("%d", port), -1)
	args, err := shlex.Split(command)
	if err != nil {
		return err
	}
	spawn := exec.Command(args[0], args[1:]...)
	spawn.Stdout = os.Stdout
	spawn.Stdin = os.Stdin
	spawn.Stderr = os.Stderr
	return spawn.Run()
}

func proxyGo(host string, port uint) error {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return err
	}

	defer conn.Close()

	logrus.Debugf("Connected to %s:%d\n", host, port)

	// Create Stdio pipes
	c1 := readAndWrite(conn, os.Stdout)
	c2 := readAndWrite(os.Stdin, conn)

	select {
	case err = <-c1:
	case err = <-c2:
	}
	if err != nil {
		return err
	}

	return nil
}

func readAndWrite(r io.Reader, w io.Writer) <-chan error {
	// Fixme: add an error channel
	buf := make([]byte, 1024)
	c := make(chan error)

	go func() {
		for {
			// Read
			n, err := r.Read(buf)
			if err != nil {
				if err != io.EOF {
					c <- err
				}
				break
			}

			// Write
			_, err = w.Write(buf[0:n])
			if err != nil {
				c <- err
			}
		}
		c <- nil
	}()
	return c
}
