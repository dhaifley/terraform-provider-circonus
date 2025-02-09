package circonus

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccCirconusCheckCAQL_basic(t *testing.T) {
	checkName := fmt.Sprintf("Consul's Go GC latency (Merged Histogram) - %s", acctest.RandString(5))
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDestroyCirconusCheckBundle,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(string(testAccCirconusCheckCAQLConfigFmt), checkName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("circonus_check.go_gc_latency", "active", "true"),
					resource.TestCheckResourceAttr("circonus_check.go_gc_latency", "collector.#", "1"),
					resource.TestCheckResourceAttr("circonus_check.go_gc_latency", "collector.0.id", "/broker/1490"),
					resource.TestCheckResourceAttr("circonus_check.go_gc_latency", "caql.#", "1"),
					resource.TestCheckResourceAttr("circonus_check.go_gc_latency", "caql.0.query", `search:metric:histogram("*consul*runtime`+"`"+`gc_pause_ns* (active:1)") | histogram:merge() | histogram:percentile(99)`+"\n"),
					resource.TestCheckResourceAttr("circonus_check.go_gc_latency", "name", checkName),
					resource.TestCheckResourceAttr("circonus_check.go_gc_latency", "period", "60s"),
					resource.TestCheckResourceAttr("circonus_check.go_gc_latency", "metric.#", "1"),

					resource.TestCheckResourceAttr("circonus_check.go_gc_latency", "tags.#", "4"),
					resource.TestCheckResourceAttr("circonus_check.go_gc_latency", "tags.0", "app:consul"),
					resource.TestCheckResourceAttr("circonus_check.go_gc_latency", "tags.1", "author:terraform"),
					resource.TestCheckResourceAttr("circonus_check.go_gc_latency", "tags.2", "lifecycle:unittest"),
					resource.TestCheckResourceAttr("circonus_check.go_gc_latency", "tags.3", "source:goruntime"),
					resource.TestCheckResourceAttr("circonus_check.go_gc_latency", "target", "q._caql"),
					resource.TestCheckResourceAttr("circonus_check.go_gc_latency", "type", "caql"),
				),
			},
		},
	})
}

const testAccCirconusCheckCAQLConfigFmt = `
variable "test_tags" {
  type = list(string)
  default = [ "app:consul", "author:terraform", "lifecycle:unittest", "source:goruntime" ]
}

resource "circonus_check" "go_gc_latency" {
  active = true
  name = "%s"
  period = "60s"

  collector {
    id = "/broker/1490"
  }

  caql {
    query = <<EOF
search:metric:histogram("*consul*runtime` + "`" + `gc_pause_ns* (active:1)") | histogram:merge() | histogram:percentile(99)
EOF
  }

  metric {
    name = "output[1]"
    type = "histogram"
  }

  tags = "${var.test_tags}"
}
`
