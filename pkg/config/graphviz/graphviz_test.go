package configviz

import (
	"strings"
	"testing"

	"github.com/moul/advanced-ssh-config/pkg/config"
	. "github.com/smartystreets/goconvey/convey"
)

func TestGraph(t *testing.T) {
	Convey("Testing Graph()", t, func() {
		yamlConfig := `hosts:
  aaa:
    gateways: [bbb, direct]
  bbb:
    gateways: [ccc, aaa]
  ccc:
    gateways: [eee, direct]
  ddd:
  eee:
  fff:
    gateways: [eee, direct]
  ggg:
`
		conf := config.New()
		err := conf.LoadConfig(strings.NewReader(yamlConfig))
		So(err, ShouldBeNil)

		graph, err := Graph(conf)
		expected := `digraph G {
	fff->eee[ color=red, label=1 ];
	aaa->bbb[ color=red, label=1 ];
	bbb->ccc[ color=red, label=1 ];
	bbb->aaa[ color=red, label=2 ];
	ccc->eee[ color=red, label=1 ];
	aaa [ color=blue ];
	bbb [ color=blue ];
	ccc [ color=blue ];
	ddd [ color=blue ];
	eee [ color=blue ];
	fff [ color=blue ];
	ggg [ color=blue ];

}
`
		So(err, ShouldBeNil)
		So(graph, ShouldEqual, expected)
	})
}
