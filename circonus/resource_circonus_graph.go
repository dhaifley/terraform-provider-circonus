package circonus

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	api "github.com/circonus-labs/go-apiclient"
	"github.com/circonus-labs/go-apiclient/config"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	// circonus_graph.* resource attribute names.
	graphDescriptionAttr   = "description"
	graphLeftAttr          = "left"
	graphLineStyleAttr     = "line_style"
	graphMetricClusterAttr = "metric_cluster"
	graphNameAttr          = "name"
	graphNotesAttr         = "notes"
	graphRightAttr         = "right"
	graphMetricAttr        = "metric"
	graphStyleAttr         = "graph_style"
	graphTagsAttr          = "tags"
	graphGuidesAttr        = "guide"

	// circonus_graph.metric.* resource attribute names.
	graphMetricActiveAttr        = "active"
	graphMetricAlphaAttr         = "alpha"
	graphMetricAxisAttr          = "axis"
	graphMetricCAQLAttr          = "caql"
	graphMetricSearchAttr        = "search"
	graphMetricCheckAttr         = "check"
	graphMetricColorAttr         = "color"
	graphMetricFormulaAttr       = "formula"
	graphMetricFormulaLegendAttr = "legend_formula"
	graphMetricFunctionAttr      = "function" // derive
	graphMetricHumanNameAttr     = "name"
	graphMetricMetricTypeAttr    = "metric_type"
	graphMetricNameAttr          = "metric_name"
	graphMetricStackAttr         = "stack"

	// circonus_graph.metric_cluster.* resource attribute names.
	graphMetricClusterActiveAttr    = "active"
	graphMetricClusterAggregateAttr = "aggregate"
	graphMetricClusterAxisAttr      = "axis"
	graphMetricClusterColorAttr     = "color"
	graphMetricClusterQueryAttr     = "query"
	graphMetricClusterHumanNameAttr = "name"

	// circonus_graph.{left,right}.* resource attribute names.
	graphAxisLogarithmicAttr = "logarithmic"
	graphAxisMaxAttr         = "max"
	graphAxisMinAttr         = "min"

	// circonus_graph.guide.* resource attribute names.
	graphGuideHiddenAttr        = "hidden"
	graphGuideColorAttr         = "color"
	graphGuideFormulaAttr       = "formula"
	graphGuideFormulaLegendAttr = "legend_formula"
	graphGuideHumanNameAttr     = "name"
)

// const (
// 	apiGraphStyleLine = "line"
// )

var graphDescriptions = attrDescrs{
	// circonus_graph.* resource attribute names
	graphDescriptionAttr:   "",
	graphLeftAttr:          "",
	graphLineStyleAttr:     "How the line should change between point. A string containing either 'stepped', 'interpolated' or null.",
	graphNameAttr:          "",
	graphNotesAttr:         "",
	graphRightAttr:         "",
	graphMetricAttr:        "",
	graphMetricClusterAttr: "",
	graphStyleAttr:         "",
	graphTagsAttr:          "",
	graphGuidesAttr:        "",
}

var graphMetricDescriptions = attrDescrs{
	// circonus_graph.metric.* resource attribute names
	graphMetricActiveAttr:        "",
	graphMetricAlphaAttr:         "",
	graphMetricAxisAttr:          "",
	graphMetricCAQLAttr:          "",
	graphMetricSearchAttr:        "",
	graphMetricCheckAttr:         "",
	graphMetricColorAttr:         "",
	graphMetricFormulaAttr:       "",
	graphMetricFormulaLegendAttr: "",
	graphMetricFunctionAttr:      "",
	graphMetricMetricTypeAttr:    "",
	graphMetricHumanNameAttr:     "",
	graphMetricNameAttr:          "",
	graphMetricStackAttr:         "",
}

var graphGuidesDescriptions = attrDescrs{
	// circonus_graph.metric.* resource attribute names
	graphGuideHiddenAttr:        "",
	graphGuideColorAttr:         "",
	graphGuideFormulaAttr:       "",
	graphGuideFormulaLegendAttr: "",
	graphGuideHumanNameAttr:     "",
}

var graphMetricClusterDescriptions = attrDescrs{
	// circonus_graph.metric_cluster.* resource attribute names
	graphMetricClusterActiveAttr:    "",
	graphMetricClusterAggregateAttr: "",
	graphMetricClusterAxisAttr:      "",
	graphMetricClusterColorAttr:     "",
	graphMetricClusterQueryAttr:     "",
	graphMetricClusterHumanNameAttr: "",
}

