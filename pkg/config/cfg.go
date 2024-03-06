package config

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	toml "github.com/BurntSushi/toml"
	"github.com/muesli/termenv"
	orderedmap "github.com/wk8/go-ordered-map/v2"
	yaml "gopkg.in/yaml.v3"
)

type OrderedMap = orderedmap.OrderedMap[string, string]

func iter(om *OrderedMap, callback func(key string, value string)) {
	current := om.Oldest()
	for {
		if current != nil {
			callback(current.Key, current.Value)
			current = current.Next()
		} else {
			break
		}
	}
}
func renderKv(buf *bytes.Buffer, prefix, key, sep, value string) {
	buf.WriteString(prefix) // comment &/ indent the line
	buf.WriteString(key)
	buf.WriteString(sep)
	buf.WriteString(value)
	buf.WriteRune('\n')
}

func renderYamlKv(buf *bytes.Buffer, prefix string, key string, value string) {
	renderKv(buf, prefix+"  - ", key, ": ", value)
}
func renderTomlKv(buf *bytes.Buffer, prefix string, key string, value string) {
	renderKv(buf, prefix+"  ", key, " = ", value)
}

func renderYaml(buf *bytes.Buffer, prefix string, header string, om *OrderedMap) {
	buf.WriteString(prefix) // comment &/ indent the line
	buf.WriteString(header + ":\n")
	current := om.Oldest()
	for {
		if current != nil {
			renderYamlKv(buf, prefix, current.Key, current.Value)
			current = current.Next()
		} else {
			break
		}
	}
}

func renderToml(buf *bytes.Buffer, prefix string, header string, om *OrderedMap) {
	buf.WriteString(prefix) // comment &/ indent the line
	buf.WriteString("[" + header + "]\n")
	current := om.Oldest()
	for {
		if current != nil {
			renderTomlKv(buf, prefix, current.Key, current.Value)
			current = current.Next()
		} else {
			break
		}
	}
}

func ZippedOrderedKeyValuePairs(om *OrderedMap) (keys []string, values []string) { // TODO: rename
	current := om.Oldest()
	for {
		if current != nil {
			keys = append(keys, current.Key)
			values = append(values, current.Value)
			current = current.Next()
		} else {
			break
		}
	}
	return
}

// const ExampleCfgFileHeader = `## commit_convention.yml
// ## omit the commit_types to use the default angular-style commit types`
// const ExampleCfgFileCommitTypes = `
// # commit_types:
// #   - type: description of what the short-form "type" means`
// const ExampleCfgFileScopes = `
// # scopes:
// #   - scope: description of what the short-form "scope" represents`
// const ExampleCfgFile = ExampleCfgFileHeader + ExampleCfgFileCommitTypes + ExampleCfgFileScopes

var (
	// see https://github.com/angular/angular.js/blob/master/DEVELOPERS.md#type
	// see https://github.com/conventional-changelog/commitlint/blob/master/%40commitlint/config-conventional/index.js#L23
	AngularCommitTypes *OrderedMap
	CentralStore       *Cfg
)

// instantiate the global more-or-less-constant AngularPresetCommitTypes
func angularCommitTypes() *OrderedMap {
	if AngularCommitTypes != nil {
		return AngularCommitTypes
	} else {
		om := orderedmap.New[string, string]()
		om.Set("feat", "adds a new feature")
		om.Set("fix", "fixes a bug")
		om.Set("docs", "changes only the documentation")
		om.Set("style", "changes the style but not the meaning of the code (such as formatting)")
		om.Set("perf", "improves performance")
		om.Set("test", "adds or corrects tests")
		om.Set("build", "changes the build system or external dependencies")
		om.Set("chore", "changes outside the code, docs, or tests")
		om.Set("ci", "changes to the Continuous Integration (CI) system")
		om.Set("refactor", "changes the code without changing behavior")
		om.Set("revert", "reverts prior changes")
		AngularCommitTypes = om
		return AngularCommitTypes
	}
}

const (
	HelpSubmit = "submit: tab/enter"
	HelpBack   = "go back: shift+tab"
	HelpCancel = "cancel: ctrl+c"
	HelpSelect = "navigate: up/down"
)

func Faint(s string) string {
	return termenv.String(s).Faint().String()
}

