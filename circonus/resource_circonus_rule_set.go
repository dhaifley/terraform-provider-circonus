package circonus

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strconv"
	"time"

	api "github.com/circonus-labs/go-apiclient"
	"github.com/circonus-labs/go-apiclient/config"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	// circonus_rule_set.* resource attribute names.
	ruleSetCheckAttr         = "check"
	ruleSetNameAttr          = "name"
	ruleSetIfAttr            = "if"
	ruleSetLinkAttr          = "link"
	ruleSetMetricTypeAttr    = "metric_type"
	ruleSetNotesAttr         = "notes"
	ruleSetUserJSONAttr      = "user_json"
	ruleSetParentAttr        = "parent"
	ruleSetMetricNameAttr    = "metric_name"
	ruleSetMetricPatternAttr = "metric_pattern"
	ruleSetMetricFilterAttr  = "metric_filter"
	ruleSetTagsAttr          = "tags"

	// circonus_rule_set.if.* resource attribute names.
	ruleSetThenAttr  = "then"
	ruleSetValueAttr = "value"

	// circonus_rule_set.if.then.* resource attribute names.
	ruleSetAfterAttr    = "after"
	ruleSetNotifyAttr   = "notify"
	ruleSetSeverityAttr = "severity"

	// circonus_rule_set.if.value.* resource attribute names.
	ruleSetAbsentAttr     = "absent"      // apiRuleSetAbsent
	ruleSetChangedAttr    = "changed"     // apiRuleSetChanged
	ruleSetContainsAttr   = "contains"    // apiRuleSetContains
	ruleSetMatchAttr      = "match"       // apiRuleSetMatch
	ruleSetMaxValueAttr   = "max_value"   // apiRuleSetMaxValue
	ruleSetMinValueAttr   = "min_value"   // apiRuleSetMinValue
	ruleSetEqValueAttr    = "eq_value"    // apiRuleSetEqValue
	ruleSetNotEqValueAttr = "neq_value"   // apiRuleSetNotEqValue
	ruleSetNotContainAttr = "not_contain" // apiRuleSetNotContains
	ruleSetNotMatchAttr   = "not_match"   // apiRuleSetNotMatch
	ruleSetOverAttr       = "over"

	// circonus_rule_set.if.value.over.* resource attribute names.
	ruleSetLastAttr    = "last"
	ruleSetUsingAttr   = "using"
	ruleSetAtLeastAttr = "atleast"

	// out attributes.
	ruleSetIDAttr = "rule_set_id"
)

const (
	// Different criteria that an api.RuleSetRule can return.
	apiRuleSetAbsent      = "on absence"       // ruleSetAbsentAttr
	apiRuleSetChanged     = "on change"        // ruleSetChangedAttr
	apiRuleSetContains    = "contains"         // ruleSetContainsAttr
	apiRuleSetMatch       = "match"            // ruleSetMatchAttr
	apiRuleSetMaxValue    = "max value"        // ruleSetMaxValueAttr
	apiRuleSetMinValue    = "min value"        // ruleSetMinValueAttr
	apiRuleSetNotContains = "does not contain" // ruleSetNotContainAttr
	apiRuleSetNotMatch    = "does not match"   // ruleSetNotMatchAttr
	apiRuleSetEqValue     = "equals"           // ruleSetEqValueAttr
	apiRuleSetNotEqValue  = "does not equal"   // ruleSetNotEqValueAttr
)

var ruleSetDescriptions = attrDescrs{
	// circonus_rule_set.* resource attribute names
	ruleSetCheckAttr:         "The CID of the check that contains the metric for this rule set",
	ruleSetNameAttr:          "The name of this ruleset, if omitted will default to the metric_name (or pattern) and filter",
	ruleSetIfAttr:            "A rule to execute for this rule set",
	ruleSetLinkAttr:          "URL to show users when this rule set is active (e.g. wiki)",
	ruleSetMetricTypeAttr:    "The type of data flowing through the specified metric stream",
	ruleSetNotesAttr:         "Notes describing this rule set",
	ruleSetUserJSONAttr:      "Opaque data that can be supplied with the result and appears in webhooks when alerts go off",
	ruleSetParentAttr:        "Parent CID that must be healthy for this rule set to be active",
	ruleSetMetricNameAttr:    "The name of the metric stream within a check to register the rule set with",
	ruleSetMetricPatternAttr: "The pattern match (regex) of the metric stream within a check to register the rule set with",
	ruleSetMetricFilterAttr:  "The tag filter a pattern match ruleset will user",
	ruleSetTagsAttr:          "Tags associated with this rule set",
	ruleSetIDAttr:            "out",
}