// NOTE(sean@): There is no way to set a description on map inputs, but if that
// does happen:
//
// var graphMetricAxisOptionDescriptions = attrDescrs{
// 	// circonus_graph.if.value.over.* resource attribute names
// 	graphAxisLogarithmicAttr: "",
// 	graphAxisMaxAttr:         "",
// 	graphAxisMinAttr:         "",
// }

func resourceGraph() *schema.Resource {
	// makeConflictsWith := func(in ...schemaAttr) []string {
	// 	out := make([]string, 0, len(in))
	// 	for _, attr := range in {
	// 		out = append(out, string(graphMetricAttr)+"."+string(attr))
	// 	}
	// 	return out
	// }

	return &schema.Resource{
		Create: graphCreate,
		Read:   graphRead,
		Update: graphUpdate,
		Delete: graphDelete,
		Exists: graphExists,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: convertToHelperSchema(graphDescriptions, map[schemaAttr]*schema.Schema{
			graphDescriptionAttr: {
				Type:      schema.TypeString,
				Optional:  true,
				StateFunc: suppressWhitespace,
			},
			graphLeftAttr: {
				Type:         schema.TypeMap,
				Elem:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateGraphAxisOptions,
			},
			graphLineStyleAttr: {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      defaultGraphLineStyle,
				ValidateFunc: validateStringIn(graphLineStyleAttr, validGraphLineStyles),
			},
			graphNameAttr: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateRegexp(graphNameAttr, `.+`),
			},
			graphNotesAttr: {
				Type:     schema.TypeString,
				Optional: true,
			},
			graphRightAttr: {
				Type:         schema.TypeMap,
				Elem:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateGraphAxisOptions,
			},
			graphGuidesAttr: {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: convertToHelperSchema(graphGuidesDescriptions, map[schemaAttr]*schema.Schema{
						graphGuideHiddenAttr: {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						graphGuideColorAttr: {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validateRegexp(graphGuideColorAttr, `^#[0-9a-fA-F]{6}$`),
						},
						graphGuideFormulaAttr: {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validateRegexp(graphGuideFormulaAttr, `^.+$`),
						},
						graphGuideFormulaLegendAttr: {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validateRegexp(graphGuideFormulaLegendAttr, `^.+$`),
						},
						graphGuideHumanNameAttr: {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validateRegexp(graphGuideHumanNameAttr, `.+`),
						},
					}),
				},
			},
			graphMetricAttr: {
				Type:     schema.TypeList,
				Optional: true,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: convertToHelperSchema(graphMetricDescriptions, map[schemaAttr]*schema.Schema{
						graphMetricActiveAttr: {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
						graphMetricAlphaAttr: {
							Type:     schema.TypeString,
							Optional: true,
						},
						graphMetricAxisAttr: {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "left",
							ValidateFunc: validateStringIn(graphMetricAxisAttr, validAxisAttrs),
						},

						graphMetricCAQLAttr: {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validateRegexp(graphMetricCAQLAttr, `.+`),
							StateFunc: func(val interface{}) string {
								return strings.TrimSpace(val.(string))
							},
							// ConflictsWith: makeConflictsWith(graphMetricCheckAttr, graphMetricNameAttr, graphMetricSearchAttr),
						},
						graphMetricSearchAttr: {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validateRegexp(graphMetricSearchAttr, `.+`),
							// ConflictsWith: makeConflictsWith(graphMetricCheckAttr, graphMetricNameAttr, graphMetricCAQLAttr),
						},
						graphMetricCheckAttr: {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validateRegexp(graphMetricCheckAttr, config.CheckCIDRegex),
							// ConflictsWith: makeConflictsWith(graphMetricCAQLAttr, graphMetricSearchAttr),
						},
						graphMetricNameAttr: {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validateRegexp(graphMetricNameAttr, `.+`),
						},

						graphMetricColorAttr: {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validateRegexp(graphMetricColorAttr, `^#[0-9a-fA-F]{6}$`),
						},
						graphMetricFormulaAttr: {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validateRegexp(graphMetricFormulaAttr, `^.+$`),
						},
						graphMetricFormulaLegendAttr: {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validateRegexp(graphMetricFormulaLegendAttr, `^.+$`),
						},
						graphMetricFunctionAttr: {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validateStringIn(graphMetricFunctionAttr, validGraphFunctionValues),
						},
						graphMetricMetricTypeAttr: {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validateStringIn(graphMetricMetricTypeAttr, validMetricTypes),
						},
						graphMetricHumanNameAttr: {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validateRegexp(graphMetricHumanNameAttr, `.+`),
						},
						graphMetricStackAttr: {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validateRegexp(graphMetricStackAttr, `^[\d]*$`),
						},
					}),
				},
			},
			graphMetricClusterAttr: {
				Type:     schema.TypeList,
				Optional: true,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: convertToHelperSchema(graphMetricClusterDescriptions, map[schemaAttr]*schema.Schema{
						graphMetricClusterActiveAttr: {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
						graphMetricClusterAggregateAttr: {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "none",
							ValidateFunc: validateStringIn(graphMetricClusterAggregateAttr, validAggregateFuncs),
						},
						graphMetricClusterAxisAttr: {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "left",
							ValidateFunc: validateStringIn(graphMetricClusterAttr, validAxisAttrs),
						},
						graphMetricClusterColorAttr: {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validateRegexp(graphMetricClusterColorAttr, `^#[0-9a-fA-F]{6}$`),
						},
						graphMetricClusterQueryAttr: {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validateRegexp(graphMetricClusterQueryAttr, config.MetricClusterCIDRegex),
						},
						graphMetricClusterHumanNameAttr: {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validateRegexp(graphMetricHumanNameAttr, `.+`),
						},
					}),
				},
			},
			graphStyleAttr: {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      defaultGraphStyle,
				ValidateFunc: validateStringIn(graphStyleAttr, validGraphStyles),
			},
			graphTagsAttr: tagMakeConfigSchema(graphTagsAttr),
		}),
	}
}

