package cmd

import (
	"fmt"
	"io"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/jenkins-x/golang-jenkins"
	"github.com/jenkins-x/jx/pkg/jx/cmd/templates"
	cmdutil "github.com/jenkins-x/jx/pkg/jx/cmd/util"
	"github.com/jenkins-x/jx/pkg/util"
)

// StartPipelineOptions contains the command line options
type StartPipelineOptions struct {
	GetOptions

	Tail   bool
	Filter string

	Jobs map[string]*gojenkins.Job
}

var (
	start_pipeline_long = templates.LongDesc(`
		Starts the pipeline build.

`)

	start_pipeline_example = templates.Examples(`
		# Start a pipeline
		jx start pipeline foo

		# Select the pipeline to start
		jx start pipeline

		# Select the pipeline to start and tail the log
		jx start pipeline -t
	`)
)

// NewCmdStartPipeline creates the command
func NewCmdStartPipeline(f cmdutil.Factory, out io.Writer, errOut io.Writer) *cobra.Command {
	options := &StartPipelineOptions{
		GetOptions: GetOptions{
			CommonOptions: CommonOptions{
				Factory: f,
				Out:     out,
				Err:     errOut,
			},
		},
	}

	cmd := &cobra.Command{
		Use:        "pipeline [flags]",
		Short:      "Starts one or more pipelines",
		Long:       start_pipeline_long,
		Example:    start_pipeline_example,
		Aliases:    []string{"pipe", "pipeline"},
		SuggestFor: []string{"run", "build"},
		Run: func(cmd *cobra.Command, args []string) {
			options.Cmd = cmd
			options.Args = args
			err := options.Run()
			cmdutil.CheckErr(err)
		},
	}
	cmd.Flags().BoolVarP(&options.Tail, "tail", "t", false, "Tails the build log to the current terminal")
	cmd.Flags().StringVarP(&options.Filter, "filter", "f", "", "Filters all the available jobs by those that contain the given text")

	return cmd
}

// Run implements this command
func (o *StartPipelineOptions) Run() error {
	jenkins, err := o.JenkinsClient()
	if err != nil {
		return err
	}
	jobs, err := jenkins.GetJobs()
	if err != nil {
		return err
	}
	o.Jobs = map[string]*gojenkins.Job{}
	o.addJobs("", jobs)

	args := o.Args
	names := []string{}
	for k, _ := range o.Jobs {
		names = append(names, k)
	}
	sort.Strings(names)

	if len(args) == 0 {
		defaultName := ""
		for _, n := range names {
			if strings.HasSuffix(n, "/master") {
				defaultName = n
				break
			}
		}
		name, err := util.PickNameWithDefault(names, "Which pipelines do you want to start: ", defaultName)
		if err != nil {
			return err
		}
		args = []string{name}
	}
	for _, a := range args {
		err = o.startJob(a, names)
		if err != nil {
			return err
		}
	}
	return nil
}

func (o *StartPipelineOptions) startJob(name string, allNames []string) error {
	job := o.Jobs[name]
	if job == nil {
		return util.InvalidArg(name, allNames)
	}
	jenkins, err := o.JenkinsClient()
	if err != nil {
		return err
	}

	// ignore errors as it could be there's no last build yet
	previous, _ := jenkins.GetLastBuild(*job)

	params := url.Values{}
	err = jenkins.Build(*job, params)
	if err != nil {
		return err
	}

	i := 0
	for {
		last, err := jenkins.GetLastBuild(*job)

		// lets ignore the first query in case there's no build yet
		if i > 0 && err != nil {
			return err
		}
		i++

		if last.Number != previous.Number {
			o.Printf("Started build of %s at %s\n", util.ColorInfo(name), util.ColorInfo(last.Url))
			o.Printf("%s %s\n", util.ColorStatus("view the log at:"), util.ColorInfo(util.UrlJoin(last.Url, "/console")))
			if o.Tail {
				return o.tailBuild(name, &last)
			}
			return nil
		}
		time.Sleep(time.Second)
	}
}

func jobName(prefix string, j *gojenkins.Job) string {
	name := j.FullName
	if name == "" {
		name = j.Name
	}
	if prefix != "" {
		name = prefix + "/" + name
	}
	return name
}

func (o *StartPipelineOptions) addJobs(prefix string, jobs []gojenkins.Job) {
	jenkins, err := o.JenkinsClient()
	if err != nil {
		return
	}
	for _, j := range jobs {
		name := jobName(prefix, &j)

		if IsPipeline(&j) {
			if o.Filter == "" || strings.Contains(name, o.Filter) {
				o.Jobs[name] = &j
			}
		}
		if j.Jobs != nil {
			o.addJobs(name, j.Jobs)
		} else {
			job, err := jenkins.GetJob(name)
			if err == nil && job.Jobs != nil {
				o.addJobs(name, job.Jobs)
			}
		}
	}
}
func (o *StartPipelineOptions) tailBuild(jobName string, build *gojenkins.Build) error {
	jenkins, err := o.JenkinsClient()
	if err != nil {
		return nil
	}

	u, err := url.Parse(build.Url)
	if err != nil {
		return err
	}
	buildPath := u.Path
	o.Printf("%s %s\n", util.ColorStatus("tailing the log of"), util.ColorInfo(fmt.Sprintf("%s #%d", jobName, build.Number)))
	return jenkins.TailLog(buildPath, o.Out, time.Second, time.Hour*100)
}

func IsPipeline(j *gojenkins.Job) bool {
	return strings.Contains(j.Class, "Job")
}
