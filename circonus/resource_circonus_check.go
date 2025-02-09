package circonus

/*
 * Note to future readers: The `circonus_check` resource is actually a facade for
 * the check_bundle call.  check_bundle is an implementation detail that we mask
 * over and expose just a "check" even though the "check" is actually a
 * check_bundle.
 *
 * Style note: There are three directions that information flows:
 *
 * 1) Terraform Config file into API Objects.  *Attr named objects are Config or
 *    Schema attribute names.  In this file, all config constants should be
 *     named check*Attr.
 *
 * 2) API Objects into Statefile data.  api*Attr named constants are parameters
 *    that originate from the API and need to be mapped into the provider's
 *    vernacular.
 */

import (
	"context"
	"fmt"
	"time"

	api "github.com/circonus-labs/go-apiclient"
	"github.com/circonus-labs/go-apiclient/config"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	// circonus_check.* global resource attribute names.
	checkActiveAttr       = "active"
	checkCAQLAttr         = "caql"
	checkCloudWatchAttr   = "cloudwatch"
	checkCollectorAttr    = "collector"
	checkConsulAttr       = "consul"
	checkDNSAttr          = "dns"
	checkExternalAttr     = "external"
	checkHTTPAttr         = "http"
	checkHTTPTrapAttr     = "httptrap"
	checkICMPPingAttr     = "icmp_ping"
	checkJMXAttr          = "jmx"
	checkJSONAttr         = "json"
	checkMemcachedAttr    = "memcached"
	checkMetricAttr       = "metric"
	checkMetricFilterAttr = "metric_filter"
	checkMetricLimitAttr  = "metric_limit"
	checkMySQLAttr        = "mysql"
	checkNameAttr         = "name"
	checkNTPAttr          = "ntp"
	checkNotesAttr        = "notes"
	checkPeriodAttr       = "period"
	checkPostgreSQLAttr   = "postgresql"
	checkPromTextAttr     = "promtext"
	checkRedisAttr        = "redis"
	checkSMTPAttr         = "smtp"
	checkSNMPAttr         = "snmp"
	checkStatsdAttr       = "statsd"
	checkTCPAttr          = "tcp"
	checkTagsAttr         = "tags"
	checkTargetAttr       = "target"
	checkTimeoutAttr      = "timeout"
	checkTypeAttr         = "type"

	// circonus_check.collector.* resource attribute names.
	checkCollectorIDAttr = "id"

	// circonus_check.metric.* resource attribute names are aliased to
	// circonus_metric.* resource attributes.

	// circonus_check.metric.* resource attribute names
	// metricIDAttr  = "id".

	// Out parameters for circonus_check.
	checkOutByCollectorAttr        = "check_by_collector"
	checkOutIDAttr                 = "check_id"
	checkOutChecksAttr             = "checks"
	checkOutCreatedAttr            = "created"
	checkOutLastModifiedAttr       = "last_modified"
	checkOutLastModifiedByAttr     = "last_modified_by"
	checkOutReverseConnectURLsAttr = "reverse_connect_urls"
	checkOutCheckUUIDsAttr         = "uuids"
)

const (
	// Circonus API constants from their API endpoints.
	apiCheckTypeCAQLAttr       apiCheckType = "caql"
	apiCheckTypeCloudWatchAttr apiCheckType = "cloudwatch"
	apiCheckTypeConsulAttr     apiCheckType = "consul"
	apiCheckTypeDNSAttr        apiCheckType = "dns"
	apiCheckTypeExternalAttr   apiCheckType = "external"
	apiCheckTypeHTTPAttr       apiCheckType = "http"
	apiCheckTypeHTTPTrapAttr   apiCheckType = "httptrap"
	apiCheckTypeJMXAttr        apiCheckType = "jmx"
	apiCheckTypeMemcachedAttr  apiCheckType = "memcached"
	apiCheckTypeICMPPingAttr   apiCheckType = "ping_icmp"
	apiCheckTypeJSONAttr       apiCheckType = "json"
	apiCheckTypeMySQLAttr      apiCheckType = "mysql"
	apiCheckTypeNTPAttr        apiCheckType = "ntp"
	apiCheckTypePostgreSQLAttr apiCheckType = "postgres"
	apiCheckTypePromTextAttr   apiCheckType = "promtext"
	apiCheckTypeRedisAttr      apiCheckType = "redis"
	apiCheckTypeSMTPAttr       apiCheckType = "smtp"
	apiCheckTypeSNMPAttr       apiCheckType = "snmp"
	apiCheckTypeStatsdAttr     apiCheckType = "statsd"
	apiCheckTypeTCPAttr        apiCheckType = "tcp"
)