func graphCreate(d *schema.ResourceData, meta interface{}) error {
	ctxt := meta.(*providerContext)
	g := newGraph()
	if err := g.ParseConfig(d); err != nil {
		return fmt.Errorf("error parsing graph schema during create: %w", err)
	}

	if err := g.Create(ctxt); err != nil {
		return fmt.Errorf("error creating graph: %w", err)
	}

	d.SetId(g.CID)

	return graphRead(d, meta)
}

func graphExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	ctxt := meta.(*providerContext)

	cid := d.Id()
	g, err := ctxt.client.FetchGraph(api.CIDType(&cid))
	if err != nil {
		if strings.Contains(err.Error(), defaultCirconus404ErrorString) {
			return false, nil
		}

		return false, err
	}

	if g.CID == "" {
		return false, nil
	}

	return true, nil
}

// graphRead pulls data out of the Graph object and stores it into the
// appropriate place in the statefile.
func graphRead(d *schema.ResourceData, meta interface{}) error {
	ctxt := meta.(*providerContext)

	cid := d.Id()
	g, err := loadGraph(ctxt, api.CIDType(&cid))
	if err != nil {
		return err
	}

	d.SetId(g.CID)

	metrics := make([]interface{}, 0, len(g.Datapoints))
	for _, datapoint := range g.Datapoints {
		dataPointAttrs := make(map[string]interface{}, 13) // 13 == len(members in api.GraphDatapoint)

		dataPointAttrs[string(graphMetricActiveAttr)] = !datapoint.Hidden

		if datapoint.Alpha != nil && *datapoint.Alpha != "0" {
			dataPointAttrs[string(graphMetricAlphaAttr)] = *datapoint.Alpha
		} else {
			dataPointAttrs[string(graphMetricAlphaAttr)] = nil
		}

		switch datapoint.Axis {
		case "l", "":
			dataPointAttrs[string(graphMetricAxisAttr)] = "left"
		case "r":
			dataPointAttrs[string(graphMetricAxisAttr)] = "right"
		default:
			return fmt.Errorf("PROVIDER BUG: Unsupported axis type %q", datapoint.Axis)
		}

		if datapoint.CAQL != nil && *datapoint.CAQL != "" {
			dataPointAttrs[string(graphMetricCAQLAttr)] = *datapoint.CAQL
		}

		if datapoint.Search != nil && *datapoint.Search != "" {
			dataPointAttrs[string(graphMetricSearchAttr)] = *datapoint.Search
		}

		if datapoint.CheckID != 0 {
			dataPointAttrs[string(graphMetricCheckAttr)] = fmt.Sprintf("%s/%d", config.CheckPrefix, datapoint.CheckID)
		}

		if datapoint.Color != nil {
			dataPointAttrs[string(graphMetricColorAttr)] = *datapoint.Color
		}

		if datapoint.DataFormula != nil {
			dataPointAttrs[string(graphMetricFormulaAttr)] = *datapoint.DataFormula
		}

		switch u := datapoint.Derive.(type) {
		case bool:
		case string:
			dataPointAttrs[string(graphMetricFunctionAttr)] = u
		default:
			return fmt.Errorf("PROVIDER BUG: Unsupported type for derive: %T", datapoint.Derive)
		}

		if datapoint.LegendFormula != nil {
			dataPointAttrs[string(graphMetricFormulaLegendAttr)] = *datapoint.LegendFormula
		}

		if datapoint.MetricName != "" {
			dataPointAttrs[string(graphMetricNameAttr)] = datapoint.MetricName
		}

		if datapoint.MetricType != "" {
			dataPointAttrs[string(graphMetricMetricTypeAttr)] = datapoint.MetricType
		}

		if datapoint.Name != "" {
			dataPointAttrs[string(graphMetricHumanNameAttr)] = datapoint.Name
		}

		if datapoint.Stack != nil {
			dataPointAttrs[string(graphMetricStackAttr)] = fmt.Sprintf("%d", *datapoint.Stack)
		}

		metrics = append(metrics, dataPointAttrs)
	}

	metricClusters := make([]interface{}, 0, len(g.MetricClusters))
	for _, metricCluster := range g.MetricClusters {
		metricClusterAttrs := make(map[string]interface{}, 8) // 8 == len(num struct attrs in api.GraphMetricCluster)

		metricClusterAttrs[string(graphMetricClusterActiveAttr)] = !metricCluster.Hidden

		if metricCluster.AggregateFunc != "" {
			metricClusterAttrs[string(graphMetricClusterAggregateAttr)] = metricCluster.AggregateFunc
		}

		switch metricCluster.Axis {
		case "l", "":
			metricClusterAttrs[string(graphMetricClusterAxisAttr)] = "left"
		case "r":
			metricClusterAttrs[string(graphMetricClusterAxisAttr)] = "right"
		default:
			return fmt.Errorf("PROVIDER BUG: Unsupported axis type %q", metricCluster.Axis)
		}

		if metricCluster.Color != nil {
			metricClusterAttrs[string(graphMetricClusterColorAttr)] = *metricCluster.Color
		}

		if metricCluster.DataFormula != nil {
			metricClusterAttrs[string(graphMetricFormulaAttr)] = *metricCluster.DataFormula
		}

		if metricCluster.LegendFormula != nil {
			metricClusterAttrs[string(graphMetricFormulaLegendAttr)] = *metricCluster.LegendFormula
		}

		if metricCluster.MetricCluster != "" {
			metricClusterAttrs[string(graphMetricClusterQueryAttr)] = metricCluster.MetricCluster
		}

		if metricCluster.Name != "" {
			metricClusterAttrs[string(graphMetricHumanNameAttr)] = metricCluster.Name
		}

		if metricCluster.Stack != nil {
			metricClusterAttrs[string(graphMetricStackAttr)] = fmt.Sprintf("%d", *metricCluster.Stack)
		}

		metricClusters = append(metricClusters, metricClusterAttrs)
	}

	leftAxisMap := make(map[string]interface{}, 3)
	if g.LogLeftY != nil {
		leftAxisMap[string(graphAxisLogarithmicAttr)] = fmt.Sprintf("%d", *g.LogLeftY)
	}

	if g.MaxLeftY != nil {
		leftAxisMap[string(graphAxisMaxAttr)] = strconv.FormatFloat(*g.MaxLeftY, 'f', -1, 64)
	}

	if g.MinLeftY != nil {
		leftAxisMap[string(graphAxisMinAttr)] = strconv.FormatFloat(*g.MinLeftY, 'f', -1, 64)
	}

	rightAxisMap := make(map[string]interface{}, 3)
	if g.LogRightY != nil {
		rightAxisMap[string(graphAxisLogarithmicAttr)] = fmt.Sprintf("%d", *g.LogRightY)
	}

	if g.MaxRightY != nil {
		rightAxisMap[string(graphAxisMaxAttr)] = strconv.FormatFloat(*g.MaxRightY, 'f', -1, 64)
	}

	if g.MinRightY != nil {
		rightAxisMap[string(graphAxisMinAttr)] = strconv.FormatFloat(*g.MinRightY, 'f', -1, 64)
	}

	_ = d.Set(graphDescriptionAttr, g.Description)

	if err := d.Set(graphLeftAttr, leftAxisMap); err != nil {
		return fmt.Errorf("Unable to store graph %q attribute: %w", graphLeftAttr, err)
	}

	_ = d.Set(graphLineStyleAttr, g.LineStyle)
	_ = d.Set(graphNameAttr, g.Title)
	_ = d.Set(graphNotesAttr, indirect(g.Notes))

	if err := d.Set(graphRightAttr, rightAxisMap); err != nil {
		return fmt.Errorf("Unable to store graph %q attribute: %w", graphRightAttr, err)
	}

	if err := d.Set(graphMetricAttr, metrics); err != nil {
		return fmt.Errorf("Unable to store graph %q attribute: %w", graphMetricAttr, err)
	}

	if err := d.Set(graphMetricClusterAttr, metricClusters); err != nil {
		return fmt.Errorf("Unable to store graph %q attribute: %w", graphMetricClusterAttr, err)
	}

	_ = d.Set(graphStyleAttr, g.Style)

	if err := d.Set(graphTagsAttr, tagsToState(apiToTags(g.Tags))); err != nil {
		return fmt.Errorf("Unable to store graph %q attribute: %w", graphTagsAttr, err)
	}

	guides := make([]interface{}, 0, len(g.Guides))
	for _, guide := range g.Guides {
		guideAttrs := make(map[string]interface{}, 5)

		guideAttrs[string(graphGuideHiddenAttr)] = guide.Hidden

		guideAttrs[string(graphGuideColorAttr)] = guide.Color

		if guide.DataFormula != nil {
			guideAttrs[string(graphGuideFormulaAttr)] = *guide.DataFormula
		}

		if guide.LegendFormula != nil {
			guideAttrs[string(graphGuideFormulaLegendAttr)] = *guide.LegendFormula
		}

		if guide.Name != "" {
			guideAttrs[string(graphGuideHumanNameAttr)] = guide.Name
		}

		guides = append(guides, guideAttrs)
	}
	_ = d.Set(graphGuidesAttr, guides)

	return nil
}

