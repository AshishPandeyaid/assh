package commands

import (
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"strings"

	"github.com/moul/advanced-ssh-config/vendor/github.com/Sirupsen/logrus"
	"github.com/moul/advanced-ssh-config/vendor/github.com/codegangsta/cli"
	shlex "github.com/moul/advanced-ssh-config/vendor/github.com/flynn/go-shlex"

	"github.com/moul/advanced-ssh-config/pkg/config"
)

func cmdProxy(c *cli.Context) {
	if len(c.Args()) < 1 {
		logrus.Fatalf("assh: \"proxy\" requires 1 argument. See 'assh proxy --help'.")
	}

	host, err := computeHost(c.Args()[0], c.Int("port"))
	if err != nil {
		logrus.Fatalf("Cannot get host '%s': %v", c.Args()[0], err)
	}

	err = proxy(host)
	if err != nil {
		logrus.Fatalf("Proxy error: %v", err)
	}
}

func computeHost(dest string, portOverride int) (*config.Host, error) {
	conf, err := config.Open()
	if err != nil {
		return nil, err
	}

	host := conf.GetHostSafe(dest)
	if portOverride > 0 {
		host.Port = uint(portOverride)
	}

	return host, nil
}

func proxy(host *config.Host) error {
	if len(host.Gateways) > 0 {
		logrus.Debugf("Trying gateways: %s", host.Gateways)
		for _, gateway := range host.Gateways {
			if gateway == "direct" {
				err := proxyDirect(host)
				if err != nil {
					logrus.Errorf("Failed to use 'direct' connection")
				}
			} else {
				gatewayHost, err := computeHost(gateway, 0)
				if err != nil {
					logrus.Fatalf("Cannot get host '%s': %v", gateway, err)
				}

				if host.ProxyCommand == "" {
					host.ProxyCommand = "nc %h %d"
				}
				command := "ssh %name -- " + commandApplyHost(host.ProxyCommand, host)

				logrus.Debugf("Using gateway '%s': %s", gateway, command)
				err = proxyCommand(gatewayHost, command)
				if err != nil {
					logrus.Errorf("Cannot use gateway '%s': %v", gateway, err)
				}
				if err == nil {
					return nil
				}
			}
		}
		return fmt.Errorf("No such available gateway")
	}

	logrus.Debugf("Connecting without gateway")
	return proxyDirect(host)
}

func commandApplyHost(command string, host *config.Host) string {
	command = strings.Replace(command, "%name", host.Name, -1)
	command = strings.Replace(command, "%h", host.Host, -1)
	command = strings.Replace(command, "%p", fmt.Sprintf("%d", host.Port), -1)
	return command
}

func proxyDirect(host *config.Host) error {
	if host.ProxyCommand != "" {
		return proxyCommand(host, host.ProxyCommand)
	}
	return proxyGo(host)
}

func proxyCommand(host *config.Host, command string) error {
	command = commandApplyHost(command, host)
	args, err := shlex.Split(command)
	logrus.Debugf("ProxyCommand: %s", command)
	if err != nil {
		return err
	}
	spawn := exec.Command(args[0], args[1:]...)
	spawn.Stdout = os.Stdout
	spawn.Stdin = os.Stdin
	spawn.Stderr = os.Stderr
	return spawn.Run()
}

func proxyGo(host *config.Host) error {
	if host.Host == "" {
		host.Host = host.Name
	}

	if host.Resolve != "" {
		logrus.Debugf("Resolving host: '%s' using '%s'", host.Host, host.Resolve)
		// FIXME: resolve using custom dns server
		results, err := net.LookupAddr(host.Host)
		if err != nil {
			return err
		}
		if len(results) > 0 {
			host.Host = results[0]
		}
	}

	logrus.Debugf("Connecting to %s:%d", host.Host, host.Port)
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", host.Host, host.Port))
	if err != nil {
		return err
	}
	defer conn.Close()

	logrus.Debugf("Connected to %s:%d", host.Host, host.Port)

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