var checkDescriptions = attrDescrs{
	checkActiveAttr:       "If the check is activate or disabled",
	checkCAQLAttr:         "CAQL check configuration",
	checkCloudWatchAttr:   "CloudWatch check configuration",
	checkCollectorAttr:    "The collector(s) that are responsible for gathering the metrics",
	checkConsulAttr:       "Consul check configuration",
	checkDNSAttr:          "DNS check configuration",
	checkExternalAttr:     "External check configuration",
	checkHTTPAttr:         "HTTP check configuration",
	checkHTTPTrapAttr:     "HTTP Trap check configuration",
	checkICMPPingAttr:     "ICMP ping check configuration",
	checkJMXAttr:          "JMX check configuration",
	checkJSONAttr:         "JSON check configuration",
	checkMemcachedAttr:    "Memcached check configuration",
	checkMetricAttr:       "Configuration for a stream of metrics",
	checkMetricFilterAttr: "Allow/deny configuration for regex based metric ingestion",
	checkMetricLimitAttr:  `Setting a metric_limit will enable all (-1), disable (0), or allow up to the specified limit of metrics for this check ("N+", where N is a positive integer)`,
	checkMySQLAttr:        "MySQL check configuration",
	checkNameAttr:         "The name of the check bundle that will be displayed in the web interface",
	checkNTPAttr:          "NTP check configuration",
	checkNotesAttr:        "Notes about this check bundle",
	checkPeriodAttr:       "The period between each time the check is made",
	checkPostgreSQLAttr:   "PostgreSQL check configuration",
	checkPromTextAttr:     "Prometheus URL scraper check configuration",
	checkSMTPAttr:         "SMTP check configuration",
	checkRedisAttr:        "Redis check configuration",
	checkSNMPAttr:         "SNMP check configuration",
	checkStatsdAttr:       "statsd check configuration",
	checkTCPAttr:          "TCP check configuration",
	checkTagsAttr:         "A list of tags assigned to the check",
	checkTargetAttr:       "The target of the check (e.g. hostname, URL, IP, etc)",
	checkTimeoutAttr:      "The length of time in seconds (and fractions of a second) before the check will timeout if no response is returned to the collector",
	checkTypeAttr:         "The check type",

	checkOutByCollectorAttr:        "",
	checkOutCheckUUIDsAttr:         "",
	checkOutChecksAttr:             "",
	checkOutCreatedAttr:            "",
	checkOutIDAttr:                 "",
	checkOutLastModifiedAttr:       "",
	checkOutLastModifiedByAttr:     "",
	checkOutReverseConnectURLsAttr: "",
}

var checkCollectorDescriptions = attrDescrs{
	checkCollectorIDAttr: "The ID of the collector",
}

var (
	checkMetricDescriptions       = metricDescriptions
	checkMetricFilterDescriptions = attrDescrs{
		"type":      "'allow' or 'deny'",
		"regex":     "Regex of the filter",
		"comment":   "Comment on this filter",
		"tag_query": "The tag query to apply",
	}
)