var ruleSetIfDescriptions = attrDescrs{
	// circonus_rule_set.if.* resource attribute names
	ruleSetThenAttr:  "Description of the action(s) to take when this rule set is active",
	ruleSetValueAttr: "Predicate that the rule set uses to evaluate a stream of metrics",
}

var ruleSetIfValueDescriptions = attrDescrs{
	// circonus_rule_set.if.value.* resource attribute names
	ruleSetAbsentAttr:     "Fire the rule set if there has been no data for the given metric stream over the last duration",
	ruleSetChangedAttr:    "Boolean indicating the value has changed",
	ruleSetContainsAttr:   "Fire the rule set if the text metric contain the following string",
	ruleSetMatchAttr:      "Fire the rule set if the text metric exactly match the following string",
	ruleSetNotMatchAttr:   "Fire the rule set if the text metric not match the following string",
	ruleSetMinValueAttr:   "Fire the rule set if the numeric value less than the specified value",
	ruleSetEqValueAttr:    "Fire the rule set if the numeric value equals the specified value",
	ruleSetNotEqValueAttr: "Fire the rule set if the numeric value does not equal the specified value",
	ruleSetNotContainAttr: "Fire the rule set if the text metric does not contain the following string",
	ruleSetMaxValueAttr:   "Fire the rule set if the numeric value is more than the specified value",
	ruleSetOverAttr:       "Use a derived value using a window",
	ruleSetThenAttr:       "Action to take when the rule set is active",
}

var ruleSetIfValueOverDescriptions = attrDescrs{
	// circonus_rule_set.if.value.over.* resource attribute names
	ruleSetLastAttr:    "Duration over which data from the last interval is examined",
	ruleSetAtLeastAttr: "Wait at least this long (seconds) before evaluating the rule",
	ruleSetUsingAttr:   "Define the window function to use over the last duration",
}

var ruleSetIfThenDescriptions = attrDescrs{
	// circonus_rule_set.if.then.* resource attribute names
	ruleSetAfterAttr:    "The length of time we should wait before contacting the contact groups after this ruleset has faulted.",
	ruleSetNotifyAttr:   "List of contact groups to notify at the following appropriate severity if this rule set is active.",
	ruleSetSeverityAttr: "Send a notification at this severity level.",
}