type Cfg struct {
	gitRepoRoot string
	gitDir      string
	// an empty string means no config file was used
	ConfigFile string
	// a custom, ordered map type is needed since maps fail to preserve the
	// insertion order of their keys: see https://go.dev/play/p/u0SB-LeqisU
	CommitTypes *OrderedMap
	Scopes      *OrderedMap
	// this caps the max len of the `type(scope): description`, not the body
	// naming inspired by conventional-changelog/commitlint
	HeaderMaxLength  int
	EnforceMaxLength bool
	DryRun           bool
}

func (c *Cfg) Clone() Cfg {
	commitTypes := orderedmap.New[string, string]()
	iter(c.CommitTypes, func(k string, v string) {
		commitTypes.Set(k, v)
	})
	scopes := orderedmap.New[string, string]()
	iter(c.Scopes, func(k string, v string) {
		scopes.Set(k, v)
	})
	return Cfg{
		gitRepoRoot:      c.gitRepoRoot,
		gitDir:           c.gitDir,
		ConfigFile:       c.ConfigFile,
		CommitTypes:      commitTypes,
		Scopes:           scopes,
		HeaderMaxLength:  c.HeaderMaxLength,
		EnforceMaxLength: c.EnforceMaxLength,
		DryRun:           c.DryRun,
	}
}

func (original *Cfg) merge(other *Cfg) {
	if other.ConfigFile != "" {
		original.ConfigFile = other.ConfigFile
	}
	if other.CommitTypes.Newest() != nil {
		original.CommitTypes = other.CommitTypes
	}
	if other.Scopes.Newest() != nil {
		original.Scopes = other.Scopes
	}
	original.EnforceMaxLength = other.EnforceMaxLength
	if other.HeaderMaxLength > 0 {
		original.HeaderMaxLength = other.HeaderMaxLength
	}
}

func ConstructDefaultFile(
	cfg *Cfg,
) string {
	buf := bytes.Buffer{}
	buf.WriteString("## omit the commit_types to use the default angular-style commit types\n")
	var render func(*bytes.Buffer, string, string, *OrderedMap)
	switch filepath.Ext(cfg.ConfigFile) {
	case ".yaml", ".yml":
		render = renderYaml
		break
	case ".toml":
		render = renderToml
		break
	default:
		log.Fatalf("unsupported default config file type: %s", cfg.ConfigFile)
	}

	render(&buf, "# ", "commit_types", cfg.CommitTypes)
	if cfg.Scopes == nil {
		cfg.Scopes = orderedmap.New[string, string]()
	}
	if cfg.Scopes.Len() > 0 {
		render(&buf, "", "scopes", cfg.Scopes)
	} else {
		cfg := cfg.Clone()
		cfg.Scopes.Set("scope", "description of what the short-form `scope` represents")
		render(&buf, "# ", "scopes", cfg.Scopes) // TODO: add a comment explaining how to add scopes
	}
	return buf.String()
}

// Find &/ read the configuration file into the passed config object
func (cfg *Cfg) ReadCfgFile(mustExist bool) (err error) {
	configFile := cfg.ConfigFile
	if configFile == "" {
		configFile, _, err = FindCCConfigFile(cfg.gitRepoRoot)
		if err != nil {
			if mustExist {
				return err
			} else {
				return nil // fall back to defaults
			}
		}
	}
	next, err := parseCCConfigurationFile(configFile)
	if err != nil {
		return err
	}
	cfg.merge(next)
	return err
}

// Initialize the global CentralStore of configuration.
func Init(dryRun bool) (*Cfg, error) {
	cfg := Cfg{
		CommitTypes:     angularCommitTypes(),
		Scopes:          orderedmap.New[string, string](),
		HeaderMaxLength: 72,
		//^ s.t. `git log --oneline` should remain within 80 columns w/ a 7-rune
		// commit hash and one space before the commit message.
		EnforceMaxLength: false,
		DryRun:           dryRun,
	}
	gitDir, err := getGitDir()
	if err != nil {
		if dryRun {
			// CentralStore.gitDir = "./.git"
		} else {
			// fatal since we need to be able to read/write .git/COMMIT_EDITMESSAGE
			return nil, err
		}
	}
	cfg.gitDir = gitDir
	repoRoot, err := GetGitRepoRoot()
	if err != nil {
		if dryRun {
			cfg.gitRepoRoot = "."
		} else {
			// fatal since we need to look for configuration there
			return nil, err
		}
	}
	cfg.gitRepoRoot = repoRoot
	if err := cfg.ReadCfgFile(false); err != nil {
		return nil, err
	}
	CentralStore = &cfg
	return CentralStore, nil
}

