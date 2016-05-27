package config

type EnvConfig struct {
	DB struct {
		Username string
		Password string
		Address  string
		Port     string
		DbName   string
	}
}

type GlobalConfig struct {
	Test struct {
		Good bool
	}
	Locations struct {
		SchoolLat float64
		SchoolLon float64
	}
}

const MYSQL_DATE_FORMAT = "2006-01-02 15:04:05"