func resourceRuleSet() *schema.Resource {
	/*
		makeConflictsWith := func(in ...schemaAttr) []string {
			out := make([]string, 0, len(in))
			for _, attr := range in {
				out = append(out, string(ruleSetIfAttr)+"."+string(ruleSetValueAttr)+"."+string(attr))
			}
			return out
		}
	*/

	return &schema.Resource{
		CreateContext: ruleSetCreate,
		ReadContext:   ruleSetRead,
		UpdateContext: ruleSetUpdate,
		DeleteContext: ruleSetDelete,
		Importer: &schema.ResourceImporter{
			State: importStatePassthroughUnescape,
		},
		Schema: convertToHelperSchema(ruleSetDescriptions, map[schemaAttr]*schema.Schema{
			// _cid
			ruleSetIDAttr: {
				Type:     schema.TypeString,
				Computed: true,
			},
			// check
			ruleSetCheckAttr: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateRegexp(ruleSetCheckAttr, config.CheckCIDRegex),
			},
			// name
			ruleSetNameAttr: {
				Type:     schema.TypeString,
				Optional: true,
			},
			// rules
			ruleSetIfAttr: {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: convertToHelperSchema(ruleSetIfDescriptions, map[schemaAttr]*schema.Schema{
						ruleSetThenAttr: {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							Elem: &schema.Resource{
								Schema: convertToHelperSchema(ruleSetIfThenDescriptions, map[schemaAttr]*schema.Schema{
									ruleSetAfterAttr: {
										Type:         schema.TypeString,
										Optional:     true,
										Default:      "0",
										ValidateFunc: validateRegexp(ruleSetAfterAttr, "^[0-9]+$"),
									},
									ruleSetNotifyAttr: {
										Type:     schema.TypeSet,
										Optional: true,
										MinItems: 0,
										Elem: &schema.Schema{
											Type:         schema.TypeString,
											ValidateFunc: validateContactGroupCID(ruleSetNotifyAttr),
										},
									},
									ruleSetSeverityAttr: {
										Type:     schema.TypeInt,
										Optional: true,
										Default:  defaultAlertSeverity,
										ValidateFunc: validateFuncs(
											validateIntMax(ruleSetSeverityAttr, maxSeverity),
											validateIntMin(ruleSetSeverityAttr, minSeverity),
										),
									},
								}),
							},
						},
						ruleSetValueAttr: {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							Elem: &schema.Resource{
								Schema: convertToHelperSchema(ruleSetIfValueDescriptions, map[schemaAttr]*schema.Schema{
									ruleSetAbsentAttr: {
										Type:         schema.TypeString, // Applies to text or numeric metrics
										Optional:     true,
										ValidateFunc: validateRegexp(ruleSetAbsentAttr, "^[0-9]+$"),
									},
									ruleSetChangedAttr: {
										Type:     schema.TypeString, // Applies to text or numeric metrics
										Optional: true,
									},
									ruleSetContainsAttr: {
										Type:         schema.TypeString, // Applies to text metrics only
										Optional:     true,
										ValidateFunc: validateRegexp(ruleSetContainsAttr, `.+`),
									},
									ruleSetMatchAttr: {
										Type:         schema.TypeString, // Applies to text metrics only
										Optional:     true,
										ValidateFunc: validateRegexp(ruleSetMatchAttr, `.+`),
									},
									ruleSetNotMatchAttr: {
										Type:         schema.TypeString, // Applies to text metrics only
										Optional:     true,
										ValidateFunc: validateRegexp(ruleSetNotMatchAttr, `.+`),
									},
									ruleSetMinValueAttr: {
										Type:         schema.TypeString, // Applies to numeric metrics only
										Optional:     true,
										ValidateFunc: validateRegexp(ruleSetMinValueAttr, `.+`), // TODO(sean): improve this regexp to match int and float
									},
									ruleSetNotContainAttr: {
										Type:         schema.TypeString, // Applies to text metrics only
										Optional:     true,
										ValidateFunc: validateRegexp(ruleSetNotContainAttr, `.+`),
									},
									ruleSetMaxValueAttr: {
										Type:         schema.TypeString, // Applies to numeric metrics only
										Optional:     true,
										ValidateFunc: validateRegexp(ruleSetMaxValueAttr, `.+`), // TODO(sean): improve this regexp to match int and float
									},
									ruleSetEqValueAttr: {
										Type:         schema.TypeString, // Applies to numeric metrics only
										Optional:     true,
										ValidateFunc: validateRegexp(ruleSetEqValueAttr, `.+`), // TODO(sean): improve this regexp to match int and float
									},
									ruleSetNotEqValueAttr: {
										Type:         schema.TypeString, // Applies to numeric metrics only
										Optional:     true,
										ValidateFunc: validateRegexp(ruleSetNotEqValueAttr, `.+`), // TODO(sean): improve this regexp to match int and float
									},
									// windowing
									ruleSetOverAttr: {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: convertToHelperSchema(ruleSetIfValueOverDescriptions, map[schemaAttr]*schema.Schema{
												// window_duration
												ruleSetLastAttr: {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validateRegexp(ruleSetLastAttr, "^[0-9]+$"),
												},
												// window_min_duration
												ruleSetAtLeastAttr: {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validateRegexp(ruleSetAtLeastAttr, "^[0-9]+$"),
												},
												// window_function
												ruleSetUsingAttr: {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validateStringIn(ruleSetUsingAttr, validRuleSetWindowFuncs),
												},
											}),
										},
									},
								}),
							},
						},
					}),
				},
			},
			// link
			ruleSetLinkAttr: {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validateHTTPURL(ruleSetLinkAttr, urlIsAbs|urlOptional),
			},
			// metric_type
			ruleSetMetricTypeAttr: {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      defaultRuleSetMetricType,
				ValidateFunc: validateStringIn(ruleSetMetricTypeAttr, validRuleSetMetricTypes),
			},
			// notes
			ruleSetNotesAttr: {
				Type:      schema.TypeString,
				Optional:  true,
				Computed:  true,
				StateFunc: suppressWhitespace,
			},
			// user_json
			ruleSetUserJSONAttr: {
				Type:      schema.TypeString,
				Optional:  true,
				StateFunc: jsonSort,
				DiffSuppressFunc: func(k, old, update string, d *schema.ResourceData) bool {
					log.Printf("Old: '%s', New: '%s'", old, update)

					var ifce interface{}
					ob := []byte(old)
					err := json.Unmarshal(ob, &ifce)
					if err != nil {
						return false
					}
					os, err := json.Marshal(ifce)
					if err != nil {
						return false
					}

					nb := []byte(update)
					err = json.Unmarshal(nb, &ifce)
					if err != nil {
						return false
					}
					ns, err := json.Marshal(ifce)
					if err != nil {
						return false
					}

					log.Printf("Old reformatted: '%s', New reformatted: '%s'", os, ns)

					return string(os) == string(ns)
				},

				Default: "{}",
			},
			// parent
			ruleSetParentAttr: {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				StateFunc:    suppressWhitespace,
				ValidateFunc: validateRegexp(ruleSetParentAttr, `^([\d]+(_[\d\w]+)?)|(\/rule_set\/[\d]+)$`),
			},
			// metric_name
			ruleSetMetricNameAttr: {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validateRegexp(ruleSetMetricNameAttr, `^.+$`),
			},
			// metric_pattern
			ruleSetMetricPatternAttr: {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validateRegexp(ruleSetMetricPatternAttr, `^.+$`),
			},
			// filter
			ruleSetMetricFilterAttr: {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validateRegexp(ruleSetMetricFilterAttr, `^.+$`),
			},
			// tags
			ruleSetTagsAttr: {
				Type:       schema.TypeSet,
				Deprecated: "tags on rule_sets are ignored and dropped. tags returned with rule_sets are check tags.",
				Optional:   true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validateTag,
				},
			},
		}),
	}
}

func ruleSetCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	ctxt := meta.(*providerContext)
	rs := newRuleSet()

	if err := rs.ParseConfig(d); err != nil {
		return diag.FromErr(err)
	}

	if err := rs.Create(ctxt); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(rs.CID)

	return ruleSetRead(ctx, d, meta)
}

// func ruleSetExists(d *schema.ResourceData, meta interface{}) (bool, error) {
// 	ctxt := meta.(*providerContext)

// 	cid := d.Id()
// 	rs, err := ctxt.client.FetchRuleSet(api.CIDType(&cid))
// 	if err != nil {
// 		return false, err
// 	}

// 	if rs.CID == "" {
// 		return false, nil
// 	}

// 	return true, nil
// }

// ruleSetRead pulls data out of the RuleSet object and stores it into the
// appropriate place in the statefile.
func ruleSetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*providerContext).client
	var diags diag.Diagnostics

	cid := d.Id()
	var rs circonusRuleSet
	crs, err := client.FetchRuleSet(api.CIDType(&cid))
	if err != nil {
		return diag.FromErr(err)
	}
	rs.RuleSet = *crs

	if rs.CID == "" {
		d.SetId("")
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "Rule set does not exist",
			Detail:   fmt.Sprintf("Rule set (%q) was not found.", cid),
		})
		return diags
	}

	d.SetId(rs.CID)
	if err = d.Set(ruleSetIDAttr, rs.CID); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set(ruleSetNameAttr, rs.Name); err != nil {
		return diag.FromErr(err)
	}

	ifRules := make([]interface{}, 0, defaultRuleSetRuleLen)
	for _, rule := range rs.Rules {
		ifAttrs := make(map[string]interface{}, 2)
		valueAttrs := make(map[string]interface{}, 2)
		valueOverAttrs := make(map[string]interface{}, 2)
		thenAttrs := make(map[string]interface{}, 3)

		switch rule.Criteria {
		case apiRuleSetAbsent:
			switch v := rule.Value.(type) {
			case string:
				valueAttrs[string(ruleSetAbsentAttr)] = v
			case float64:
				d, _ := time.ParseDuration(fmt.Sprintf("%fs", v))
				valueAttrs[string(ruleSetAbsentAttr)] = fmt.Sprintf("%d", int(d.Seconds()))
			default:
				valueAttrs[string(ruleSetAbsentAttr)] = fmt.Sprintf("%v", v)
			}
		case apiRuleSetChanged:
			valueAttrs[string(ruleSetChangedAttr)] = "true"
		case apiRuleSetContains:
			valueAttrs[string(ruleSetContainsAttr)] = rule.Value
		case apiRuleSetMatch:
			valueAttrs[string(ruleSetMatchAttr)] = rule.Value
		case apiRuleSetMaxValue:
			valueAttrs[string(ruleSetMaxValueAttr)] = rule.Value
		case apiRuleSetMinValue:
			valueAttrs[string(ruleSetMinValueAttr)] = rule.Value
		case apiRuleSetEqValue:
			valueAttrs[string(ruleSetEqValueAttr)] = rule.Value
		case apiRuleSetNotEqValue:
			valueAttrs[string(ruleSetNotEqValueAttr)] = rule.Value
		case apiRuleSetNotContains:
			valueAttrs[string(ruleSetNotContainAttr)] = rule.Value
		case apiRuleSetNotMatch:
			valueAttrs[string(ruleSetNotMatchAttr)] = rule.Value
		default:
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unsupported criteria",
				Detail:   fmt.Sprintf("Unable to add rule, unknown/unsupported criteria: %q", rule.Criteria),
			})
			return diags
		}

		thenAttrs[string(ruleSetAfterAttr)] = fmt.Sprintf("%d", 60*rule.Wait)
		thenAttrs[string(ruleSetSeverityAttr)] = int(rule.Severity)
		if int(rule.Severity) > 0 {
			if contactGroups, ok := rs.ContactGroups[uint8(rule.Severity)]; ok {
				sort.Strings(contactGroups)
				thenAttrs[string(ruleSetNotifyAttr)] = contactGroups
			} else {
				thenAttrs[string(ruleSetNotifyAttr)] = make([]string, 0)
			}
		}
		thenSet := make([]interface{}, 0)
		thenSet = append(thenSet, thenAttrs)
		ifAttrs[string(ruleSetThenAttr)] = thenSet

		if rule.WindowingFunction != nil {
			valueOverAttrs[string(ruleSetUsingAttr)] = *rule.WindowingFunction
			// NOTE: Only save the window duration if a function was specified
			valueOverAttrs[string(ruleSetLastAttr)] = fmt.Sprintf("%d", rule.WindowingDuration)
			valueOverAttrs[string(ruleSetAtLeastAttr)] = fmt.Sprintf("%d", rule.WindowingMinDuration)
			valueOverSet := make([]interface{}, 0)
			valueOverSet = append(valueOverSet, valueOverAttrs)
			valueAttrs[string(ruleSetOverAttr)] = valueOverSet
		}

		valueSet := make([]interface{}, 0)
		valueSet = append(valueSet, valueAttrs)
		ifAttrs[string(ruleSetValueAttr)] = valueSet

		ifRules = append(ifRules, ifAttrs)
	}

	if err = d.Set(ruleSetCheckAttr, rs.CheckCID); err != nil {
		return diag.FromErr(err)
	}

	if err = d.Set(ruleSetIfAttr, ifRules); err != nil {
		s, _ := json.MarshalIndent(ifRules, "", "  ")
		log.Printf("%s", s)
		return diag.FromErr(err)
	}

	if err = d.Set(ruleSetLinkAttr, indirect(rs.Link)); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set(ruleSetMetricNameAttr, rs.MetricName); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set(ruleSetMetricPatternAttr, rs.MetricPattern); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set(ruleSetMetricFilterAttr, rs.Filter); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set(ruleSetMetricTypeAttr, rs.MetricType); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set(ruleSetNotesAttr, indirect(rs.Notes)); err != nil {
		return diag.FromErr(err)
	}

	j, err := rs.UserJSON.MarshalJSON()
	rj := json.RawMessage(string(j))
	log.Printf("%s", string(rj))
	if err == nil {
		_ = d.Set(ruleSetUserJSONAttr, string(rj))
	} else {
		_ = d.Set(ruleSetUserJSONAttr, "{}")
	}
	_ = d.Set(ruleSetParentAttr, indirect(rs.Parent))

	// if err := d.Set(ruleSetTagsAttr, tagsToState(apiToTags(rs.Tags))); err != nil {
	// 	return fmt.Errorf("Unable to store rule set %q attribute: %w", ruleSetTagsAttr, err)
	// }

	return diags
}

func ruleSetUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	ctxt := meta.(*providerContext)
	rs := newRuleSet()

	if err := rs.ParseConfig(d); err != nil {
		return diag.FromErr(err)
	}

	rs.CID = d.Id()

	if err := rs.Update(ctxt); err != nil {
		diag.FromErr(err)
	}

	return ruleSetRead(ctx, d, meta)
}

func ruleSetDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	ctxt := meta.(*providerContext)
	var diags diag.Diagnostics

	cid := d.Id()
	if _, err := ctxt.client.DeleteRuleSetByCID(api.CIDType(&cid)); err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	_ = d.Set(ruleSetIDAttr, "")

	return diags
}

type circonusRuleSet struct {
	api.RuleSet
}

func newRuleSet() circonusRuleSet {
	rs := circonusRuleSet{
		RuleSet: *api.NewRuleSet(),
	}

	rs.ContactGroups = make(map[uint8][]string, config.NumSeverityLevels)
	for i := uint8(0); i < config.NumSeverityLevels; i++ {
		rs.ContactGroups[i+1] = make([]string, 0, 1)
	}

	rs.Rules = make([]api.RuleSetRule, 0)

	return rs
}

// func loadRuleSet(ctxt *providerContext, cid api.CIDType) (circonusRuleSet, error) {
// 	var rs circonusRuleSet
// 	crs, err := ctxt.client.FetchRuleSet(cid)
// 	if err != nil {
// 		return circonusRuleSet{}, err
// 	}
// 	rs.RuleSet = *crs

