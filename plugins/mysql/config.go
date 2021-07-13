package mysql

import (
	"net"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/n9e/n9e-agentd/pkg/autodiscovery/integration"
	"github.com/n9e/n9e-agentd/pkg/util"
	"github.com/n9e/n9e-agentd/pkg/util/db"
	"github.com/n9e/n9e-agentd/pkg/util/tls"
	"sigs.k8s.io/yaml"
)

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
	UseGlobalCustomQueries string            `json:"useGlobalCustomQueries" description:"extent,true,false"`
	CustomQueries          []db.CustomQuery  `json:"customQueries"`
	MaxCustomQueries       int               `json:"maxCustomQueries"`
	Options                Options           `json:"options"`
	DeepDatabaseMonitoring bool              `json:"deepDatabaseMonitoring"`
	StatementMetricsLimits map[string][2]int `json:"statementMetricsLimits"`
	StatementSamples       StatementSamples  `json:"statementSamples"`

	InitConfig `json:"-"`
	server     string
	port       string
}

type StatementSamples struct {
	Enabled                            bool   `json:"enabled"`
	RunSync                            bool   `json:"runSync"`
	CollectionsPerSecond               int    `json:"collectionsPerSecond"`
	SamplesPerHourPerQuery             int    `json:"samplesPerHourPerQuery"`
	ExplainedStatementsCacheMaxsize    int    `json:"explainedStatementsCacheMaxsize"`
	ExplainedStatementsPerHourPerQuery int    `json:"explainedStatementsPerHourPerQuery"`
	SeenSamplesCacheMaxsize            int    `json:"seenSamplesCacheMaxsize"`
	EventsStatementsRowLimit           int    `json:"eventsStatementsRowLimit"`
	EventsStatementsTable              string `json:"eventsStatementsTable"`
	ExplainProcedure                   string `json:"explainProcedure"`
	FullyQualifiedExplainProcedure     string `json:"fullyQualifiedExplainProcedure"`
	EventsStatementsEnableProcedure    string `json:"eventsStatementsEnableProcedure"`
	EventsStatementsTempTableName      string `json:"eventsStatementsTempTableName"`
	CollectionStrategyCacheMaxsize     int    `json:"collectionStrategyCacheMaxsize"`
	CollectionStrategyCacheTtl         int    `json:"collectionStrategyCacheTtl"`
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

	{
		cfg, _ := mysql.ParseDSN(c.Dsn)
		if cfg.Net == "tcp" {
			c.server, c.port, _ = net.SplitHostPort(cfg.Addr)
		} else if cfg.Net == "unix" {
			c.server = cfg.Addr
			c.port = "unix_socket"
		}
	}

	if len(c.StatementMetricsLimits) == 0 {
		c.StatementMetricsLimits = DEFAULT_STATEMENT_METRICS_LIMITS
	}

	return nil
}

func defaultInstanceConfig() InstanceConfig {
	return InstanceConfig{
		MaxCustomQueries: 20,
		StatementSamples: StatementSamples{
			Enabled:                            false,
			CollectionsPerSecond:               -1,
			EventsStatementsRowLimit:           5000,
			ExplainProcedure:                   "explain_statement",
			FullyQualifiedExplainProcedure:     "n9e.explain_statement",
			EventsStatementsTempTableName:      "n9e.temp_events",
			EventsStatementsEnableProcedure:    "n9e.enable_events_statements_consumers",
			EventsStatementsTable:              "",
			CollectionStrategyCacheMaxsize:     1000,
			CollectionStrategyCacheTtl:         300,
			ExplainedStatementsCacheMaxsize:    5000,
			ExplainedStatementsPerHourPerQuery: 60,
			SeenSamplesCacheMaxsize:            10000,
			SamplesPerHourPerQuery:             15,
		},
	}
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
