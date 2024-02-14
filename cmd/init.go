package cmd

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime/debug"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/manifoldco/promptui"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"github.com/teamkeel/keel/codegen"
	"github.com/teamkeel/keel/colors"
	"github.com/teamkeel/keel/node"
	"github.com/teamkeel/keel/schema"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initializes a new Keel project",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		defer panicHandler()

		state := &InitState{}
		steps := []func(state *InitState){
			initState,
			initStepDir,
			initStepTemplate,
			initStepPackageManager,
			initStepGit,
			initStepCreateProject,
		}

		printLogo()
		fmt.Println(" Welcome to Keel!")
		fmt.Println(colors.Gray(" Let's build something great"))
		fmt.Println("")

		for _, step := range steps {
			step(state)
			fmt.Println("")
		}

		fmt.Println("Your new Keel project is ready to roll ✨")
		fmt.Println("")
	},
}

func panicHandler() {
	if r := recover(); r != nil {
		if err, ok := r.(error); ok && err == promptui.ErrInterrupt {
			fmt.Println("| Aborting...")
			fmt.Println("")
			return
		}

		errStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("1"))

		fmt.Println("")
		fmt.Println(errStyle.Render("======= Oh no =========="))
		fmt.Println("Something seems to have gone wrong.")
		fmt.Println("This is likely a bug with Keel - please let us know via:")
		fmt.Println(" - Discord (https://discord.gg/HV8g38nBnm)")
		fmt.Println(" - GitHub Issue (https://github.com/teamkeel/keel/issues/new)")
		fmt.Println("")
		fmt.Println("Please include the following stack trace in your report:")
		fmt.Println(colors.Gray(string(debug.Stack())))
		fmt.Println(errStyle.Render("========================"))
		fmt.Println("")
	}
}

type InitState struct {
	cwd            string
	gitRoot        string
	targetDir      string
	initGitRepo    bool
	files          codegen.GeneratedFiles
	packageManager string
}

func initState(state *InitState) {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	state.cwd = wd

	c := exec.Command("git", "rev-parse", "--show-toplevel")
	out, err := c.Output()
	if err == nil {
		state.gitRoot = strings.TrimSpace(string(out))
	}
}

func initStepDir(state *InitState) {
	initSectionHeading("Directory")

	entries, err := os.ReadDir(state.cwd)
	if err != nil {
		panic(err)
	}

	defaultDir := ""

	switch {
	case len(entries) == 0:
		defaultDir = "."
	case lo.ContainsBy(entries, func(e fs.DirEntry) bool { return e.Name() == "package.json" }):
		defaultDir = "./keel"
	default:
		defaultDir = "./my-keel-app"
	}

	prompt := promptui.Prompt{
		Label:     "Where should we create your project?",
		Default:   defaultDir,
		AllowEdit: false,
		Validate: func(v string) error {
			e, _ := os.ReadDir(v)
			if len(e) > 0 {
				return errors.New("directory is not empty")
			}
			return nil
		},
		Pointer: promptui.PipeCursor,
	}

	dir, err := prompt.Run()
	if err != nil {
		panic(err)
	}

	state.targetDir = dir
}

func initStepTemplate(state *InitState) {
	initSectionHeading("Template")

	optionBlank := "Blank project"
	optionStarter := "Starter template"

	template := promptui.Select{
		Label: "How would you like to start your new project?",
		Items: []string{optionStarter, optionBlank},
		Templates: &promptui.SelectTemplates{
			Label:    "{{ . }}",
			Selected: fmt.Sprintf("%s {{ . | bold }} ", promptui.IconGood),
			Active:   "{{ . | cyan }}",
			Inactive: "{{ . }}",
		},
	}

	_, result, err := template.Run()
	if err != nil {
		panic(err)
	}

	if result == optionBlank {
		state.files = append(state.files, &codegen.GeneratedFile{
			Path: ".gitignore",
			Contents: `node_modules/
.DS_Store
*.local

# Keel
.build/
			`,
		})

		state.files = append(state.files, &codegen.GeneratedFile{
			Path:     "schema.keel",
			Contents: "// Visit https://docs.keel.so/ for documentation on how to get started",
		})

		state.files = append(state.files, &codegen.GeneratedFile{
			Path: "keelconfig.yaml",
			Contents: `
# Visit https://docs.keel.so/authentication/getting-started for more information about authentication
auth:

# Visit https://docs.keel.so/envvars for more information about environment variables
environment:

# Visit https://docs.keel.so/secrets for more information about secrets
secrets:
`,
		})
		return
	}

	res, err := http.Get("https://api.github.com/repos/teamkeel/starter-templates/zipball/main")
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	zipBytes, err := io.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}

	type StarterTemplate struct {
		Name string `json:"name"`
		Path string `json:"path"`
	}

	starterFiles := map[string][]byte{}
	templates := []*StarterTemplate{}

	zipReader, err := zip.NewReader(bytes.NewReader(zipBytes), int64(len(zipBytes)))
	if err != nil {
		panic(err)
	}

	for _, f := range zipReader.File {
		if f.FileInfo().IsDir() {
			continue
		}

		r, err := f.Open()
		if err != nil {
			panic(err)
		}

		b, err := io.ReadAll(r)
		if err != nil {
			panic(err)
		}

		// The zip archive contains a directory which contains the repo contents. We don't
		// care about the top-level directory so we drop it from the name
		name := filepath.Join(strings.Split(f.Name, "/")[1:]...)

		starterFiles[name] = b

		if name == "templates.json" {
			err = json.Unmarshal(b, &templates)
			if err != nil {
				panic(err)
			}
		}
	}

	templateNames := lo.Map(templates, func(v *StarterTemplate, _ int) string {
		return v.Name
	})

	starter := promptui.Select{
		Label: "Which template would you like to use?",
		Items: templateNames,
		Templates: &promptui.SelectTemplates{
			Label:    "{{ . }}",
			Selected: fmt.Sprintf("%s {{ . | bold }} ", promptui.IconGood),
			Active:   "{{ . | cyan }}",
			Inactive: "{{ . }}",
		},
	}

	idx, _, err := starter.Run()
	if err != nil {
		panic(err)
	}

	for k, v := range starterFiles {
		if strings.HasPrefix(k, templates[idx].Path) {
			path := strings.TrimPrefix(k, templates[idx].Path+"/")
			state.files = append(state.files, &codegen.GeneratedFile{
				Path:     path,
				Contents: string(v),
			})
		}
	}
}