func resourceCheck() *schema.Resource {
	return &schema.Resource{
		CreateContext: checkCreate,
		ReadContext:   checkRead,
		UpdateContext: checkUpdate,
		DeleteContext: checkDelete,
		// Exists: checkExists,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: convertToHelperSchema(checkDescriptions, map[schemaAttr]*schema.Schema{
			// Out parameters
			// _cid
			checkOutIDAttr: {
				Type:     schema.TypeString,
				Computed: true,
			},
			// _brokers
			checkOutByCollectorAttr: {
				Type:     schema.TypeMap,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			// _check_uuids
			checkOutCheckUUIDsAttr: {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			// _checks
			checkOutChecksAttr: {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			// _created
			checkOutCreatedAttr: {
				Type:     schema.TypeInt,
				Computed: true,
			},
			// _last_modified
			checkOutLastModifiedAttr: {
				Type:     schema.TypeInt,
				Computed: true,
			},
			// _last_modified_by
			checkOutLastModifiedByAttr: {
				Type:     schema.TypeString,
				Computed: true,
			},
			// _reverse_connection_urls
			checkOutReverseConnectURLsAttr: {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			// brokers
			checkCollectorAttr: {
				Type:     schema.TypeSet,
				Optional: true,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: convertToHelperSchema(checkCollectorDescriptions, map[schemaAttr]*schema.Schema{
						checkCollectorIDAttr: {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validateRegexp(checkCollectorIDAttr, config.BrokerCIDRegex),
						},
					}),
				},
			},
			// display_name
			checkNameAttr: {
				Type:     schema.TypeString,
				Optional: true,
			},
			// metric_filters
			checkMetricFilterAttr: {
				Type:     schema.TypeList, // order matters here so use a List
				Optional: true,
				MinItems: 0,
				Elem: &schema.Resource{
					Schema: convertToHelperSchema(checkMetricFilterDescriptions, map[schemaAttr]*schema.Schema{
						"type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validateRegexp("type", `allow|deny`),
						},
						"regex": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validateRegexp(metricNameAttr, `.+`),
						},
						"comment": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validateRegexp(metricNameAttr, `.+`),
						},
						"tag_query": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validateRegexp(metricNameAttr, `.+`),
						},
					}),
				},
			},
			// metric_limit
			checkMetricLimitAttr: {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
				ValidateFunc: validateFuncs(
					validateIntMin(checkMetricLimitAttr, -1),
				),
			},
			// metrics
			checkMetricAttr: {
				Type:     schema.TypeList,
				Optional: true,
				MinItems: 0,
				Elem: &schema.Resource{
					Schema: convertToHelperSchema(checkMetricDescriptions, map[schemaAttr]*schema.Schema{
						metricActiveAttr: {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
						metricNameAttr: {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validateRegexp(metricNameAttr, `[\S]+`),
						},
						metricTypeAttr: {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validateMetricType,
						},
					}),
				},
			},
			// notes
			checkNotesAttr: {
				Type:      schema.TypeString,
				Optional:  true,
				Computed:  true,
				StateFunc: suppressWhitespace,
			},
			// period
			checkPeriodAttr: {
				Type:      schema.TypeString,
				Optional:  true,
				Computed:  true,
				StateFunc: normalizeTimeDurationStringToSeconds,
				ValidateFunc: validateFuncs(
					validateDurationMin(checkPeriodAttr, defaultCirconusCheckPeriodMin),
					validateDurationMax(checkPeriodAttr, defaultCirconusCheckPeriodMax),
				),
			},
			// status
			checkActiveAttr: {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			// tags
			checkTagsAttr: tagMakeConfigSchema(checkTagsAttr),
			// target
			checkTargetAttr: {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validateRegexp(checkTargetAttr, `.+`),
			},
			// timeout
			checkTimeoutAttr: {
				Type:      schema.TypeString,
				Optional:  true,
				Computed:  true,
				StateFunc: normalizeTimeDurationStringToSeconds,
				ValidateFunc: validateFuncs(
					validateDurationMin(checkTimeoutAttr, defaultCirconusTimeoutMin),
					validateDurationMax(checkTimeoutAttr, defaultCirconusTimeoutMax),
				),
			},
			// type
			checkTypeAttr: {
				Type:         schema.TypeString,
				Computed:     true,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validateCheckType,
			},
			//
			// specific check types, their attributes go into
			// the check_bundle.config attribute
			//
			checkCAQLAttr:       schemaCheckCAQL,
			checkCloudWatchAttr: schemaCheckCloudWatch,
			checkConsulAttr:     schemaCheckConsul,
			checkDNSAttr:        schemaCheckDNS,
			checkExternalAttr:   schemaCheckExternal,
			checkHTTPAttr:       schemaCheckHTTP,
			checkHTTPTrapAttr:   schemaCheckHTTPTrap,
			checkICMPPingAttr:   schemaCheckICMPPing,
			checkJMXAttr:        schemaCheckJMX,
			checkMemcachedAttr:  schemaCheckMemcached,
			checkMySQLAttr:      schemaCheckMySQL,
			checkNTPAttr:        schemaCheckNTP,
			checkJSONAttr:       schemaCheckJSON,
			checkPostgreSQLAttr: schemaCheckPostgreSQL,
			checkPromTextAttr:   schemaCheckPromText,
			checkRedisAttr:      schemaCheckRedis,
			checkSMTPAttr:       schemaCheckSMTP,
			checkSNMPAttr:       schemaCheckSNMP,
			checkStatsdAttr:     schemaCheckStatsd,
			checkTCPAttr:        schemaCheckTCP,
		}),
	}
}

func checkCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	ctxt := meta.(*providerContext)
	c := newCheck()
	if err := c.ParseConfig(d); err != nil {
		return diag.FromErr(err)
	}

	if err := c.Create(ctxt); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(c.CID)

	return checkRead(ctx, d, meta)
}

// checkRead now covers "existence"
// func checkExists(d *schema.ResourceData, meta interface{}) (bool, error) {
// 	ctxt := meta.(*providerContext)

// 	cid := d.Id()
// 	cb, err := ctxt.client.FetchCheckBundle(api.CIDType(&cid))
// 	if err != nil {
// 		return false, err
// 	}

// 	if cb.CID == "" {
// 		return false, nil
// 	}

// 	return true, nil
// }

// checkRead pulls data out of the CheckBundle object and stores it into the
// appropriate place in the statefile.
func checkRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	ctxt := meta.(*providerContext)
	var diags diag.Diagnostics

	cid := d.Id()
	var c circonusCheck
	c, err := loadCheck(ctxt, api.CIDType(&cid))
	if err != nil {
		return diag.FromErr(err)
	}

	if c.CID == "" {
		d.SetId("")
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "Check Bundle does not exist",
			Detail:   fmt.Sprintf("Check Bundle (%q) was not found.", cid),
		})
		return diags
	}

	d.SetId(c.CID)

	// Global circonus_check attributes are saved first, followed by the check
	// type specific attributes handled below in their respective checkRead*().

	checkIDsByCollector := make(map[string]interface{}, len(c.Checks))
	for i, b := range c.Brokers {
		checkIDsByCollector[b] = c.Checks[i]
	}

	var checkID string
	if len(c.Checks) == 1 {
		checkID = c.Checks[0]
	}

	metrics := make([]interface{}, 0)
	for _, m := range c.Metrics {
		metricAttrs := map[string]interface{}{
			string(metricActiveAttr): metricAPIStatusToBool(m.Status),
			string(metricNameAttr):   m.Name,
			string(metricTypeAttr):   m.Type,
		}

		metrics = append(metrics, metricAttrs)
	}

	metricFilters := make([]interface{}, 0)
	for _, m := range c.MetricFilters {
		metricFilterAttrs := map[string]interface{}{
			"type":  m[0],
			"regex": m[1],
		}
		if m[2] == "tags" {
			metricFilterAttrs["tag_query"] = m[3]
			metricFilterAttrs["comment"] = m[4]
		} else {
			metricFilterAttrs["tag_query"] = ""
			metricFilterAttrs["comment"] = m[2]
		}

		metricFilters = append(metricFilters, metricFilterAttrs)
	}

	// Write the global circonus_check parameters followed by the check
	// type-specific parameters.

	if err := d.Set(checkActiveAttr, checkAPIStatusToBool(c.Status)); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set(checkCollectorAttr, stringListToSet(c.Brokers, checkCollectorIDAttr)); err != nil {
		return diag.FromErr(err) // fmt.Errorf("Unable to store check %q attribute: %w", checkCollectorAttr, err)
	}

	if err := d.Set(checkMetricLimitAttr, c.MetricLimit); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set(checkNameAttr, c.DisplayName); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set(checkNotesAttr, c.Notes); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set(checkPeriodAttr, fmt.Sprintf("%ds", c.Period)); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set(checkMetricAttr, metrics); err != nil {
		return diag.FromErr(err) // fmt.Errorf("Unable to store check %q attribute: %w", checkMetricAttr, err)
	}

	if err := d.Set(checkMetricFilterAttr, metricFilters); err != nil {
		return diag.FromErr(err) // fmt.Errorf("Unable to store check %q attribute: %w", checkMetricFilterAttr, err)
	}

	if err := d.Set(checkTagsAttr, c.Tags); err != nil {
		return diag.FromErr(err) // fmt.Errorf("Unable to store check %q attribute: %w", checkTagsAttr, err)
	}

	if err := d.Set(checkTargetAttr, c.Target); err != nil {
		return diag.FromErr(err)
	}

	{
		t, err := time.ParseDuration(fmt.Sprintf("%fs", c.Timeout))
		if err != nil {
			return diag.FromErr(err)
		}
		if err = d.Set(checkTimeoutAttr, t.String()); err != nil {
			return diag.FromErr(err)
		}
	}

	if err := d.Set(checkTypeAttr, c.Type); err != nil {
		return diag.FromErr(err)
	}

	// Last step: parse a check_bundle's config into the statefile.
	if err := parseCheckTypeConfig(&c, d); err != nil {
		return diag.FromErr(err) // fmt.Errorf("Unable to parse check config: %w", err)
	}

	// Out parameters
	if err := d.Set(checkOutByCollectorAttr, checkIDsByCollector); err != nil {
		return diag.FromErr(err) // fmt.Errorf("Unable to store check %q attribute: %w", checkOutByCollectorAttr, err)
	}

	if err := d.Set(checkOutCheckUUIDsAttr, c.CheckUUIDs); err != nil {
		return diag.FromErr(err) // fmt.Errorf("Unable to store check %q attribute: %w", checkOutCheckUUIDsAttr, err)
	}

	if err := d.Set(checkOutChecksAttr, c.Checks); err != nil {
		return diag.FromErr(err) // fmt.Errorf("Unable to store check %q attribute: %w", checkOutChecksAttr, err)
	}

	if checkID != "" {
		if err := d.Set(checkOutIDAttr, checkID); err != nil {
			return diag.FromErr(err)
		}
	}

	if err := d.Set(checkOutCreatedAttr, c.Created); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set(checkOutLastModifiedAttr, c.LastModified); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set(checkOutLastModifiedByAttr, c.LastModifedBy); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set(checkOutReverseConnectURLsAttr, c.ReverseConnectURLs); err != nil {
		return diag.FromErr(err) // fmt.Errorf("Unable to store check %q attribute: %w", checkOutReverseConnectURLsAttr, err)
	}

	return nil
}

func checkUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	ctxt := meta.(*providerContext)
	c := newCheck()
	if err := c.ParseConfig(d); err != nil {
		return diag.FromErr(err)
	}

	c.CID = d.Id()
	if err := c.Update(ctxt); err != nil {
		return diag.FromErr(err) // fmt.Errorf("unable to update check %q: %w", d.Id(), err)
	}

	return checkRead(ctx, d, meta)
}

func checkDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	ctxt := meta.(*providerContext)

	if _, err := ctxt.client.Delete(d.Id()); err != nil {
		return diag.FromErr(err) // fmt.Errorf("unable to delete check %q: %w", d.Id(), err)
	}

	d.SetId("")

	return nil
}

// ParseConfig reads Terraform config data and stores the information into a
// Circonus CheckBundle object.
func (c *circonusCheck) ParseConfig(d *schema.ResourceData) error {
	if v, found := d.GetOk(checkActiveAttr); found {
		c.Status = checkActiveToAPIStatus(v.(bool))
	}

	if v, found := d.GetOk(checkCollectorAttr); found {
		l := v.(*schema.Set).List()
		c.Brokers = make([]string, 0, len(l))

		for _, mapRaw := range l {
			mapAttrs := mapRaw.(map[string]interface{})

			if mv, mapFound := mapAttrs[checkCollectorIDAttr]; mapFound {
				c.Brokers = append(c.Brokers, mv.(string))
			}
		}
	}

	if v, found := d.GetOk(checkMetricLimitAttr); found {
		c.MetricLimit = v.(int)
	}

	if v, found := d.GetOk(checkNameAttr); found {
		c.DisplayName = v.(string)
	}

	if v, found := d.GetOk(checkNotesAttr); found {
		s := v.(string)
		c.Notes = &s
	}

	if v, found := d.GetOk(checkPeriodAttr); found {
		d, err := time.ParseDuration(v.(string))
		if err != nil {
			return fmt.Errorf("unable to parse %q as a duration: %w", checkPeriodAttr, err)
		}

		c.Period = uint(d.Seconds())
	}

	if v, found := d.GetOk(checkMetricAttr); found {
		metricList := v.([]interface{})
		c.Metrics = make([]api.CheckBundleMetric, 0, len(metricList))

		for _, metricListRaw := range metricList {
			metricAttrs := metricListRaw.(map[string]interface{})

			var id string
			if av, found := metricAttrs[metricIDAttr]; found {
				id = av.(string)
			} else {
				var err error
				id, err = newMetricID()
				if err != nil {
					return fmt.Errorf("unable to create a new metric ID: %w", err)
				}
			}

			m := newMetric()
			if err := m.ParseConfigMap(id, metricAttrs); err != nil {
				return fmt.Errorf("unable to parse config: %w", err)
			}

			c.Metrics = append(c.Metrics, m.CheckBundleMetric)
		}
	} else {
		c.Metrics = make([]api.CheckBundleMetric, 0)
	}

	if v, found := d.GetOk(checkMetricFilterAttr); found {
		metricFilterList := v.([]interface{})
		c.MetricFilters = make([][]string, 0, len(metricFilterList))

		for _, metricFilterListRaw := range metricFilterList {
			metricFilterAttrs := metricFilterListRaw.(map[string]interface{})

			m := make([]string, 0, 3)
			if av, found := metricFilterAttrs["type"]; found {
				m = append(m, av.(string))
			}
			if av, found := metricFilterAttrs["regex"]; found {
				m = append(m, av.(string))
			}
			if av, found := metricFilterAttrs["tag_query"]; found {
				m = append(m, "tags")
				m = append(m, av.(string))
			}

			if av, found := metricFilterAttrs["comment"]; found {
				m = append(m, av.(string))
			}
			c.MetricFilters = append(c.MetricFilters, m)
		}
	}

	if v, found := d.GetOk(checkTagsAttr); found {
		c.Tags = derefStringList(flattenSet(v.(*schema.Set)))
	}

	if v, found := d.GetOk(checkTargetAttr); found {
		c.Target = v.(string)
	}

	if v, found := d.GetOk(checkTimeoutAttr); found {
		d, err := time.ParseDuration(v.(string))
		if err != nil {
			return fmt.Errorf("unable to parse %q as a duration: %w", checkTimeoutAttr, err)
		}

		t := float32(d.Seconds())
		c.Timeout = t
	}

	// Last step: parse the individual check types
	if err := checkConfigToAPI(c, d); err != nil {
		return fmt.Errorf("unable to parse check type: %w", err)
	}

	if err := c.Fixup(); err != nil {
		return err
	}

	if err := c.Validate(); err != nil {
		return err
	}

	return nil
}

