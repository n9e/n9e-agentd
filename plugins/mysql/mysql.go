package mysql

import (
	"fmt"
	"time"

	"database/sql"

	"github.com/go-sql-driver/mysql"
	"github.com/n9e/n9e-agentd/pkg/autodiscovery/integration"
	"github.com/n9e/n9e-agentd/pkg/util"
	"github.com/n9e/n9e-agentd/pkg/util/tls"
	"github.com/n9e/n9e-agentd/plugins/mysql/db"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/aggregator"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/collector/check"
	core "github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/collector/corechecks"
	"sigs.k8s.io/yaml"
)

const checkName = "mysql"

type InitConfig struct {
	GlobalCustomQueries []db.CustomQuery `json:"globalCustomQueries"`
	Service             string           `json:"service"`
}

type InstanceConfig struct {
	Dsn string `json:"dsn" description:"[username[:password]@][protocol[(address)]]/[?tls=[true|false|skip-verify|custom]] see https://github.com/go-sql-driver/mysql#dsn-data-source-name"`
	//Host                   string            `json:"host"`
	//User                   string            `json:"user"`
	//Pass                   []string          `json:"pass"`
	//Port                   int               `json:"port"`
	//Sock                   string            `json:"user"`
	//ConnectTimeout         int               `json:"connectTimeout"`
	//Charset                string            `json:"user"`
	//DefaultsFile           string            `json:"user"`
	TLS                    tls.ClientConfig  `json:"tls"`
	UseGlobalCustomQueries string            `json:"useGlobalCustomQueries"`
	CustomQueries          []db.CustomQuery  `json:"customQuery"`
	MaxCustomQueries       int               `json:"maxCustomQueries"`
	Options                Options           `json:"options"`
	DeepDatabaseMonitoring bool              `json:"deepDatabaseMonitoring"`
	StatementMetricsLimits map[string][2]int `json:"statementMetricsLimits"`
	StatementSamples       StatementSamples  `json:"statementSamples"`

	InitConfig `json:"-"`
}

type StatementSamples struct {
	Enabled                         bool   `json:"Enabled"`
	CollectionsPerSecond            int    `json:"CollectionsPerSecond"`
	SamplesPerHourPerQuery          int    `json:"SamplesPerHourPerQuery"`
	ExplainedStatementsCacheMaxsize int    `json:"ExplainedStatementsCacheMaxsize"`
	SeenSamplesCacheMaxsize         int    `json:"SeenSamplesCacheMaxsize"`
	EventsStatementsRowLimit        int    `json:"EventsStatementsRowLimit"`
	EventsStatementsTable           string `json:"EventsStatementsTable"`
	ExplainProcedure                string `json:"ExplainProcedure"`
	FullyQualifiedExplainProcedure  string `json:"FullyQualifiedExplainProcedure"`
	EventsStatementsEnableProcedure string `json:"EventsStatementsEnableProcedure"`
	EventsStatementsTempTableName   string `json:"EventsStatementsTempTableName"`
	CollectionStrategyCacheMaxsize  int    `json:"CollectionStrategyCacheMaxsize"`
	CollectionStrategyCacheTtl      string `json:"CollectionStrategyCacheTtl"`
}

type Options struct {
	Replication                  bool   `json:"replication"`
	ReplicationChannel           string `json:"replicationChannel"`
	ReplicationNonBlockingStatus bool   `json:"replicationNonBlockingStatus"`
	GaleraCluster                bool   `json:"galeraCluster"`
	ExtraStatusMetrics           bool   `json:"extraStatusMetrics"`
	ExtraInnodbMetrics           bool   `json:"extraInnodbMetrics"`
	DisableInnodbMetrics         bool   `json:"disableInnodbMetrics"`
	SchemaSizeMetrics            bool   `json:"schemaSizeMetrics"`
	ExtraPerformanceMetrics      bool   `json:"extraPerformanceMetrics"`
}

type Config struct {
	InstanceConfig
	InitConfig
}

func (p Config) String() string {
	return util.Prettify(p)
}

func (c *Config) Validate() error {
	dsn, err := dsnAddTimeout(c.Dsn)
	if err != nil {
		return err
	}
	c.Dsn = dsn

	return nil
}

func defaultInstanceConfig() InstanceConfig {
	return InstanceConfig{MaxCustomQueries: 20}
}

func dsnAddTimeout(dsn string) (string, error) {
	conf, err := mysql.ParseDSN(dsn)
	if err != nil {
		return "", err
	}

	if conf.Timeout == 0 {
		conf.Timeout = time.Second * 10
	}

	return conf.FormatDSN(), nil
}

func buildConfig(rawInstance integration.Data, rawInitConfig integration.Data) (*Config, error) {
	instance := defaultInstanceConfig()
	initConfig := InitConfig{}

	err := yaml.Unmarshal(rawInitConfig, &initConfig)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(rawInstance, &instance)
	if err != nil {
		return nil, err
	}

	c := &Config{
		InitConfig:     initConfig,
		InstanceConfig: instance,
	}

	if err := c.Validate(); err != nil {
		return nil, err
	}

	return c, nil
}

// Check doesn't need additional fields
type Check struct {
	core.CheckBase
	sender aggregator.Sender

	config *Config
	db     *sql.DB

	qcacheStats      interface{}
	version          MySQLVersion
	queryManager     *db.QueryManager
	innodbStats      *InnoDBMetrics
	statementMetrics *MySQLStatementMetrics
	statementSamples *MySQLStatementSamples
}

// Run executes the check
func (c *Check) Run() (err error) {
	if c.sender, err = aggregator.GetSender(c.ID()); err != nil {
		return err
	}

	if c.db, err = sql.Open("mysql", c.config.Dsn); err != nil {
		return err
	}
	defer c.db.Close()

	if err := c.check(); err != nil {
		return err
	}

	c.sender.Commit()
	return nil
}

func (c *Check) Cancel() {
	defer c.CheckBase.Cancel()
	c.cancel()
}

// Configure the Prom check
func (c *Check) Configure(rawInstance integration.Data, rawInitConfig integration.Data, source string) error {
	// Must be called before c.CommonConfigure
	c.BuildID(rawInstance, rawInitConfig)

	err := c.CommonConfigure(rawInstance, source)
	if err != nil {
		return fmt.Errorf("common configure failed: %s", err)
	}

	config, err := buildConfig(rawInstance, rawInitConfig)
	if err != nil {
		return fmt.Errorf("build config failed: %s", err)
	}

	c.config = config

	if err := c.init(); err != nil {
		return err
	}

	return nil
}

func promFactory() check.Check {
	return &Check{
		CheckBase: core.NewCheckBase(checkName),
	}
}

func init() {
	core.RegisterCheck(checkName, promFactory)
}
