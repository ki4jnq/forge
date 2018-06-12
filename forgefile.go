package forge

func NewParser(env string) *Unformatter {
	envUnformatter := EnvUnformatter{
		Version: SectionUnformatter{"version"},
		Db:      SectionUnformatter{"db"},
		Run:     SectionUnformatter{"run"},
		Deploy:  SectionUnformatter{"deploy"},
	}

	unformatter := &Unformatter{All: envUnformatter}

	switch env {
	case "qa":
		unformatter.Qa = envUnformatter
	case "staging":
		unformatter.Staging = envUnformatter
	case "production":
		unformatter.Production = envUnformatter
	case "test":
		unformatter.Test = envUnformatter
	case "uat":
		unformatter.Uat = envUnformatter
	case "development":
		fallthrough
	default:
		// Default to development.
		unformatter.Development = envUnformatter
	}
	return unformatter
}

// All possible top-level objects in the Forgefile.
type Unformatter struct {
	All         EnvUnformatter
	Development EnvUnformatter
	Test        EnvUnformatter
	Qa          EnvUnformatter
	Staging     EnvUnformatter
	Uat         EnvUnformatter
	Production  EnvUnformatter
}

// All possible tier-2 objects in the Forgefile.
type EnvUnformatter struct {
	Version SectionUnformatter
	Db      SectionUnformatter
	Run     SectionUnformatter
	Deploy  SectionUnformatter
}

// SectionUnformatter will read the configurations for any given subcommand
// into the registered command's SubConfig object.
type SectionUnformatter struct {
	cmdName string
}

func (sf *SectionUnformatter) UnmarshalYAML(unmarshal func(interface{}) error) error {
	cmd, ok := Registry[sf.cmdName]
	if !ok {
		return nil
	}
	return unmarshal(cmd.SubConf)
}