// turn []string, map[string]string, or []map[string]string into an OrderedMap
func toOrderedMap(raw interface{}) (om *OrderedMap, err error) {
	insert := func(om *orderedmap.OrderedMap[string, string], key string, value string) (err error) {
		if _, present := om.Set(key, value); present {
			err = fmt.Errorf("duplicate key: %s", key)
		}
		return
	}

	handleMap := func(om *orderedmap.OrderedMap[string, string], m map[string]interface{}) (err error) {
		// alphabetize the keys to keep output deterministic
		kvp := make([][2]string, 0, len(m))
		for k, v := range m {
			switch v2 := v.(type) {
			case string:
				kvp = append(kvp, [2]string{k, v2})
			default:
				err = fmt.Errorf("unexpected type: %+v", v2)
				return err
			}
		}
		sort.SliceStable(kvp, func(i, j int) bool {
			return kvp[i][0] < kvp[j][0]
		})
		for _, pair := range kvp {
			if err = insert(om, pair[0], pair[1]); err != nil {
				return err
			}
		}
		return err
	}

	switch intermediate1 := raw.(type) {
	case []interface{}:
		// guess the capacity to minimize allocations
		om = orderedmap.New[string, string](orderedmap.WithCapacity[string, string](len(intermediate1)))
		for _, intermediate2 := range intermediate1 {
			switch intermediate3 := intermediate2.(type) {
			case string:
				if _, present := om.Set(intermediate3, ""); present {
					return nil, fmt.Errorf("duplicate value: %s", intermediate3)
				}
			case map[string]interface{}:
				if err = handleMap(om, intermediate3); err != nil {
					return nil, err
				}
			default:
				err = fmt.Errorf("unknown value `%v`", intermediate3)
				return
			}
		}
		return
	case map[string]interface{}:
		om = orderedmap.New[string, string](orderedmap.WithCapacity[string, string](len(intermediate1)))
		if err = handleMap(om, intermediate1); err != nil {
			return
		}
		return
	case *orderedmap.OrderedMap[string, interface{}]:
		panic("..") // FIXME
	default:
		_ = intermediate1.(map[string]string)
		// for k, v := range i {
		// 	switch v2 := v {
		// 	case string:
		// 		break
		// 	default:
		// 		panic(fmt.Errorf("unexpected type '%s' for key %s: '%+v'", reflect.TypeOf(intermediate1).Name(), k, v2))
		// 	}
		// }
		return nil, fmt.Errorf("unexpected format: %+v => %+v", intermediate1, reflect.TypeOf(intermediate1).Name())
	}
}

// func parsePackageJson(data []byte) (*Cfg, error) {
// 	om := orderedmap.New[string, interface{}]() // :/
// 	if err := om.UnmarshalJSON(data); err != nil {
// 		return nil, err
// 	}
// 	var cfg Cfg
// 	if raw, present := om.Get("git-cc"); present {
// 		switch section := raw.(type) {
// 		case orderedmap.OrderedMap[string, interface{}]:
// 			// extract configuration from val
// 			if rawScopes, ok := section.Get("scopes"); ok {
// 				switch intermediate := rawScopes.(type) {
// 				case []interface{}:
// 					om, err:=toOrderedMap(rawScopes)
// 				}
// 				cfg.Scopes = rawScopes.(*orderedmap.OrderedMap[string, string])
// 			}
// 				if rawTypes, ok := section.Get("commit_types"); ok {
// 					types, err := toOrderedMap(rawTypes)
// 					if err != nil {
// 						return nil, err
// 					}
// 					cfg.CommitTypes = types
// 				}
// 				if maxLen, ok := raw["header_max_length"]; ok {
// 					switch max := maxLen.(type) {
// 					case int:
// 						cfg.HeaderMaxLength = max
// 					default:
// 						return nil, fmt.Errorf("unexpected type of value \"header_max_length\" in %s: `%+v`", configFile, max)
// 					}
// 				}
// 				if enforcedLen, ok := raw["enforce_header_max_length"]; ok {
// 					switch enforced := enforcedLen.(type) {
// 					case bool:
// 						cfg.EnforceMaxLength = enforced
// 					default:
// 						return nil, fmt.Errorf("Unexpected type for \"header_max_length_enforced\" in %s: `%+v`", configFile, enforcedLen)
// 					}
// 				}
// 			}
// 		}
// 		panic("FIXME")
// 		return &cfg, nil
// 	}
// 	return nil, fmt.Errorf("key \"git-cc\" missing from package.json")
// }