// 	return rs, nil
// }

// ParseConfig reads Terraform config data and stores the information into a
// Circonus RuleSet object.  ParseConfig and ruleSetRead()
// must be kept in sync.
func (rs *circonusRuleSet) ParseConfig(d *schema.ResourceData) error {
	if v, found := d.GetOk(ruleSetCheckAttr); found {
		rs.CheckCID = v.(string)
	}

	if v, found := d.GetOk(ruleSetNameAttr); found {
		rs.Name = v.(string)
	}

	if v, found := d.GetOk(ruleSetLinkAttr); found {
		s := v.(string)
		rs.Link = &s
	}

	if v, found := d.GetOk(ruleSetMetricTypeAttr); found {
		rs.MetricType = v.(string)
	}

	if v, found := d.GetOk(ruleSetNotesAttr); found {
		s := v.(string)
		rs.Notes = &s
	}

	if v, found := d.GetOk(ruleSetUserJSONAttr); found {
		rs.UserJSON = json.RawMessage(v.(string))
	}

	if v, found := d.GetOk(ruleSetParentAttr); found {
		s := v.(string)
		rs.Parent = &s
	}

	if v, found := d.GetOk(ruleSetMetricNameAttr); found {
		rs.MetricName = v.(string)
	}

	if v, found := d.GetOk(ruleSetMetricPatternAttr); found {
		rs.MetricPattern = v.(string)
	}

	if v, found := d.GetOk(ruleSetMetricFilterAttr); found {
		rs.Filter = v.(string)
	}

	rs.Rules = make([]api.RuleSetRule, 0)
	if ifListRaw, found := d.GetOk(ruleSetIfAttr); found {
		ifList := ifListRaw.([]interface{})
		for _, ifListElem := range ifList {
			ifAttrs := ifListElem.(map[string]interface{})

			rule := api.RuleSetRule{}
			rule.WindowingFunction = nil
			rule.Wait = 0

			if thenListRaw, found := ifAttrs[ruleSetThenAttr]; found {
				thenList := thenListRaw.([]interface{})

				for _, thenListRaw := range thenList {
					thenAttrs := thenListRaw.(map[string]interface{})

					if v, found := thenAttrs[ruleSetAfterAttr]; found {
						s := v.(string)
						if s != "" {
							d, err := time.ParseDuration(v.(string) + "s")
							if err != nil {
								return fmt.Errorf("unable to parse %q duration %q: %w", ruleSetAfterAttr, v.(string), err)
							}
							rule.Wait = uint(d.Minutes())
						}
					}

					// NOTE: break from convention of alpha sorting attributes and handle Notify after Severity

					if i, found := thenAttrs[ruleSetSeverityAttr]; found {
						rule.Severity = uint(i.(int))
					}

					if notifyListRaw, found := thenAttrs[ruleSetNotifyAttr]; found {
						notifyList := notifyListRaw.(*schema.Set).List()

						sev := uint8(rule.Severity)
						for _, contactGroupCID := range notifyList {
							var found bool
							if contactGroups, ok := rs.ContactGroups[sev]; ok {
								for _, contactGroup := range contactGroups {
									if contactGroup == contactGroupCID {
										found = true
										break
									}
								}
							}
							if !found {
								rs.ContactGroups[sev] = append(rs.ContactGroups[sev], contactGroupCID.(string))
							}
						}
					}
				}
			}

			if ruleSetValueListRaw, found := ifAttrs[ruleSetValueAttr]; found {
				ruleSetValueList := ruleSetValueListRaw.([]interface{})
				vr := ruleSetValueList[0]
				valueAttrs := vr.(map[string]interface{})

				switch rs.MetricType {
				case ruleSetMetricTypeNumeric:
					if v, found := valueAttrs[ruleSetAbsentAttr]; found && v.(string) != "" {
						s := v.(string)
						if s != "" {
							d, _ := time.ParseDuration(s + "s")
							rule.Criteria = apiRuleSetAbsent
							rule.Value = d.Seconds()
						}
					} else if v, found := valueAttrs[ruleSetChangedAttr]; found && v.(string) != "" {
						b := v.(string)
						if b == "true" {
							rule.Criteria = apiRuleSetChanged
						}
					} else if v, found := valueAttrs[ruleSetMinValueAttr]; found && v.(string) != "" {
						s := v.(string)
						if s != "" {
							rule.Criteria = apiRuleSetMinValue
							rule.Value = s
						}
					} else if v, found := valueAttrs[ruleSetMaxValueAttr]; found && v.(string) != "" {
						s := v.(string)
						if s != "" {
							rule.Criteria = apiRuleSetMaxValue
							rule.Value = s
						}
					} else if v, found := valueAttrs[ruleSetEqValueAttr]; found && v.(string) != "" {
						s := v.(string)
						if s != "" {
							rule.Criteria = apiRuleSetEqValue
							rule.Value = s
						}
					} else if v, found := valueAttrs[ruleSetNotEqValueAttr]; found && v.(string) != "" {
						s := v.(string)
						if s != "" {
							rule.Criteria = apiRuleSetNotEqValue
							rule.Value = s
						}
					}
				case ruleSetMetricTypeText:
					if v, found := valueAttrs[ruleSetAbsentAttr]; found && v.(string) != "" {
						s := v.(string)
						if s != "" {
							d, _ := time.ParseDuration(s + "s")
							rule.Criteria = apiRuleSetAbsent
							rule.Value = d.Seconds()
						}
					} else if v, found := valueAttrs[ruleSetChangedAttr]; found && v.(string) != "" {
						b := v.(string)
						if b == "true" {
							rule.Criteria = apiRuleSetChanged
						}
					} else if v, found := valueAttrs[ruleSetContainsAttr]; found && v.(string) != "" {
						s := v.(string)
						if s != "" {
							rule.Criteria = apiRuleSetContains
							rule.Value = s
						}
					} else if v, found := valueAttrs[ruleSetMatchAttr]; found && v.(string) != "" {
						s := v.(string)
						if s != "" {
							rule.Criteria = apiRuleSetMatch
							rule.Value = s
						}
					} else if v, found := valueAttrs[ruleSetNotMatchAttr]; found && v.(string) != "" {
						s := v.(string)
						if s != "" {
							rule.Criteria = apiRuleSetNotMatch
							rule.Value = s
						}
					} else if v, found := valueAttrs[ruleSetNotContainAttr]; found && v.(string) != "" {
						s := v.(string)
						if s != "" {
							rule.Criteria = apiRuleSetNotContains
							rule.Value = s
						}
					}
				default:
					return fmt.Errorf("PROVIDER BUG: unsupported rule set metric type: %q", rs.MetricType)
				}

				if ruleSetOverListRaw, found := valueAttrs[ruleSetOverAttr]; found {
					overList := ruleSetOverListRaw.([]interface{})
					for _, overListRaw := range overList {
						overAttrs := overListRaw.(map[string]interface{})

						windowDuration := uint(0)
						windowMinDuration := uint(0)
						windowFunction := ""

						if v, found := overAttrs[ruleSetLastAttr]; found && v != "" {
							i, err := strconv.Atoi(v.(string))
							if err != nil {
								return fmt.Errorf("unable to parse %q duration %q: %w", ruleSetLastAttr, v.(string), err)
							}
							windowDuration = uint(i)
						}
						if v, found := overAttrs[ruleSetAtLeastAttr]; found && v != "" {
							i, err := strconv.Atoi(v.(string))
							if err != nil {
								return fmt.Errorf("unable to parse %q duration %q: %w", ruleSetAtLeastAttr, v.(string), err)
							}
							windowMinDuration = uint(i)
						}

						if v, found := overAttrs[ruleSetUsingAttr]; found {
							windowFunction = v.(string)
						}

						if windowFunction != "" && windowDuration > 0 {
							rule.WindowingFunction = &windowFunction
							rule.WindowingDuration = windowDuration
							if windowMinDuration > 0 {
								rule.WindowingMinDuration = windowMinDuration
							}
						}
					}
				}
			}
			if rule.Criteria != "" {
				rs.Rules = append(rs.Rules, rule)
			}
		}
	}

	// if v, found := d.GetOk(ruleSetTagsAttr); found {
	// 	rs.Tags = derefStringList(flattenSet(v.(*schema.Set)))
	// }

	if err := rs.Validate(); err != nil {
		return err
	}

	return nil
}

