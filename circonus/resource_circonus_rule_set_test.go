package circonus

import (
	"fmt"
	"strings"
	"testing"

	api "github.com/circonus-labs/go-apiclient"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var rulesetCheckName = fmt.Sprintf("ICMP Ping check - %s", acctest.RandString(5))

func TestAccCirconusRuleSet_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDestroyCirconusRuleSet,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testAccCirconusRuleSetConfigFmt, rulesetCheckName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("circonus_rule_set.icmp-latency-alarm", "check"),
					resource.TestCheckResourceAttr("circonus_rule_set.icmp-latency-alarm", "metric_name", "maximum"),
					resource.TestCheckResourceAttr("circonus_rule_set.icmp-latency-alarm", "metric_type", "numeric"),
					resource.TestCheckResourceAttr("circonus_rule_set.icmp-latency-alarm", "notes", "Simple check to create notifications based on ICMP performance."),
					resource.TestCheckResourceAttr("circonus_rule_set.icmp-latency-alarm", "link", "https://wiki.example.org/playbook/what-to-do-when-high-latency-strikes"),
					resource.TestCheckResourceAttr("circonus_rule_set.icmp-latency-alarm", "if.#", "7"),

					resource.TestCheckResourceAttr("circonus_rule_set.icmp-latency-alarm", "if.0.value.#", "1"),
					resource.TestCheckResourceAttr("circonus_rule_set.icmp-latency-alarm", "if.0.value.0.absent", "70"),
					resource.TestCheckResourceAttr("circonus_rule_set.icmp-latency-alarm", "if.0.value.0.over.#", "0"),
					resource.TestCheckResourceAttr("circonus_rule_set.icmp-latency-alarm", "if.0.then.#", "1"),
					resource.TestCheckResourceAttr("circonus_rule_set.icmp-latency-alarm", "if.0.then.0.notify.#", "2"),
					resource.TestCheckResourceAttr("circonus_rule_set.icmp-latency-alarm", "if.0.then.0.severity", "1"),

					resource.TestCheckResourceAttr("circonus_rule_set.icmp-latency-alarm", "if.1.value.#", "1"),
					resource.TestCheckResourceAttr("circonus_rule_set.icmp-latency-alarm", "if.1.value.0.over.#", "1"),
					resource.TestCheckResourceAttr("circonus_rule_set.icmp-latency-alarm", "if.1.value.0.over.0.atleast", "30"),
					resource.TestCheckResourceAttr("circonus_rule_set.icmp-latency-alarm", "if.1.value.0.over.0.last", "120"),
					resource.TestCheckResourceAttr("circonus_rule_set.icmp-latency-alarm", "if.1.value.0.over.0.using", "average"),
					resource.TestCheckResourceAttr("circonus_rule_set.icmp-latency-alarm", "if.1.value.0.min_value", "2"),
					resource.TestCheckResourceAttr("circonus_rule_set.icmp-latency-alarm", "if.1.then.#", "1"),
					resource.TestCheckResourceAttr("circonus_rule_set.icmp-latency-alarm", "if.1.then.0.notify.#", "1"),
					resource.TestCheckResourceAttr("circonus_rule_set.icmp-latency-alarm", "if.1.then.0.severity", "2"),

					resource.TestCheckResourceAttr("circonus_rule_set.icmp-latency-alarm", "if.2.value.#", "1"),
					resource.TestCheckResourceAttr("circonus_rule_set.icmp-latency-alarm", "if.2.value.0.over.#", "1"),
					resource.TestCheckResourceAttr("circonus_rule_set.icmp-latency-alarm", "if.2.value.0.over.0.atleast", "30"),
					resource.TestCheckResourceAttr("circonus_rule_set.icmp-latency-alarm", "if.2.value.0.over.0.last", "180"),
					resource.TestCheckResourceAttr("circonus_rule_set.icmp-latency-alarm", "if.2.value.0.over.0.using", "average"),
					resource.TestCheckResourceAttr("circonus_rule_set.icmp-latency-alarm", "if.2.value.0.max_value", "300"),
					resource.TestCheckResourceAttr("circonus_rule_set.icmp-latency-alarm", "if.2.then.#", "1"),
					resource.TestCheckResourceAttr("circonus_rule_set.icmp-latency-alarm", "if.2.then.0.notify.#", "1"),
					resource.TestCheckResourceAttr("circonus_rule_set.icmp-latency-alarm", "if.2.then.0.severity", "3"),

					resource.TestCheckResourceAttr("circonus_rule_set.icmp-latency-alarm", "if.3.value.#", "1"),
					resource.TestCheckResourceAttr("circonus_rule_set.icmp-latency-alarm", "if.3.value.0.over.#", "0"),
					resource.TestCheckResourceAttr("circonus_rule_set.icmp-latency-alarm", "if.3.value.0.max_value", "400"),
					resource.TestCheckResourceAttr("circonus_rule_set.icmp-latency-alarm", "if.3.then.0.notify.#", "1"),
					resource.TestCheckResourceAttr("circonus_rule_set.icmp-latency-alarm", "if.3.then.0.after", "2400"),
					resource.TestCheckResourceAttr("circonus_rule_set.icmp-latency-alarm", "if.3.then.0.severity", "4"),

					resource.TestCheckResourceAttr("circonus_rule_set.icmp-latency-alarm", "if.4.value.#", "1"),
					resource.TestCheckResourceAttr("circonus_rule_set.icmp-latency-alarm", "if.4.value.0.over.#", "0"),
					resource.TestCheckResourceAttr("circonus_rule_set.icmp-latency-alarm", "if.4.value.0.max_value", "500"),
					resource.TestCheckNoResourceAttr("circonus_rule_set.icmp-latency-alarm", "if.4.then.0.notify"),
					resource.TestCheckResourceAttr("circonus_rule_set.icmp-latency-alarm", "if.4.then.0.severity", "0"),

					resource.TestCheckResourceAttr("circonus_rule_set.icmp-latency-alarm", "if.5.value.#", "1"),
					resource.TestCheckResourceAttr("circonus_rule_set.icmp-latency-alarm", "if.5.value.0.over.#", "0"),
					resource.TestCheckResourceAttr("circonus_rule_set.icmp-latency-alarm", "if.5.value.0.eq_value", "600"),
					resource.TestCheckNoResourceAttr("circonus_rule_set.icmp-latency-alarm", "if.5.then.0.notify"),
					resource.TestCheckResourceAttr("circonus_rule_set.icmp-latency-alarm", "if.5.then.0.severity", "0"),

					resource.TestCheckResourceAttr("circonus_rule_set.icmp-latency-alarm", "if.6.value.#", "1"),
					resource.TestCheckResourceAttr("circonus_rule_set.icmp-latency-alarm", "if.6.value.0.over.#", "0"),
					resource.TestCheckResourceAttr("circonus_rule_set.icmp-latency-alarm", "if.6.value.0.neq_value", "600"),
					resource.TestCheckNoResourceAttr("circonus_rule_set.icmp-latency-alarm", "if.6.then.0.notify"),
					resource.TestCheckResourceAttr("circonus_rule_set.icmp-latency-alarm", "if.6.then.0.severity", "0"),

					resource.TestCheckResourceAttr("circonus_rule_set.blank-user-json-test", "user_json", "{}"),

					resource.TestCheckResourceAttr("circonus_rule_set.circ-6825", "if.#", "3"),
				),
			},
			{
				Config: fmt.Sprintf(testAccCirconusRuleSetConfigUpdateFmt, rulesetCheckName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("circonus_rule_set.circ-6825", "check"),
					resource.TestCheckResourceAttr("circonus_rule_set.circ-6825", "metric_name", "average"),
					resource.TestCheckResourceAttr("circonus_rule_set.circ-6825", "metric_type", "numeric"),
					resource.TestCheckResourceAttr("circonus_rule_set.circ-6825", "notes", "CIRC-6825"),
					resource.TestCheckResourceAttr("circonus_rule_set.circ-6825", "if.#", "3"),

					resource.TestCheckResourceAttr("circonus_rule_set.circ-6825", "if.0.value.#", "1"),
					resource.TestCheckResourceAttr("circonus_rule_set.circ-6825", "if.0.value.0.absent", "300"),
					resource.TestCheckResourceAttr("circonus_rule_set.circ-6825", "if.0.value.0.over.#", "0"),
					resource.TestCheckResourceAttr("circonus_rule_set.circ-6825", "if.0.then.#", "1"),
					resource.TestCheckResourceAttr("circonus_rule_set.circ-6825", "if.0.then.0.notify.#", "1"),
					resource.TestCheckResourceAttr("circonus_rule_set.circ-6825", "if.0.then.0.severity", "1"),

					resource.TestCheckResourceAttr("circonus_rule_set.circ-6825", "if.1.value.#", "1"),
					resource.TestCheckResourceAttr("circonus_rule_set.circ-6825", "if.1.value.0.absent", "120"),
					resource.TestCheckResourceAttr("circonus_rule_set.circ-6825", "if.1.value.0.over.#", "0"),
					resource.TestCheckResourceAttr("circonus_rule_set.circ-6825", "if.1.then.#", "1"),
					resource.TestCheckResourceAttr("circonus_rule_set.circ-6825", "if.1.then.0.notify.#", "1"),
					resource.TestCheckResourceAttr("circonus_rule_set.circ-6825", "if.1.then.0.severity", "4"),

					resource.TestCheckResourceAttr("circonus_rule_set.circ-6825", "if.2.value.#", "1"),
					resource.TestCheckResourceAttr("circonus_rule_set.circ-6825", "if.2.value.0.over.#", "1"),
					resource.TestCheckResourceAttr("circonus_rule_set.circ-6825", "if.2.value.0.over.0.atleast", "0"),
					resource.TestCheckResourceAttr("circonus_rule_set.circ-6825", "if.2.value.0.over.0.last", "180"),
					resource.TestCheckResourceAttr("circonus_rule_set.circ-6825", "if.2.value.0.over.0.using", "average"),
					resource.TestCheckResourceAttr("circonus_rule_set.circ-6825", "if.2.value.0.max_value", "8000"),
					resource.TestCheckResourceAttr("circonus_rule_set.circ-6825", "if.2.then.#", "1"),
					resource.TestCheckResourceAttr("circonus_rule_set.circ-6825", "if.2.then.0.notify.#", "1"),
					resource.TestCheckResourceAttr("circonus_rule_set.circ-6825", "if.2.then.0.severity", "2"),
				),
			},
		},
	})
}

