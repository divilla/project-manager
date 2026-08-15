package models

type DocumentKind string

const (
	KindRoot     DocumentKind = "root"
	KindDefaults DocumentKind = "defaults"
	KindSteps    DocumentKind = "steps"
)

type Metadata struct {
	Name   string   `yaml:"name"`
	Labels []string `yaml:"labels"`
}

type Defaults struct {
	BaseURL  string            `yaml:"baseUrl"`
	BasePath string            `yaml:"basePath"`
	Headers  map[string]string `yaml:"headers"`
	Timeout  int               `yaml:"timeout"`
	Retries  int               `yaml:"retries"`
}

type RuntimeDefaults struct {
	BaseURL  string
	BasePath string
	Headers  map[string]string
	Timeout  int
	Retries  int
}

type RootFile struct {
	App      string       `yaml:"app"`
	Kind     DocumentKind `yaml:"kind"`
	Metadata Metadata     `yaml:"metadata"`
	Spec     Defaults     `yaml:"spec"`
}

type DefaultsFile struct {
	App      string       `yaml:"app"`
	Kind     DocumentKind `yaml:"kind"`
	Metadata Metadata     `yaml:"metadata"`
	Spec     Defaults     `yaml:"spec"`
}

type StepsFile struct {
	App      string       `yaml:"app"`
	Kind     DocumentKind `yaml:"kind"`
	Metadata Metadata     `yaml:"metadata"`
	Spec     struct {
		Steps []Step `yaml:"steps"`
	} `yaml:"spec"`
}

type File struct {
	FilePath        string
	Kind            DocumentKind
	Root            *RootFile
	Defaults        *DefaultsFile
	Steps           *StepsFile
	ParentDefaults  *Defaults
	RuntimeDefaults RuntimeDefaults
	RuntimeSteps    []RuntimeStep
}

type Directory struct {
	Stage           int
	Dir             string
	RuntimeDefaults RuntimeDefaults
	Files           []*File
	Children        []*Directory
}

// Suite is the parsed tree anchored by the selected working directory.
type Suite struct {
	WorkDir string
	Root    *Directory
}

type Step struct {
	Vars    map[string]any `yaml:"vars"`
	Request struct {
		Method   string            `yaml:"method"`
		BaseURL  string            `yaml:"baseUrl"`
		BasePath string            `yaml:"basePath"`
		Path     string            `yaml:"path"`
		Headers  map[string]string `yaml:"headers"`
		Timeout  int               `yaml:"timeout"`
		Retries  int               `yaml:"retries"`
		Query    string            `yaml:"query"`
		Body     YAMLString        `yaml:"body"`
	} `yaml:"request"`
	Response struct {
		Status   []int               `yaml:"status"`
		Capture  map[string]string   `yaml:"capture"`
		Types    map[string][]string `yaml:"types"`
		Expected YAMLString          `yaml:"expected"`
	} `yaml:"response"`
}

type RuntimeStep Step

type YAMLString string
