package sql

import "fmt"

// Configuration for sql db
type Configuration struct {
	SQLType    string            `json:"sqlType"`
	User       string            `json:"user"`
	Password   string            `json:"password"`
	Protocol   string            `json:"protocol"`
	Host       string            `json:"host"`
	Port       int               `json:"port"`
	Database   string            `json:"database"`
	Parameters map[string]string `json:"parameters"`
}

// NewConfiguration creates a new configuration with some default values
func NewConfiguration() *Configuration {
	conf := &Configuration{
		Protocol:   "tcp",
		Port:       3306,
		Parameters: map[string]string{},
		SQLType:    "mysql",
	}

	return conf
}

func (config *Configuration) parameterString() string {
	s := ""

	if len(config.Parameters) > 0 {
		s = "?"
	}

	firstParam := true

	for k, p := range config.Parameters {
		if !firstParam {
			s += "&"
		}

		firstParam = false

		s += k + "=" + p
	}

	return s
}

// ConnectionString creates a connection string for sql.Open()
func (config *Configuration) ConnectionString() string {
	return fmt.Sprintf("%s:%s@%s(%s:%d)/%s%s",
		config.User,
		config.Password,
		config.Protocol,
		config.Host,
		config.Port,
		config.Database,
		config.parameterString(),
	)
}