func testAccCheckDestroyCirconusRuleSet(s *terraform.State) error {
	ctxt := testAccProvider.Meta().(*providerContext)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "circonus_rule_set" {
			continue
		}

		cid := rs.Primary.ID
		exists, err := checkRuleSetExists(ctxt, api.CIDType(&cid))
		switch {
		case !exists:
			// noop
		case exists:
			return fmt.Errorf("rule set still exists after destroy")
		case err != nil:
			return fmt.Errorf("Error checking rule set: %v", err)
		}
	}

	return nil
}

func checkRuleSetExists(c *providerContext, ruleSetCID api.CIDType) (bool, error) {
	rs, err := c.client.FetchRuleSet(ruleSetCID)
	if err != nil {
		if strings.Contains(err.Error(), defaultCirconus404ErrorString) {
			return false, nil
		}

		return false, err
	}

	if api.CIDType(&rs.CID) == ruleSetCID {
		return true, nil
	}

	return false, nil
}

const testAccCirconusRuleSetConfigFmt = `
variable "test_tags" {
  type = list(string)
  default = [ "author:terraform", "lifecycle:unittest" ]
}

resource "circonus_check" "api_latency" {
  active = true
  name = "%s"
  period = "60s"

  collector {
    id = "/broker/1"
  }

  icmp_ping {
    count = 1
  }

  metric {
    name = "average"
    type = "numeric"
  }

  metric {
    name = "maximum"
    type = "numeric"
  }

  metric {
    name = "minimum"
    type = "numeric"
  }

  tags = "${var.test_tags}"
  target = "api.circonus.com"
}

resource "circonus_rule_set" "icmp-latency-alarm" {
  check = "${circonus_check.api_latency.checks[0]}"
  metric_name = "maximum"
  notes = <<-EOF
Simple check to create notifications based on ICMP performance.
EOF
  user_json = jsonencode({
    foo = "bar"
    baz = {
      quux = [1,2,3]
      bing = "florp"
    }
  })
  link = "https://wiki.example.org/playbook/what-to-do-when-high-latency-strikes"

  if {
    value {
      absent = "70"
    }

    then {
      notify = [
        "/contact_group/4680",
        "/contact_group/4679"
      ]
      severity = 1
    }
  }

  if {
    value {
      over {
        atleast = "30"
        last = "120"
        using = "average"
      }
      min_value = 2
    }

    then {
      notify = [ "/contact_group/4679" ]
      severity = 2
    }
  }

  if {
    value {
      over {
        atleast = "30"
        last = "180"
        using = "average"
      }

      max_value = 300
    }

    then {
      notify = [ "/contact_group/4679" ]
      severity = 3
    }
  }

  if {
    value {
      max_value = 400
    }

    then {
      notify = [ "/contact_group/4679" ]
      after = "2400"
      severity = 4
    }
  }

  if {
    value {
      max_value = 500
    }

    then {
      severity = 0
    }
  }

  if {
    value {
      eq_value = 600
    }
    then {
      severity = 0
    }
  }

  if {
    value {
      neq_value = 600
    }
    then {
      severity = 0
    }
  }
}

resource "circonus_rule_set" "blank-user-json-test" {
  check = "${circonus_check.api_latency.checks[0]}"
  metric_name = "minimum"
  notes = <<-EOF
Simple check to create notifications based on ICMP performance.
EOF
  link = "https://wiki.example.org/playbook/what-to-do-when-high-latency-strikes"

  if {
    value {
      absent = "70"
    }

    then {
      notify = [
        "/contact_group/4680",
        "/contact_group/4679"
      ]
      severity = 1
    }
  }
}

resource "circonus_rule_set" "circ-6825" {
  check = "${circonus_check.api_latency.checks[0]}"
  metric_name = "average"
  notes = <<-EOF
CIRC-6825
EOF

  if {
    value {
      absent = "300"
    }
    then {
      severity = 1
      notify = [
        "/contact_group/4680",
      ]
    }
  }
  if {
    value {
      absent = "70"
    }
    then {
      severity = 4
      notify = [
        "/contact_group/4680",
      ]
    }
  }
  if {
    value {
      max_value = "8000"
      over {
        atleast = "0"
        last    = "180"
        using   = "average"
      }
    }
    then {
      notify = [
        "/contact_group/4680",
      ]
      severity = 2
    }
  }
}
`