func graphUpdate(d *schema.ResourceData, meta interface{}) error {
	ctxt := meta.(*providerContext)
	g := newGraph()
	if err := g.ParseConfig(d); err != nil {
		return err
	}

	g.CID = d.Id()
	if err := g.Update(ctxt); err != nil {
		return fmt.Errorf("unable to update graph %q: %w", d.Id(), err)
	}

	return graphRead(d, meta)
}

func graphDelete(d *schema.ResourceData, meta interface{}) error {
	ctxt := meta.(*providerContext)

	cid := d.Id()
	if _, err := ctxt.client.DeleteGraphByCID(api.CIDType(&cid)); err != nil {
		return fmt.Errorf("unable to delete graph %q: %w", d.Id(), err)
	}

	d.SetId("")

	return nil
}

type circonusGraph struct {
	api.Graph
}

func newGraph() circonusGraph {
	g := circonusGraph{
		Graph: *api.NewGraph(),
	}

	return g
}

func loadGraph(ctxt *providerContext, cid api.CIDType) (circonusGraph, error) {
	var g circonusGraph
	ng, err := ctxt.client.FetchGraph(cid)
	if err != nil {
		return circonusGraph{}, err
	}
	g.Graph = *ng
	log.Printf("[loadGraph] %#v\n", *ng)

	return g, nil
}

