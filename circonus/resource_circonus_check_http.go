package circonus

import (
	"bytes"
	"fmt"
	"log"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"github.com/circonus-labs/go-apiclient/config"
	"github.com/circonus-labs/terraform-provider-circonus/internal/hashcode"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	// circonus_check.http.* resource attribute names.
	checkHTTPAuthMethodAttr   = "auth_method"
	checkHTTPAuthPasswordAttr = "auth_password"
	checkHTTPAuthUserAttr     = "auth_user"
	checkHTTPBodyRegexpAttr   = "body_regexp"
	checkHTTPCAChainAttr      = "ca_chain"
	checkHTTPCertFileAttr     = "certificate_file"
	checkHTTPCiphersAttr      = "ciphers"
	checkHTTPCodeRegexpAttr   = "code"
	checkHTTPExtractAttr      = "extract"
	checkHTTPHeadersAttr      = "headers"
	checkHTTPKeyFileAttr      = "key_file"
	checkHTTPMethodAttr       = "method"
	checkHTTPPayloadAttr      = "payload"
	checkHTTPReadLimitAttr    = "read_limit"
	checkHTTPURLAttr          = "url"
	checkHTTPVersionAttr      = "version"
	checkHTTPRedirectsAttr    = "redirects"
)

var checkHTTPDescriptions = attrDescrs{
	checkHTTPAuthMethodAttr:   "The HTTP Authentication method",
	checkHTTPAuthPasswordAttr: "The HTTP Authentication user password",
	checkHTTPAuthUserAttr:     "The HTTP Authentication user name",
	checkHTTPBodyRegexpAttr:   `This regular expression is matched against the body of the response. If a match is not found, the check will be marked as "bad.`,
	checkHTTPCAChainAttr:      "A path to a file containing all the certificate authorities that should be loaded to validate the remote certificate (for TLS checks)",
	checkHTTPCodeRegexpAttr:   `The HTTP code that is expected. If the code received does not match this regular expression, the check is marked as "bad."`,
	checkHTTPCiphersAttr:      "A list of ciphers to be used in the TLS protocol (for HTTPS checks)",
	checkHTTPCertFileAttr:     "A path to a file containing the client certificate that will be presented to the remote server (for TLS-enabled checks)",
	checkHTTPExtractAttr:      "This regular expression is matched against the body of the response globally. The first capturing match is the key and the second capturing match is the value. Each key/value extracted is registered as a metric for the check.",
	checkHTTPHeadersAttr:      "Map of HTTP Headers to send along with HTTP Requests",
	checkHTTPKeyFileAttr:      "A path to a file containing key to be used in conjunction with the cilent certificate (for TLS checks)",
	checkHTTPMethodAttr:       "The HTTP method to use",
	checkHTTPPayloadAttr:      "The information transferred as the payload of an HTTP request",
	checkHTTPReadLimitAttr:    "Sets an approximate limit on the data read (0 means no limit)",
	checkHTTPURLAttr:          "The URL to use as the target of the check",
	checkHTTPVersionAttr:      "Sets the HTTP version for the check to use",
	checkHTTPRedirectsAttr:    "The maximum number of Location header redirects to follow.",
}

var schemaCheckHTTP = &schema.Schema{
	Type:     schema.TypeSet,
	Optional: true,
	MaxItems: 1,
	MinItems: 1,
	Set:      hashCheckHTTP,
	Elem: &schema.Resource{
		Schema: convertToHelperSchema(checkHTTPDescriptions, map[schemaAttr]*schema.Schema{
			checkHTTPAuthMethodAttr: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateRegexp(checkHTTPAuthMethodAttr, `^(?:Basic|Digest|Auto)$`),
			},
			checkHTTPAuthPasswordAttr: {
				Type:         schema.TypeString,
				Optional:     true,
				Sensitive:    true,
				ValidateFunc: validateRegexp(checkHTTPAuthPasswordAttr, `^.*`),
			},
			checkHTTPAuthUserAttr: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateRegexp(checkHTTPAuthUserAttr, `[^:]+`),
			},
			checkHTTPBodyRegexpAttr: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateRegexp(checkHTTPBodyRegexpAttr, `.+`),
			},
			checkHTTPCAChainAttr: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateRegexp(checkHTTPCAChainAttr, `.+`),
			},
			checkHTTPCertFileAttr: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateRegexp(checkHTTPCertFileAttr, `.+`),
			},
			checkHTTPCiphersAttr: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateRegexp(checkHTTPCiphersAttr, `.+`),
			},
			checkHTTPCodeRegexpAttr: {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      defaultCheckHTTPCodeRegexp,
				ValidateFunc: validateRegexp(checkHTTPCodeRegexpAttr, `.+`),
			},
			checkHTTPExtractAttr: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateRegexp(checkHTTPExtractAttr, `.+`),
			},
			checkHTTPHeadersAttr: {
				Type:         schema.TypeMap,
				Elem:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateHTTPHeaders,
			},
			checkHTTPKeyFileAttr: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateRegexp(checkHTTPKeyFileAttr, `.+`),
			},
			checkHTTPMethodAttr: {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      defaultCheckHTTPMethod,
				ValidateFunc: validateRegexp(checkHTTPMethodAttr, `\S+`),
			},
			checkHTTPPayloadAttr: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateRegexp(checkHTTPPayloadAttr, `\S+`),
			},
			checkHTTPReadLimitAttr: {
				Type:     schema.TypeInt,
				Optional: true,
				ValidateFunc: validateFuncs(
					validateIntMin(checkHTTPReadLimitAttr, 0),
				),
			},
			checkHTTPURLAttr: {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validateFuncs(
					validateHTTPURL(checkHTTPURLAttr, urlIsAbs),
				),
			},
			checkHTTPVersionAttr: {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      defaultCheckHTTPVersion,
				ValidateFunc: validateStringIn(checkHTTPVersionAttr, supportedHTTPVersions),
			},
			checkHTTPRedirectsAttr: {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      defaultCheckHTTPRedirects,
				ValidateFunc: validateRegexp(checkHTTPRedirectsAttr, `^[0-9]+$`),
			},
		}),
	},
}

