package config

type (
	Wrapper struct {
		// original file path
		FilePath string
		// Config in first parent that contains it for `config` type or `config` in current dir
		// or first parent containing it for `steps` files
		ParentConfig *Config
		// Effective config for both steps and config files
		RuntimeConfig Config
		// Steps populated with Runtime config values
		RuntimeSteps []Step
	}

	// WrapperGroup is Dir organized structure
	WrapperGroup struct {
		// Stage 0 is starting dir - all children are stage 1 and so one
		// Stage with same nr is executed in paralle
		// Different stages execute sequentaly from 0 on...
		Stage int
		// Original Dir
		Dir string
		// Runtime Config passed to all steps files
		RuntimeConfig *Config
		// Same dir
		Wrappers []*Wrapper
		// Child dirs
		WrapperGroups []*WrapperGroup
	}

	Config struct {
		App      string `yaml:"app"`
		Kind     string `yaml:"kind"`
		Metadata struct {
			Name   string   `yaml:"name"`
			Labels []string `yaml:"labels"`
		} `yaml:"metadata"`
		Spec struct {
			BaseURL  string            `yaml:"baseUrl"`
			BasePath string            `yaml:"basePath"`
			Headers  map[string]string `yaml:"headers"`
			Steps    []Step            `yaml:"steps"`
		}
	}

	Step struct {
		Request struct {
			Vars     map[string]interface{} `yaml:"vars"`
			Method   string                 `yaml:"method"`
			BaseURL  string                 `yaml:"baseUrl"`
			BasePath string                 `yaml:"basePath"`
			Path     string                 `yaml:"path"`
			Headers  map[string]string      `yaml:"headers"`
			Query    string                 `yaml:"query"`
			Body     YAMLString             `yaml:"body"`
		} `yaml:"request"`
		Response struct {
			Vars     map[string]string `yaml:"setVars"`
			Types    map[string]string `yaml:"setVars"`
			Expected YAMLString        `yaml:"expected"`
		}
	}

	Parser struct {
		WorkDir string
		Files   []string
	}

	YAMLString string
)

func NewParser(workDir string, files []string) *Parser {
	return &Parser{
		WorkDir: workDir,
		Files:   files,
	}
}

func (p *Parser) Parse() *WrapperGroup {
	return &WrapperGroup{}
}