// ParseConfig reads Terraform config data and stores the information into a
// Circonus Graph object.  ParseConfig and graphRead() must be kept in sync.
func (g *circonusGraph) ParseConfig(d *schema.ResourceData) error {
	g.Datapoints = make([]api.GraphDatapoint, 0, defaultGraphDatapoints)

	if v, found := d.GetOk(graphLeftAttr); found {
		listRaw := v.(map[string]interface{})
		leftAxisMap := make(map[string]interface{}, len(listRaw))
		for k, v := range listRaw {
			leftAxisMap[k] = v
		}

		if v, ok := leftAxisMap[string(graphAxisLogarithmicAttr)]; ok {
			i64, _ := strconv.ParseInt(v.(string), 10, 64)
			i := int(i64)
			g.LogLeftY = &i
		}

		if v, ok := leftAxisMap[string(graphAxisMaxAttr)]; ok && v.(string) != "" {
			f, _ := strconv.ParseFloat(v.(string), 64)
			g.MaxLeftY = &f
		}

		if v, ok := leftAxisMap[string(graphAxisMinAttr)]; ok && v.(string) != "" {
			f, _ := strconv.ParseFloat(v.(string), 64)
			g.MinLeftY = &f
		}
	}

	if v, found := d.GetOk(graphRightAttr); found {
		listRaw := v.(map[string]interface{})
		rightAxisMap := make(map[string]interface{}, len(listRaw))
		for k, v := range listRaw {
			rightAxisMap[k] = v
		}

		if v, ok := rightAxisMap[string(graphAxisLogarithmicAttr)]; ok {
			i64, _ := strconv.ParseInt(v.(string), 10, 64)
			i := int(i64)
			g.LogRightY = &i
		}

		if v, ok := rightAxisMap[string(graphAxisMaxAttr)]; ok && v.(string) != "" {
			f, _ := strconv.ParseFloat(v.(string), 64)
			g.MaxRightY = &f
		}

		if v, ok := rightAxisMap[string(graphAxisMinAttr)]; ok && v.(string) != "" {
			f, _ := strconv.ParseFloat(v.(string), 64)
			g.MinRightY = &f
		}
	}

	if v, found := d.GetOk(graphDescriptionAttr); found {
		g.Description = v.(string)
	}

	if v, found := d.GetOk(graphLineStyleAttr); found {
		switch v := v.(type) {
		case string:
			s := v
			g.LineStyle = &s
		case *string:
			g.LineStyle = v
		default:
			return fmt.Errorf("PROVIDER BUG: unsupported type for %q: %T", graphLineStyleAttr, v)
		}
	}

	if v, found := d.GetOk(graphNameAttr); found {
		g.Title = v.(string)
	}

	if v, found := d.GetOk(graphNotesAttr); found {
		s := v.(string)
		g.Notes = &s
	}

	if listRaw, found := d.GetOk(graphMetricAttr); found {
		metricList := listRaw.([]interface{})
		for metricIdx, metricListElem := range metricList {
			metricAttrs := newInterfaceMap(metricListElem.(map[string]interface{}))
			datapoint := api.GraphDatapoint{}

			if v, found := metricAttrs[graphMetricActiveAttr]; found {
				datapoint.Hidden = !(v.(bool))
			}

			if v, found := metricAttrs[graphMetricAlphaAttr]; found {
				f := v.(string)
				datapoint.Alpha = &f
			} else {
				datapoint.Alpha = nil
			}

			defaultAlpha := "0"
			if datapoint.Alpha == nil || *datapoint.Alpha == "" {
				datapoint.Alpha = &defaultAlpha
			}

			if v, found := metricAttrs[graphMetricAxisAttr]; found {
				switch v.(string) {
				case "left", "":
					datapoint.Axis = "l"
				case "right":
					datapoint.Axis = "r"
				default:
					return fmt.Errorf("PROVIDER BUG: Unsupported axis attribute %q: %q", graphMetricAxisAttr, v.(string))
				}
			}

			if v, found := metricAttrs[graphMetricColorAttr]; found {
				s := v.(string)
				datapoint.Color = &s
			}

			if v, found := metricAttrs[graphMetricFormulaAttr]; found {
				s := v.(string)
				if s != "" {
					datapoint.DataFormula = &s
				} else {
					datapoint.DataFormula = nil
				}
			} else {
				datapoint.DataFormula = nil
			}

			if v, found := metricAttrs[graphMetricFunctionAttr]; found {
				s := v.(string)
				if s != "" {
					datapoint.Derive = s
				} else {
					datapoint.Derive = false
				}
			} else {
				datapoint.Derive = false
			}

			if v, found := metricAttrs[graphMetricFormulaLegendAttr]; found {
				s := v.(string)
				if s != "" {
					datapoint.LegendFormula = &s
				} else {
					datapoint.LegendFormula = nil
				}
			} else {
				datapoint.LegendFormula = nil
			}

			if v, found := metricAttrs[graphMetricMetricTypeAttr]; found {
				s := v.(string)
				if s != "" {
					datapoint.MetricType = s
				}
			}

			if v, found := metricAttrs[graphMetricHumanNameAttr]; found {
				s := v.(string)
				if s != "" {
					s = strings.TrimSpace(s)
					datapoint.Name = s
				}
			}

			if v, found := metricAttrs[graphMetricStackAttr]; found {
				s := v.(string)
				if s != "" {
					u64, _ := strconv.ParseUint(s, 10, 64)
					u := uint(u64)
					datapoint.Stack = &u
				}
			}

			//
			// metric locator can be ONE of the following:
			//   check id + metric name
			//   caql query
			//   search expression
			//
			// ConflictWith no longer works on non-list schema elements,
			// so we have to enforce it here.
			caql := ""
			search := ""
			check := uint(0)
			name := ""

			if v, found := metricAttrs[graphMetricNameAttr]; found {
				s := strings.TrimSpace(v.(string))
				if s != "" {
					name = s
				}
			}
			if v, found := metricAttrs[graphMetricCheckAttr]; found {
				re := regexp.MustCompile(config.CheckCIDRegex)
				matches := re.FindStringSubmatch(v.(string))
				if len(matches) == 3 {
					checkID, _ := strconv.ParseUint(matches[2], 10, 64)
					check = uint(checkID)
				}
			}

			if v, found := metricAttrs[graphMetricCAQLAttr]; found {
				s := strings.TrimSpace(v.(string))
				if s != "" {
					caql = s
				}
			}

			if v, found := metricAttrs[graphMetricSearchAttr]; found {
				s := strings.TrimSpace(v.(string))
				if s != "" {
					search = s
				}
			}

			metricLocatorError := fmt.Errorf("metric[%d] name=%q: locator issue - %q(%v) + %q(%v) OR %q(%v) OR %q(%v)",
				metricIdx, datapoint.Name,
				graphMetricCheckAttr, check,
				graphMetricNameAttr, name,
				graphMetricCAQLAttr, caql,
				graphMetricSearchAttr, search)
			datapoint.CAQL = nil
			datapoint.Search = nil

			switch {
			case check == 0 && name != "":
				return fmt.Errorf("metric[%d] name=%q: locator using %q requires %q", metricIdx, datapoint.Name, graphMetricNameAttr, graphMetricCheckAttr)
			case check > 0 && name == "":
				return fmt.Errorf("metric[%d] name=%q: locator using %q requires %q", metricIdx, datapoint.Name, graphMetricCheckAttr, graphMetricNameAttr)
			case check > 0 && (caql != "" || search != ""):
				return metricLocatorError
			case caql != "" && (check != 0 || name != "" || search != ""):
				return metricLocatorError
			case search != "" && (check != 0 || name != "" || caql != ""):
				return metricLocatorError
			default:
				switch {
				case check > 0:
					datapoint.CheckID = check
					datapoint.MetricName = name
				case caql != "":
					datapoint.CAQL = &caql
				case search != "":
					datapoint.Search = &search
				}
			}

			g.Datapoints = append(g.Datapoints, datapoint)
		}
	}

	if listRaw, found := d.GetOk(graphMetricClusterAttr); found {
		metricClusterList := listRaw.([]interface{})

		for _, metricClusterListRaw := range metricClusterList {
			metricClusterAttrs := newInterfaceMap(metricClusterListRaw.(map[string]interface{}))

			metricCluster := api.GraphMetricCluster{}

			if v, found := metricClusterAttrs[graphMetricClusterActiveAttr]; found {
				metricCluster.Hidden = !(v.(bool))
			}

			if v, found := metricClusterAttrs[graphMetricClusterAggregateAttr]; found {
				metricCluster.AggregateFunc = v.(string)
			}

			if v, found := metricClusterAttrs[graphMetricClusterAxisAttr]; found {
				switch v.(string) {
				case "left", "":
					metricCluster.Axis = "l"
				case "right":
					metricCluster.Axis = "r"
				default:
					return fmt.Errorf("PROVIDER BUG: Unsupported axis attribute %q: %q", graphMetricClusterAxisAttr, v.(string))
				}
			}

			if v, found := metricClusterAttrs[graphMetricClusterColorAttr]; found {
				s := v.(string)
				if s != "" {
					metricCluster.Color = &s
				}
			}

			if v, found := metricClusterAttrs[graphMetricFormulaAttr]; found {
				switch v := v.(type) {
				case string:
					s := v
					metricCluster.DataFormula = &s
				case *string:
					metricCluster.DataFormula = v
				default:
					return fmt.Errorf("PROVIDER BUG: unsupported type for %q: %T", graphMetricFormulaAttr, v)
				}
			}

			if v, found := metricClusterAttrs[graphMetricFormulaLegendAttr]; found {
				switch v := v.(type) {
				case string:
					s := v
					metricCluster.LegendFormula = &s
				case *string:
					metricCluster.LegendFormula = v
				default:
					return fmt.Errorf("PROVIDER BUG: unsupported type for %q: %T", graphMetricFormulaLegendAttr, v)
				}
			}

			if v, found := metricClusterAttrs[graphMetricClusterQueryAttr]; found {
				s := v.(string)
				if s != "" {
					metricCluster.MetricCluster = s
				}
			}

			if v, found := metricClusterAttrs[graphMetricHumanNameAttr]; found {
				s := v.(string)
				if s != "" {
					metricCluster.Name = s
				}
			}

			if v, found := metricClusterAttrs[graphMetricStackAttr]; found {
				var stackStr string
				switch u := v.(type) {
				case string:
					stackStr = u
				case *string:
					if u != nil {
						stackStr = *u
					}
				default:
					return fmt.Errorf("PROVIDER BUG: unsupported type for %q: %T", graphMetricStackAttr, v)
				}

				if stackStr != "" {
					u64, _ := strconv.ParseUint(stackStr, 10, 64)
					u := uint(u64)
					metricCluster.Stack = &u
				}
			}

			g.MetricClusters = append(g.MetricClusters, metricCluster)
		}
	}

	if v, found := d.GetOk(graphStyleAttr); found {
		switch v := v.(type) {
		case string:
			s := v
			g.Style = &s
		case *string:
			g.Style = v
		default:
			return fmt.Errorf("PROVIDER BUG: unsupported type for %q: %T", graphStyleAttr, v)
		}
	}

	if v, found := d.GetOk(graphTagsAttr); found {
		g.Tags = derefStringList(flattenSet(v.(*schema.Set)))
	}

	if listRaw, found := d.GetOk(graphGuidesAttr); found {
		guideList := listRaw.([]interface{})
		for _, guideListElem := range guideList {
			guideAttrs := newInterfaceMap(guideListElem.(map[string]interface{}))
			guide := api.GraphGuide{}

			if v, found := guideAttrs[graphGuideHiddenAttr]; found {
				guide.Hidden = (v.(bool))
			}

			if v, found := guideAttrs[graphGuideColorAttr]; found {
				guide.Color = v.(string)
			}

			if v, found := guideAttrs[graphGuideFormulaAttr]; found {
				s := v.(string)
				if s != "" {
					guide.DataFormula = &s
				} else {
					guide.DataFormula = nil
				}
			} else {
				guide.DataFormula = nil
			}

			if v, found := guideAttrs[graphGuideFormulaLegendAttr]; found {
				s := v.(string)
				if s != "" {
					guide.LegendFormula = &s
				} else {
					guide.LegendFormula = nil
				}
			} else {
				guide.LegendFormula = nil
			}

			if v, found := guideAttrs[graphGuideHumanNameAttr]; found {
				s := v.(string)
				if s != "" {
					guide.Name = s
				}
			}

			g.Guides = append(g.Guides, guide)
		}
	}

	log.Printf("[ParseConfig] %#v\n", g.Graph)

	if err := g.Validate(); err != nil {
		return err
	}

	return nil
}