func parseCCConfigurationFile(configFile string) (*Cfg, error) {
	f, err := os.Stat(configFile)
	if err != nil {
		return nil, err
	}
	if f.IsDir() {
		return nil, fmt.Errorf("found directory `%s`", configFile)
	}
	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, err
	}
	name := f.Name()
	// if name == "package.json" {
	// 	// allowed as a special case. Otherwise, prefer writing configuration
	// 	// in a format that allows comments
	// 	return parsePackageJson(data)
	// }
	var raw map[string]interface{}
	ext := filepath.Ext(name)
	switch ext {
	case ".yaml", ".yml": // FIXME: order not preserved in {[string]: string} maps
		if err = yaml.Unmarshal(data, &raw); err != nil {
			return nil, err
		}
	case ".toml":
		if err = toml.Unmarshal(data, &raw); err != nil {
			return nil, err
		}
	// case ".json":
	// 	return nil, fmt.Errorf("only package.json supported, %s found", configFile)
	default:
		// all file extensions should already be known when searching for config
		// files
		panic("Unsupported config file type: " + configFile)
	}

	var cfg Cfg
	if rawScopes, ok := raw["scopes"]; ok {
		scopes, err := toOrderedMap(rawScopes)
		if err != nil {
			return nil, err
		}
		cfg.Scopes = scopes
	}
	if rawTypes, present := raw["commit_types"]; present {
		types, err := toOrderedMap(rawTypes)
		if err != nil {
			return nil, err
		}
		cfg.CommitTypes = types
	}
	if maxLen, present := raw["header_max_length"]; present {
		switch max := maxLen.(type) {
		case int:
			cfg.HeaderMaxLength = max
		default:
			return nil, fmt.Errorf("unexpected type of value \"header_max_length\" in %s: `%+v`", configFile, max)
		}
	}
	if enforcedLen, present := raw["enforce_header_max_length"]; present {
		switch enforced := enforcedLen.(type) {
		case bool:
			cfg.EnforceMaxLength = enforced
		default:
			return nil, fmt.Errorf("unexpected type for \"header_max_length_enforced\" in %s: `%+v`", configFile, enforcedLen)
		}
	}

	cfg.ConfigFile = configFile // always an absolute path
	return &cfg, nil
}

func FindCCConfigFile(gitRepoRoot string) (string, []string, error) {
	// pkgMeta := map[string]map[string]interface{}{} // cache the unmarshalled package.json/pyproject.toml for reuse
	candidateFiles := [...]string{
		"commit_convention.yaml",
		"commit_convention.yml",
		"commit_convention.toml",
		// TODO: support commitlint config
		// ".commitlintrc",
		// ".commitlintrc.json",
		// ".commitlintrc.yaml",
		// ".commitlintrc.yml",
		// TODO: handle .commitlintrc.{j,t,cj,ct}s, commitlint.config.{j,t,cj,ct}s
		// "package.json",
		// "pyproject.toml",
	}
	dirsToSearch := make([]string, 0, 3)

	cwd, err := filepath.Abs(".")
	if err == nil {
		dirsToSearch = append(dirsToSearch, cwd)
	}
	if gitRepoRoot != "" {
		if gitRepoRoot != cwd {
			dirsToSearch = append(dirsToSearch, gitRepoRoot)
		}
		dotConfigDir := path.Join(gitRepoRoot, ".config")
		if dirInfo, err := os.Stat(dotConfigDir); err == nil {
			if dirInfo.IsDir() {
				dirsToSearch = append(dirsToSearch, dotConfigDir)
			}
		}
	}
	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		configHome = os.Getenv("HOME")
	}
	if configHome != "" {
		dirsToSearch = append(dirsToSearch, configHome)
	}
	tried := make([]string, 0, len(candidateFiles)*len(dirsToSearch))
	for _, dir := range dirsToSearch {
		for _, candidate := range candidateFiles {
			configFile := path.Join(dir, candidate)
			_, err := os.Stat(configFile)
			if err == nil {
				return configFile, tried, nil
			} else {
				tried = append(tried, configFile)
			}
		}
	}
	return "", tried, fmt.Errorf("no configuration found in \n  - %s", strings.Join(tried, "\n  - "))
}