func (rs *circonusRuleSet) Create(ctxt *providerContext) error {
	crs, err := ctxt.client.CreateRuleSet(&rs.RuleSet)
	if err != nil {
		return err
	}

	rs.CID = crs.CID

	return nil
}

func (rs *circonusRuleSet) Update(ctxt *providerContext) error {
	_, err := ctxt.client.UpdateRuleSet(&rs.RuleSet)
	if err != nil {
		return fmt.Errorf("Unable to update rule set %s: %w", rs.CID, err)
	}

	return nil
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func (rs *circonusRuleSet) Validate() error {
	// TODO(sean@): From https://login.circonus.com/resources/api/calls/rule_set
	// under `value`:
	//
	// For an 'on absence' rule this is the number of seconds the metric must not
	// have been collected for, and should not be lower than either the period or
	// timeout of the metric being collected.

	if len(rs.MetricName) > 0 && len(rs.MetricPattern) > 0 {
		return fmt.Errorf("RuleSet for check ID %s has both metric_name and metric_pattern, must be one or the other", rs.CheckCID)
	}

	if len(rs.MetricName) == 0 && len(rs.MetricPattern) == 0 {
		return fmt.Errorf("RuleSet for check ID %s must supply either metric_name or metric_pattern", rs.CheckCID)
	}

	for i, rule := range rs.Rules {
		if rule.Criteria == "" {
			return fmt.Errorf("rule %d for check ID %s has an empty criteria", i, rs.CheckCID)
		}

		if rule.WindowingMinDuration > rule.WindowingDuration {
			return fmt.Errorf("rule %d for check ID %s cannot have a window_min_duration (atleast) greater than the window duration (last)", i, rs.CheckCID)
		}

		if stringInSlice(rule.Criteria, []string{apiRuleSetMatch, apiRuleSetNotMatch, apiRuleSetContains, apiRuleSetNotContains}) {
			if rs.MetricType != "text" {
				return fmt.Errorf("rule %d for check ID %s is using a textual criteria '%s' but is flagged as a numeric type.  Did you mean 'metric_type = \"text\"'?", i, rs.CheckCID, rule.Criteria)
			}
		}
		if stringInSlice(rule.Criteria, []string{apiRuleSetMaxValue, apiRuleSetMinValue, apiRuleSetEqValue, apiRuleSetNotEqValue}) {
			if rs.MetricType != "numeric" {
				return fmt.Errorf("rule %d for check ID %s is using a numeric criteria '%s' but is flagged as a text type.  Did you mean 'metric_type = \"numeric\"'?", i, rs.CheckCID, rule.Criteria)
			}
		}
	}

	return nil
}
