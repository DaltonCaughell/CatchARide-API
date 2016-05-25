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
