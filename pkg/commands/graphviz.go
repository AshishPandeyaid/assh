package commands

import (
	"fmt"

	"github.com/urfave/cli"

	"github.com/moul/advanced-ssh-config/pkg/config"
	"github.com/moul/advanced-ssh-config/pkg/config/graphviz"
	"github.com/moul/advanced-ssh-config/pkg/logger"
)

func cmdGraphviz(c *cli.Context) error {
	conf, err := config.Open(c.GlobalString("config"))
	if err != nil {
		logger.Logger.Fatalf("Cannot load configuration: %v", err)
		return nil
	}

	settings := configviz.GraphSettings{
		ShowIsolatedHosts: c.Bool("show-isolated-hosts"),
		NoResolveWildcard: c.Bool("no-resolve-wildcard"),
		NoInherits:        c.Bool("no-inheritance-links"),
	}
	graph, err := configviz.Graph(conf, &settings)
	if err != nil {
		logger.Logger.Fatalf("failed to build graph: %v", err)
		return nil
	}

	fmt.Println(graph)
	return nil
}
