package bufbuild

import (
	"flag"
	"log"
	"path"
	"path/filepath"
	"strings"

	"github.com/bazelbuild/bazel-gazelle/label"
	"github.com/stackb/rules_proto/pkg/protoc"
)

func init() {
	protoc.Plugins().MustRegisterPlugin(&ConnectGoProto{})
}

// ConnectGoProto implements Plugin for the bufbuild/connect-go plugin.
type ConnectGoProto struct{}

// Name implements part of the Plugin interface.
func (p *ConnectGoProto) Name() string {
	return "bufbuild:connect-go"
}

// Configure implements part of the Plugin interface.
func (p *ConnectGoProto) Configure(ctx *protoc.PluginContext) *protoc.PluginConfiguration {
	flags := parseConnectGoProtoOptions(p.Name(), ctx.PluginConfig.GetFlags())
	imports := make(map[string]bool)
	for _, file := range ctx.ProtoLibrary.Files() {
		for _, imp := range file.Imports() {
			imports[imp.Filename] = true
		}
	}
	// TODO: get target option from directive
	var options = []string{"keep_empty_files=true"}
	goFiles := make([]string, 0)
	for _, file := range ctx.ProtoLibrary.Files() {
		goFile := file.Name + "_connect.pb.go"
		if flags.excludeOutput[filepath.Base(goFile)] {
			continue
		}
		if ctx.Rel != "" {
			goFile = path.Join(ctx.Rel, goFile)
		}
		goFiles = append(goFiles, goFile)
	}

	pc := &protoc.PluginConfiguration{
		Label:   label.New("build_stack_rules_proto", "plugin/bufbuild", "connect-go"),
		Outputs: protoc.DeduplicateAndSort(goFiles),
		Options: protoc.DeduplicateAndSort(options),
	}
	if len(pc.Outputs) == 0 {
		pc.Outputs = nil
	}
	return pc
}

// ConnectGoProtoOptions represents the parsed flag configuration for the
// ConnectGoProto implementation.
type ConnectGoProtoOptions struct {
	excludeOutput map[string]bool
}

func parseConnectGoProtoOptions(kindName string, args []string) *ConnectGoProtoOptions {
	flags := flag.NewFlagSet(kindName, flag.ExitOnError)

	var excludeOutput string
	flags.StringVar(&excludeOutput, "exclude_output", "", "--exclude_output=foo_connect.pb.go suppresses the file 'foo_connect.pb.go' from the output list")

	if err := flags.Parse(args); err != nil {
		log.Fatalf("failed to parse flags for %q: %v", kindName, err)
	}
	config := &ConnectGoProtoOptions{
		excludeOutput: make(map[string]bool),
	}
	for _, value := range strings.Split(excludeOutput, ",") {
		config.excludeOutput[value] = true
	}

	return config
}
