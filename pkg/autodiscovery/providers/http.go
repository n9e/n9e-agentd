package providers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"net/url"

	"github.com/n9e/n9e-agentd/pkg/api"
	"github.com/n9e/n9e-agentd/pkg/autodiscovery/integration"
	"github.com/n9e/n9e-agentd/pkg/autodiscovery/providers/names"
	"github.com/n9e/n9e-agentd/pkg/config"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/util/log"
	"k8s.io/klog/v2"
)

type Client struct {
	endpoint string
	path     string
	agentID  string
	token    string
	Header   http.Header
	client   http.Client
}

func newClient(endpoint, path, agentID, token string) (*Client, error) {
	cli := &Client{
		endpoint: endpoint,
		path:     path,
		agentID:  agentID,
		Header:   make(http.Header, 0),
		token:    token,
	}
	if token != "" {
		cli.Header.Set("Authorization", "Bearer "+token)
	}
	return cli, nil
}

func (c *Client) get(path string) ([]byte, error) {
	url := c.endpoint + path
	klog.V(6).Infof("get %s", url)

	req, err := http.NewRequest("GET", url, nil)
	logURL := log.SanitizeURL(url) // sanitized url that can be logged
	if err != nil {
		return nil, fmt.Errorf("Could not create request for transaction to invalid URL %q (dropping transaction): %s", logURL, err)
	}
	//req = req.WithContext(ctx)
	req.Header = c.Header
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Fail to read the response Body: %s", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("Error code %q received while sending transaction to %q: %s, dropping it", resp.Status, logURL, string(body))
	}

	klog.V(11).Infof("body %s", string(body))
	return body, nil
}

func (c *Client) getCollectRules() ([]api.CollectRule, error) {
	b, err := c.get(api.RoutePathGetCollectRules + "?ident=" + url.QueryEscape(c.agentID))
	if err != nil {
		return nil, err
	}

	var rules []api.CollectRule
	if err := json.Unmarshal(b, &rules); err != nil {
		return nil, err
	}

	return rules, nil
}

func (c *Client) getCollectRulesSummary() (*api.CollectRulesSummary, error) {
	b, err := c.get(api.RoutePathGetCollectRulesSummary + "?ident=" + url.QueryEscape(c.agentID))
	if err != nil {
		return nil, err
	}

	var summary api.CollectRulesSummary
	if err := json.Unmarshal(b, &summary); err != nil {
		return nil, err
	}

	return &summary, nil
}

// HttpConfigProvider implements the Config Provider interface
// It should be called periodically and returns templates from http for AutoConf.
type HttpConfigProvider struct {
	Client      *Client
	templateDir string
	cache       *ProviderCache
}

// NewHttpConfigProvider creates a client connection to http and create a new HttpConfigProvider
func NewHttpConfigProvider(cf config.ConfigurationProviders) (ConfigProvider, error) {
	cli, err := newClient(cf.TemplateURL, cf.TemplateDir, config.C.Ident, cf.Token)
	if err != nil {
		return nil, fmt.Errorf("Unable to instantiate the http client: %s", err)
	}
	cache := NewCPCache()
	return &HttpConfigProvider{
		Client:      cli,
		templateDir: cf.TemplateDir,
		cache:       cache,
	}, nil
}

// Collect retrieves templates from http, builds Config objects and returns them
// TODO: cache templates and last-modified index to avoid future full crawl if no template changed.
func (p *HttpConfigProvider) Collect() ([]integration.Config, error) {
	klog.V(6).Infof("entering Collect()")
	rules, err := p.Client.getCollectRules()
	if err != nil {
		klog.V(6).Infof("GetCollectRules err %s", err)
		return nil, err
	}
	klog.V(6).Infof("Collect() get %d rules", len(rules))

	var configs []integration.Config
	for _, rule := range rules {
		config, err := p.convertConfig(rule)
		if err != nil {
			klog.Warningf("%s %s is not a valid config file: %s", rule.Type, rule.Name, err)
			continue
		}
		configs = append(configs, *config)
	}

	return configs, nil
}

func (p *HttpConfigProvider) convertConfig(rule api.CollectRule) (*integration.Config, error) {
	config, err := ParseJSONConfig([]byte(rule.Data))
	if err != nil {
		return nil, err
	}
	config.Name = rule.Type
	config.Provider = names.Http
	config.Source = fmt.Sprintf("http:%s:%d", rule.Name, rule.ID)

	return config, nil
}

// IsUpToDate updates the list of AD templates versions in the Agent's cache and checks the list is up to date compared to http's data.
func (p *HttpConfigProvider) IsUpToDate() (bool, error) {
	klog.V(6).Infof("entering IsUpToDate")
	summary, err := p.Client.getCollectRulesSummary()
	if err != nil {
		klog.V(6).Infof("IsUpToDate %s", err)
		return false, err
	}

	adListUpdated := false
	dateIdx := p.cache.LatestTemplateIdx

	// When a node is deleted the Modified time of the children processed isn't changed.
	if p.cache.NumAdTemplates != summary.Total {
		if p.cache.NumAdTemplates != 0 {
			klog.V(5).Infof("List of AD Template was modified, updating cache.")
			adListUpdated = true
		}
		klog.V(5).Infof("Initializing cache for %v", p.String())
		p.cache.NumAdTemplates = summary.Total
	}

	dateIdx = math.Max(float64(summary.LatestUpdatedAt), dateIdx)
	if dateIdx > p.cache.LatestTemplateIdx || adListUpdated {
		klog.V(5).Infof("Idx was %v and is now %v", p.cache.LatestTemplateIdx, dateIdx)
		p.cache.LatestTemplateIdx = dateIdx
		klog.Infof("cache updated for %v", p.String())
		return false, nil
	}
	klog.Infof("cache up to date for %v", p.String())
	return true, nil
}

// String returns a string representation of the HttpConfigProvider
func (p *HttpConfigProvider) String() string {
	return names.Http
}

func init() {
	RegisterProvider("http", NewHttpConfigProvider)
}
