package config // import "admin-console/internal/config"
import (
	"net/url"
	"time"
)

type configValueType int

const (
	stringType configValueType = iota
	stringListType
	boolType
	intType
	int64Type
	urlType
	secondType
	minuteType
	hourType
	dayType
	secretFileType
	bytesType
)

type configValue struct {
	parsedStringValue string
	parsedIntValue    int
	parsedInt64Value  int64
	parsedDuration    time.Duration
	parsedStringList  []string
	parsedURLValue    *url.URL
	parsedBytesValue  []byte
	parsedBoolValue   bool

	secret    bool
	rawValue  string
	valueType configValueType
	targetKey string

	validator func(string) error
}

type configOptions struct {
	rootURL string
	baseURL string
	options map[string]*configValue
}

// NewConfigOptions creates a new instance of ConfigOptions with default values.
func NewConfigOptions() *configOptions {
	return &configOptions{
		rootURL: "http://localhost",
		baseURL: "",
		options: map[string]*configValue{
			"NODE_ID": {
				parsedIntValue: 1,
				rawValue:       "1",
				valueType:      intType,
				validator: func(rawValue string) error {
					return validateGreaterOrEqualThan(rawValue, 1)
				},
			},
			"DATABASE_CONNECTION_LIFETIME": {
				parsedDuration: time.Minute * 5,
				rawValue:       "5",
				valueType:      minuteType,
				validator: func(rawValue string) error {
					return validateGreaterThan(rawValue, 0)
				},
			},
			"DATABASE_MAX_CONNS": {
				parsedIntValue: 20,
				rawValue:       "20",
				valueType:      intType,
				validator: func(rawValue string) error {
					return validateGreaterOrEqualThan(rawValue, 1)
				},
			},
			"DATABASE_MIN_CONNS": {
				parsedIntValue: 1,
				rawValue:       "1",
				valueType:      intType,
				validator: func(rawValue string) error {
					return validateGreaterOrEqualThan(rawValue, 0)
				},
			},
			"DATABASE_URL_AC": {
				parsedStringValue: "user=postgres password=postgres dbname=admin-console sslmode=disable",
				rawValue:          "user=postgres password=postgres dbname=admin-console sslmode=disable",
				valueType:         stringType,
				secret:            true,
			},
			"DATABASE_URL_FILE": {
				parsedStringValue: "",
				rawValue:          "",
				valueType:         secretFileType,
				targetKey:         "DATABASE_URL_AC",
			},
			"LISTEN_ADDR": {
				parsedStringList: []string{"127.0.0.1:8080"},
				rawValue:         "127.0.0.1:8080",
				valueType:        stringListType,
			},
			"LOG_DATE_TIME": {
				parsedBoolValue: false,
				rawValue:        "0",
				valueType:       boolType,
			},
			"LOG_FILE": {
				parsedStringValue: "stderr",
				rawValue:          "stderr",
				valueType:         stringType,
			},
			"LOG_FORMAT": {
				parsedStringValue: "text",
				rawValue:          "text",
				valueType:         stringType,
				validator: func(rawValue string) error {
					return validateChoices(rawValue, []string{"text", "json"})
				},
			},
			"LOG_LEVEL": {
				parsedStringValue: "info",
				rawValue:          "info",
				valueType:         stringType,
				validator: func(rawValue string) error {
					return validateChoices(rawValue, []string{"debug", "info", "warning", "error"})
				},
			},
		},
	}
}

func (c *configOptions) DatabaseConnectionLifetime() time.Duration {
	return c.options["DATABASE_CONNECTION_LIFETIME"].parsedDuration
}

func (c *configOptions) DatabaseMaxConns() int {
	return c.options["DATABASE_MAX_CONNS"].parsedIntValue
}

func (c *configOptions) DatabaseMinConns() int {
	return c.options["DATABASE_MIN_CONNS"].parsedIntValue
}

func (c *configOptions) DatabaseURL() string {
	return c.options["DATABASE_URL_AC"].parsedStringValue
}

func (c *configOptions) ListenAddr() []string {
	return c.options["LISTEN_ADDR"].parsedStringList
}

func (c *configOptions) LogFile() string {
	return c.options["LOG_FILE"].parsedStringValue
}

func (c *configOptions) LogDateTime() bool {
	return c.options["LOG_DATE_TIME"].parsedBoolValue
}

func (c *configOptions) LogFormat() string {
	return c.options["LOG_FORMAT"].parsedStringValue
}

func (c *configOptions) LogLevel() string {
	return c.options["LOG_LEVEL"].parsedStringValue
}

func (c *configOptions) SetLogLevel(level string) {
	c.options["LOG_LEVEL"].parsedStringValue = level
	c.options["LOG_LEVEL"].rawValue = level
}