// find the root of the tree that git is working on
func GetGitRepoRoot() (string, error) {
	if env := os.Getenv("GIT_WORK_TREE"); env != "" {
		// there might be a `$GIT_COMMON_DIR?`
		return env, nil
	}
	out, err := stdoutFrom("git", "rev-parse", "--show-toplevel")
	if err != nil {
		return "", err
	} else {
		return strings.TrimRight(out, "\r\n"), nil
	}
}

// find the git directory (usually ./.git)
func getGitDir() (string, error) {
	if env := os.Getenv("GIT_COMMON_DIR"); env != "" {
		return env, nil
	}
	if env := os.Getenv("GIT_DIR"); env != "" {
		return env, nil
	}
	out, err := stdoutFrom("git", "rev-parse", "--absolute-git-dir")
	if err != nil {
		return "", err
	}
	return out, nil
}

func stdoutFrom(args ...string) (string, error) {
	cmd := exec.Command(args[0], args[1:]...)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	return out.String(), err
}

func getGitVar(var_name string) (string, error) {
	out, err := stdoutFrom("git", "var", var_name)
	if err != nil {
		return "", err
	} else {
		return strings.TrimRight(out, " \t\r\n"), err
	}
}

func GetEditor() string {
	editor := os.Getenv("EDITOR")
	if editor != "" {
		return editor
	}
	var err error
	editor, err = getGitVar("GIT_EDITOR")
	if err != nil {
		return editor
	}
	editor = "vi"
	_, err = exec.LookPath(editor)
	if err != nil {
		msg := "unable to open the fallback editor"
		hint := "hint: set the EDITOR env variable or install vi"
		log.Fatalf(fmt.Sprintf("%s: %q\n%s\n", msg, editor, hint))
	}
	return editor
}

func GetCommitMessageFile() string {
	out := CentralStore.gitDir
	return strings.Join(
		[]string{strings.TrimRight(out, " \t\r\n"), "COMMIT_EDITMSG"},
		string(os.PathSeparator),
	)
}

func InitDefaultCfgFile(cfg *Cfg, format string) error {
	if cfg.ConfigFile != "" && !cfg.DryRun {
		return fmt.Errorf("config file already exists: %s", cfg.ConfigFile)
	} else {
		dir := path.Join(cfg.gitRepoRoot, ".config")
		repoPermissions, err := os.Stat(cfg.gitRepoRoot)
		if err != nil {
			log.Fatalf("unable to stat directory %s: %v", cfg.gitRepoRoot, err)
		}
		if _, err := os.Stat(dir); err != nil {
			if os.IsNotExist(err) && !cfg.DryRun {
				err = os.Mkdir(dir, repoPermissions.Mode().Perm())
				// use the same permissions as the repo root
				if err != nil {
					log.Fatalf("unable to create directory %s: %v", dir, err)
				}
			} else {
				log.Fatalf("unable to stat directory %s: %v", dir, err)
			}
		}
		cfg.ConfigFile = path.Join(dir, "commit_convention."+format)
		contents := ConstructDefaultFile(cfg)
		if cfg.DryRun {
			fmt.Println(contents)
		} else {
			f, err := os.Create(cfg.ConfigFile)
			if err != nil {
				log.Fatalf("unable to create file %s: '%+v'", cfg.ConfigFile, err)
			}
			_, err = f.WriteString(contents)
			if err != nil {
				log.Fatalf("unable to write to file: %v", err)
			}
			err = f.Close()
			if err != nil {
				log.Fatalf("unable to close file: %s : %v", cfg.ConfigFile, err)
			}
		}
	}
	return nil
}

// interactively edit the config file, if any was used.
func EditCfgFileCmd(cfg *Cfg) *exec.Cmd {
	editCmd := []string{}
	// sometimes `$EDITOR` can be a script with spaces, like `code --wait`
	// TODO: handle quotes in `$EDITOR`?
	for _, part := range strings.Split(GetEditor(), " ") {
		if part != "" {
			editCmd = append(editCmd, part)
		}
	}
	if cfg.ConfigFile == "" {
		InitDefaultCfgFile(cfg, "yaml")
	}
	editCmd = append(editCmd, cfg.ConfigFile)
	cmd := exec.Command(editCmd[0], editCmd[1:]...)
	return cmd
}