// checkConfigToAPI parses the Terraform config into the respective per-check
// type api.Config attributes.
func checkConfigToAPI(c *circonusCheck, d *schema.ResourceData) error {
	checkTypeParseMap := map[string]func(*circonusCheck, interfaceList) error{
		checkCAQLAttr:       checkConfigToAPICAQL,
		checkCloudWatchAttr: checkConfigToAPICloudWatch,
		checkConsulAttr:     checkConfigToAPIConsul,
		checkDNSAttr:        checkConfigToAPIDNS,
		checkExternalAttr:   checkConfigToAPIExternal,
		checkHTTPAttr:       checkConfigToAPIHTTP,
		checkHTTPTrapAttr:   checkConfigToAPIHTTPTrap,
		checkICMPPingAttr:   checkConfigToAPIICMPPing,
		checkJMXAttr:        checkConfigToAPIJMX,
		checkMemcachedAttr:  checkConfigToAPIMemcached,
		checkJSONAttr:       checkConfigToAPIJSON,
		checkMySQLAttr:      checkConfigToAPIMySQL,
		checkNTPAttr:        checkConfigToAPINTP,
		checkPostgreSQLAttr: checkConfigToAPIPostgreSQL,
		checkPromTextAttr:   checkConfigToAPIPromText,
		checkRedisAttr:      checkConfigToAPIRedis,
		checkSMTPAttr:       checkConfigToAPISMTP,
		checkSNMPAttr:       checkConfigToAPISNMP,
		checkStatsdAttr:     checkConfigToAPIStatsd,
		checkTCPAttr:        checkConfigToAPITCP,
	}

	for checkType, fn := range checkTypeParseMap {
		if listRaw, found := d.GetOk(checkType); found {
			switch u := listRaw.(type) {
			case []interface{}:
				if err := fn(c, u); err != nil {
					return fmt.Errorf("Unable to parse type %q: %w", checkType, err)
				}
			case *schema.Set:
				if err := fn(c, u.List()); err != nil {
					return fmt.Errorf("Unable to parse type %q: %w", checkType, err)
				}
			default:
				return fmt.Errorf("PROVIDER BUG: unsupported check type interface: %q", checkType)
			}
		}
	}

	return nil
}