// checkAPIToStateHTTP reads the Config data out of circonusCheck.CheckBundle into the
// statefile.
func checkAPIToStateHTTP(c *circonusCheck, d *schema.ResourceData) error {
	httpConfig := make(map[string]interface{}, len(c.Config))

	// swamp is a sanity check: it must be empty by the time this method returns
	swamp := make(map[config.Key]string, len(c.Config))
	for k, v := range c.Config {
		swamp[k] = v
	}

	saveStringConfigToState := func(apiKey config.Key, attrName schemaAttr) {
		if v, ok := c.Config[apiKey]; ok {
			httpConfig[string(attrName)] = v
		}

		delete(swamp, apiKey)
	}

	saveIntConfigToState := func(apiKey config.Key, attrName schemaAttr) {
		if v, ok := c.Config[apiKey]; ok {
			i, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				log.Printf("[ERROR]: Unable to convert %s to an integer: %v", apiKey, err)
				return
			}

			httpConfig[string(attrName)] = int(i)
		}

		delete(swamp, apiKey)
	}

	saveStringConfigToState(config.AuthMethod, checkHTTPAuthMethodAttr)
	saveStringConfigToState(config.AuthPassword, checkHTTPAuthPasswordAttr)
	saveStringConfigToState(config.AuthUser, checkHTTPAuthUserAttr)
	saveStringConfigToState(config.Body, checkHTTPBodyRegexpAttr)
	saveStringConfigToState(config.CAChain, checkHTTPCAChainAttr)
	saveStringConfigToState(config.CertFile, checkHTTPCertFileAttr)
	saveStringConfigToState(config.Ciphers, checkHTTPCiphersAttr)
	saveStringConfigToState(config.Code, checkHTTPCodeRegexpAttr)
	saveStringConfigToState(config.Extract, checkHTTPExtractAttr)

	headers := make(map[string]interface{}, len(c.Config))
	headerPrefixLen := len(config.HeaderPrefix)
	for k, v := range c.Config {
		if len(k) <= headerPrefixLen {
			continue
		}

		if strings.Compare(string(k[:headerPrefixLen]), string(config.HeaderPrefix)) == 0 {
			key := k[headerPrefixLen:]
			headers[string(key)] = v
		}
		delete(swamp, k)
	}
	httpConfig[string(checkHTTPHeadersAttr)] = headers

	saveStringConfigToState(config.KeyFile, checkHTTPKeyFileAttr)
	saveStringConfigToState(config.Method, checkHTTPMethodAttr)
	saveStringConfigToState(config.Payload, checkHTTPPayloadAttr)
	saveIntConfigToState(config.ReadLimit, checkHTTPReadLimitAttr)
	saveStringConfigToState(config.URL, checkHTTPURLAttr)
	saveStringConfigToState(config.HTTPVersion, checkHTTPVersionAttr)
	saveStringConfigToState(config.Redirects, checkHTTPRedirectsAttr)

	whitelistedConfigKeys := map[config.Key]struct{}{
		config.ReverseSecretKey: {},
		config.SubmissionURL:    {},
	}

	for k := range swamp {
		if _, ok := whitelistedConfigKeys[k]; ok {
			delete(c.Config, k)
		}

		if _, ok := whitelistedConfigKeys[k]; !ok {
			return fmt.Errorf("PROVIDER BUG: API Config not empty: %#v", swamp)
		}
	}

	if err := d.Set(checkHTTPAttr, schema.NewSet(hashCheckHTTP, []interface{}{httpConfig})); err != nil {
		return fmt.Errorf("Unable to store check %q attribute: %w", checkHTTPAttr, err)
	}

	return nil
}