const testAccCirconusRuleSetConfigUpdateFmt = `
variable "test_tags" {
  type = list(string)
  default = [ "author:terraform", "lifecycle:unittest" ]
}

resource "circonus_check" "api_latency" {
  active = true
  name = "%s"
  period = "60s"

  collector {
    id = "/broker/1"
  }

  icmp_ping {
    count = 1
  }

  metric {
    name = "average"
    type = "numeric"
  }

  metric {
    name = "maximum"
    type = "numeric"
  }

  metric {
    name = "minimum"
    type = "numeric"
  }

  tags = "${var.test_tags}"
  target = "api.circonus.com"
}

resource "circonus_rule_set" "circ-6825" {
  check = "${circonus_check.api_latency.checks[0]}"
  metric_name = "average"
  notes = <<-EOF
CIRC-6825
EOF

  if {
    value {
      absent = "300"
    }
    then {
      severity = 1
      notify = [
        "/contact_group/4680",
      ]
    }
  }
  if {
    value {
      absent = "120"
    }
    then {
      severity = 4
      notify = [
        "/contact_group/4680",
      ]
    }
  }
  if {
    value {
      max_value = "8000"
      over {
        atleast = "0"
        last    = "180"
        using   = "average"
      }
    }
    then {
      notify = [
        "/contact_group/4680",
      ]
      severity = 2
    }
  }
}
`