// parseCheckTypeConfig parses an API Config object and stores the result in the
// statefile.
func parseCheckTypeConfig(c *circonusCheck, d *schema.ResourceData) error {
	checkTypeConfigHandlers := map[apiCheckType]func(*circonusCheck, *schema.ResourceData) error{
		apiCheckTypeCAQLAttr:       checkAPIToStateCAQL,
		apiCheckTypeCloudWatchAttr: checkAPIToStateCloudWatch,
		apiCheckTypeConsulAttr:     checkAPIToStateConsul,
		apiCheckTypeDNSAttr:        checkAPIToStateDNS,
		apiCheckTypeExternalAttr:   checkAPIToStateExternal,
		apiCheckTypeHTTPAttr:       checkAPIToStateHTTP,
		apiCheckTypeHTTPTrapAttr:   checkAPIToStateHTTPTrap,
		apiCheckTypeICMPPingAttr:   checkAPIToStateICMPPing,
		apiCheckTypeJMXAttr:        checkAPIToStateJMX,
		apiCheckTypeMemcachedAttr:  checkAPIToStateMemcached,
		apiCheckTypeJSONAttr:       checkAPIToStateJSON,
		apiCheckTypeMySQLAttr:      checkAPIToStateMySQL,
		apiCheckTypeNTPAttr:        checkAPIToStateNTP,
		apiCheckTypePostgreSQLAttr: checkAPIToStatePostgreSQL,
		apiCheckTypePromTextAttr:   checkAPIToStatePromText,
		apiCheckTypeRedisAttr:      checkAPIToStateRedis,
		apiCheckTypeSMTPAttr:       checkAPIToStateSMTP,
		apiCheckTypeSNMPAttr:       checkAPIToStateSNMP,
		apiCheckTypeStatsdAttr:     checkAPIToStateStatsd,
		apiCheckTypeTCPAttr:        checkAPIToStateTCP,
	}

	var checkType apiCheckType = apiCheckType(c.Type)
	fn, ok := checkTypeConfigHandlers[checkType]
	if !ok {
		return fmt.Errorf("check type %q not supported", c.Type)
	}

	if err := fn(c, d); err != nil {
		return fmt.Errorf("unable to parse the API config for %q: %w", c.Type, err)
	}

	return nil
}