// hashCheckHTTP creates a stable hash of the normalized values.
func hashCheckHTTP(v interface{}) int {
	m := v.(map[string]interface{})
	b := &bytes.Buffer{}
	b.Grow(defaultHashBufSize)

	writeInt := func(attrName schemaAttr) {
		if v, ok := m[string(attrName)]; ok {
			fmt.Fprintf(b, "%x", v.(int))
		}
	}

	writeString := func(attrName schemaAttr) {
		if v, ok := m[string(attrName)]; ok && v.(string) != "" {
			fmt.Fprint(b, strings.TrimSpace(v.(string)))
		}
	}

	// Order writes to the buffer using lexically sorted list for easy visual
	// reconciliation with other lists.
	writeString(checkHTTPAuthMethodAttr)
	writeString(checkHTTPAuthPasswordAttr)
	writeString(checkHTTPAuthUserAttr)
	writeString(checkHTTPBodyRegexpAttr)
	writeString(checkHTTPCAChainAttr)
	writeString(checkHTTPCertFileAttr)
	writeString(checkHTTPCiphersAttr)
	writeString(checkHTTPCodeRegexpAttr)
	writeString(checkHTTPExtractAttr)

	if headersRaw, ok := m[string(checkHTTPHeadersAttr)]; ok {
		headerMap := headersRaw.(map[string]interface{})
		headers := make([]string, 0, len(headerMap))
		for k := range headerMap {
			headers = append(headers, k)
		}

		sort.Strings(headers)
		for i := range headers {
			fmt.Fprint(b, headers[i])
			fmt.Fprint(b, headerMap[headers[i]].(string))
		}
	}

	writeString(checkHTTPKeyFileAttr)
	writeString(checkHTTPMethodAttr)
	writeString(checkHTTPPayloadAttr)
	writeInt(checkHTTPReadLimitAttr)
	writeString(checkHTTPURLAttr)
	writeString(checkHTTPVersionAttr)
	writeString(checkHTTPRedirectsAttr)

	s := b.String()
	return hashcode.String(s)
}

func checkConfigToAPIHTTP(c *circonusCheck, l interfaceList) error {
	c.Type = string(apiCheckTypeHTTP)

	// Iterate over all `http` attributes, even though we have a max of 1 in the
	// schema.

	//
	// Getting TWO, first one the correct config, the second one is blank...
	// which, with this arbitrary for loop results in the configuration being
	// overwritten with all blank values. (resulting in an API error for missing config attributes)
	//
	// just use the first one in the list if there are > 0 elements

	if len(l) == 0 {
		return fmt.Errorf("%d http configs found in list", len(l))
	}

	httpConfig := newInterfaceMap(l[0])
	// for _, mapRaw := range l {
	// 	httpConfig := newInterfaceMap(mapRaw)

	if v, found := httpConfig[checkHTTPAuthMethodAttr]; found {
		c.Config[config.AuthMethod] = v.(string)
	}

	if v, found := httpConfig[checkHTTPAuthPasswordAttr]; found {
		c.Config[config.AuthPassword] = v.(string)
	}

	if v, found := httpConfig[checkHTTPAuthUserAttr]; found {
		c.Config[config.AuthUser] = v.(string)
	}

	if v, found := httpConfig[checkHTTPBodyRegexpAttr]; found {
		c.Config[config.Body] = v.(string)
	}

	if v, found := httpConfig[checkHTTPCAChainAttr]; found {
		c.Config[config.CAChain] = v.(string)
	}

	if v, found := httpConfig[checkHTTPCertFileAttr]; found {
		c.Config[config.CertFile] = v.(string)
	}

	if v, found := httpConfig[checkHTTPCiphersAttr]; found {
		c.Config[config.Ciphers] = v.(string)
	}

	if v, found := httpConfig[checkHTTPCodeRegexpAttr]; found {
		c.Config[config.Code] = v.(string)
	}

	if v, found := httpConfig[checkHTTPExtractAttr]; found {
		c.Config[config.Extract] = v.(string)
	}

	for k, v := range httpConfig.CollectMap(checkHTTPHeadersAttr) {
		h := config.HeaderPrefix + config.Key(k)
		c.Config[h] = v
	}

	if v, found := httpConfig[checkHTTPKeyFileAttr]; found {
		c.Config[config.KeyFile] = v.(string)
	}

	if v, found := httpConfig[checkHTTPMethodAttr]; found {
		c.Config[config.Method] = v.(string)
	}

	if v, found := httpConfig[checkHTTPPayloadAttr]; found {
		c.Config[config.Payload] = v.(string)
	}

	if v, found := httpConfig[checkHTTPReadLimitAttr]; found {
		c.Config[config.ReadLimit] = fmt.Sprintf("%d", v.(int))
	}

	if v, found := httpConfig[checkHTTPURLAttr]; found {
		c.Config[config.URL] = v.(string)

		u, _ := url.Parse(v.(string))
		hostInfo := strings.SplitN(u.Host, ":", 2)
		if len(c.Target) == 0 {
			c.Target = hostInfo[0]
		}

		if len(hostInfo) > 1 && c.Config[config.Port] == "" {
			c.Config[config.Port] = hostInfo[1]
		}
	}

	if v, found := httpConfig[checkHTTPVersionAttr]; found {
		c.Config[config.HTTPVersion] = v.(string)
	}

	if v, found := httpConfig[checkHTTPRedirectsAttr]; found {
		c.Config[config.Redirects] = v.(string)
	}
	// }

	return nil
}
