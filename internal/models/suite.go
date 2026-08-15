package models

type DocumentKind string

const (
	KindRoot     DocumentKind = "root"
	KindDefaults DocumentKind = "defaults"
	KindSteps    DocumentKind = "steps"
)

// Suite is the parsed tree anchored by the selected working directory.
type Suite struct {
	WorkDir string
	Root    *Directory
}

type Directory struct {
	Stage              int
	Path               string
	Parent             *Directory
	Children           []*Directory
	Files              []*File
	DefaultsFile       *File
	StepsFiles         []*File
	DefaultsDefinition *DefaultsDefinition
	StepsDefinitions   []*StepsDefinition
	ResolvedDefaults   Defaults
	ResolvedSteps      []Step
}

type File struct {
	Stage     string `yaml:"stage"`
	Path      string
	Kind      DocumentKind
	Directory *Directory
	Bytes     []byte
	//Root            *RootFile
	//Defaults        *DefaultsFile
	//Steps           *StepsFile
	//ParentDefaults  *Defaults
	//RuntimeDefaults RuntimeDefaults
	//RuntimeSteps    []RuntimeStep
}

type BaseDefinition struct {
	App  string     `yaml:"app"`
	Kind string     `yaml:"kind"`
	Spec YAMLString `yaml:"spec"`
}

type DefaultsDefinition struct {
	App      string       `yaml:"app"`
	Kind     DocumentKind `yaml:"kind"`
	Metadata Metadata     `yaml:"metadata"`
	Spec     Defaults     `yaml:"spec"`
	File     *File        `yaml:"-"`
}

type StepsDefinition struct {
	App      string       `yaml:"app"`
	Kind     DocumentKind `yaml:"kind"`
	Metadata Metadata     `yaml:"metadata"`
	Spec     struct {
		Steps []Step `yaml:"steps"`
	} `yaml:"spec"`
	File *File `yaml:"-"`
}

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

type YAMLString string