func (g *circonusGraph) Create(ctxt *providerContext) error {
	ng, err := ctxt.client.CreateGraph(&g.Graph)
	if err != nil {
		return err
	}

	g.CID = ng.CID

	return nil
}

func (g *circonusGraph) Update(ctxt *providerContext) error {
	_, err := ctxt.client.UpdateGraph(&g.Graph)
	if err != nil {
		return fmt.Errorf("Unable to update graph %s: %w", g.CID, err)
	}

	return nil
}

func (g *circonusGraph) Validate() error {
	for i, datapoint := range g.Datapoints {
		// if *g.Style == apiGraphStyleLine && datapoint.Alpha != nil && *datapoint.Alpha != "0" {
		// 	return fmt.Errorf("%s can not be set on graphs with style %s", graphMetricAlphaAttr, apiGraphStyleLine)
		// }

		if datapoint.CheckID != 0 && datapoint.MetricName == "" {
			return fmt.Errorf("Error with %s[%d] name=%q: %s is set, missing attribute %s must also be set", graphMetricAttr, i, datapoint.Name, graphMetricCheckAttr, graphMetricNameAttr)
		}

		if datapoint.CheckID == 0 && datapoint.MetricName != "" {
			return fmt.Errorf("Error with %s[%d] name=%q: %s is set, missing attribute %s must also be set", graphMetricAttr, i, datapoint.Name, graphMetricNameAttr, graphMetricCheckAttr)
		}

		// if datapoint.CAQL != nil && (datapoint.CheckID != 0 || datapoint.MetricName != "") {
		// 	return fmt.Errorf("Error with %s[%d] name=%q: %q attribute is mutually exclusive with attributes %s or %s or %s", graphMetricAttr, i, datapoint.Name, graphMetricCAQLAttr, graphMetricNameAttr, graphMetricCheckAttr, graphMetricSearchAttr)
		// }

		// if datapoint.Search != nil && (datapoint.CheckID != 0 || datapoint.MetricName != "") {
		// 	return fmt.Errorf("Error with %s[%d] name=%q: %q attribute is mutually exclusive with attributes %s or %s or %s", graphMetricAttr, i, datapoint.Name, graphMetricSearchAttr, graphMetricNameAttr, graphMetricCheckAttr, graphMetricCAQLAttr)
		// }

		if datapoint.MetricType == "text" && datapoint.Derive != nil {
			v := datapoint.Derive
			switch v.(type) {
			case bool:
			default:
				return fmt.Errorf("Error with %s[%d] (name=%q): attribute %q is mutually exclusive when %s=%q", graphMetricAttr, i, datapoint.Name, graphMetricFunctionAttr, graphMetricMetricTypeAttr, "text")
			}
		}
	}

	for i, mc := range g.MetricClusters {
		if mc.AggregateFunc != "" && (mc.Color == nil || *mc.Color == "") {
			return fmt.Errorf("Error with %s[%d] name=%q: %s is a required attribute for graphs with %s set", graphMetricClusterAttr, i, mc.Name, graphMetricClusterColorAttr, graphMetricClusterAggregateAttr)
		}
	}

	return nil
}