func initStepPackageManager(state *InitState) {
	initSectionHeading("Package Manager")

	rootDir := state.cwd
	if state.gitRoot != "" {
		rootDir = state.gitRoot
	}

	packageManager, err := resolvePackageManager(rootDir, true)
	if err != nil {
		panic(err)
	}

	state.packageManager = packageManager
}

func initStepGit(state *InitState) {
	initSectionHeading("Version Control")

	if state.gitRoot != "" {
		printSuccess(fmt.Sprintf("Git repo detected: %s", colors.Gray(state.gitRoot).String()))
		return
	}

	starter := promptui.Prompt{
		Label:     "Should we initialise a Git repo in your new project?",
		IsConfirm: true,
	}

	_, err := starter.Run()
	confirmed := !errors.Is(err, promptui.ErrAbort)
	if err != nil && confirmed {
		panic(err)
	}

	state.initGitRepo = confirmed
}

func initStepCreateProject(state *InitState) {
	initSectionHeading("Generating Project")

	err := state.files.Write(state.targetDir)
	if err != nil {
		panic(err)
	}

	err = node.Bootstrap(
		state.targetDir,
		node.WithPackagesPath(flagNodePackagesPath),
		node.WithPackageManager(state.packageManager),
		node.WithLogger(func(s string) {
			fmt.Println("|", colors.Gray(s))
		}),
		node.WithOutputWriter(os.Stdout),
	)
	if err != nil {
		panic(err)
	}

	b := schema.Builder{}
	schema, err := b.MakeFromDirectory(state.targetDir)
	if err != nil {
		panic(err)
	}

	files, err := node.Generate(context.Background(), schema, node.WithDevelopmentServer(true))
	if err != nil {
		panic(err)
	}

	err = files.Write(state.targetDir)
	if err != nil {
		panic(err)
	}

	fmt.Println("| Generated @teamkeel/sdk")
	fmt.Println("| Generated @teamkeel/testing")

	if state.initGitRepo {
		fmt.Println("| Initialising git repo")
		c := exec.Command("git", "init")
		c.Dir = state.targetDir
		err := c.Run()
		if err != nil {
			panic(err)
		}
	}
}

func initSectionHeading(v string) {
	style := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("15")).
		Background(lipgloss.Color("#7D56F4")).
		PaddingLeft(1).
		PaddingRight(1)

	fmt.Println(style.Render(v))
}

func printSuccess(v string) {
	fmt.Println(colors.Green("✔"), v)
}

func printLogo() {
	logo := `	                                       
                  bbbb                 
               bbbbbbbbb               
            bbbbbbbbbbbbbb            
              bbbbbbbbbbbbbb          
       y        bbbbbbbbbbbbbb       
    yyyyyyy       bbbbbbbbbbbbbb     
  yyyyyyyyyyy       bbbbbbbbbbbbbb  
  ooooooooooo      ppppppppppppppp  
    ooooooo      ppppppppppppppp     
       o       ppppppppppppppp       
             ppppppppppppppp          
            pppppppppppppp            
               ppppppppp               
                  pppp                
	`

	blue := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#0094FF")).Bold(true)

	purple := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#AB84FF")).Bold(true)

	yellow := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFBA17")).Bold(true)

	orange := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF866F")).Bold(true)

	logo = strings.ReplaceAll(logo, "b", blue.Render("●"))
	logo = strings.ReplaceAll(logo, "p", purple.Render("●"))
	logo = strings.ReplaceAll(logo, "y", yellow.Render("●"))
	logo = strings.ReplaceAll(logo, "o", orange.Render("●"))

	fmt.Println(logo)
}

func init() {
	rootCmd.AddCommand(initCmd)

	if enabledDebugFlags == "true" {
		initCmd.Flags().StringVar(&flagNodePackagesPath, "node-packages-path", "", "path to local @teamkeel npm packages")
	}
}
