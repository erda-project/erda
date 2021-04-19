// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"fmt"
	"os"
	"reflect"
	"runtime"
	"strings"
	"text/template"

	"github.com/erda-project/erda/tools/cli/command"
)

//go:generate go run collect/collect.go
//go:generate go run generate.go collectCMDs.go validate.go
func main() {
	outputPath := "../../generated_cmd/"
	if err := os.RemoveAll(outputPath); err != nil {
		panic(err)
	}
	if err := os.Mkdir(outputPath, 0777); err != nil {
		panic(err)
	}
	for idx, cmd := range CMDs {
		if err := validate(cmd, CMDNames[idx]); err != nil {
			panic(err)
		}
		var buf strings.Builder
		arg := genTemplateArg(cmd, CMDNames[idx])
		if err := CMDtemplate.Execute(&buf, arg); err != nil {
			panic(err)
		}

		f, err := os.OpenFile(outputPath+CMDNames[idx]+".go", os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0666)
		if err != nil {
			panic(err)
		}
		f.WriteString(buf.String())
		fmt.Printf("generated [%s]\n", CMDNames[idx])
	}
}

func genTemplateArg(cmd command.Command, cmdname string) templateArg {
	runcmd := ""
	if cmd.Run != nil {
		splited := strings.Split(runtime.FuncForPC(reflect.ValueOf(cmd.Run).Pointer()).Name(), "/")
		runcmd = splited[len(splited)-1]
	}
	parent := "command.RootCmd"
	if cmd.ParentName != "" {
		parent = cmd.ParentName + "Cmd"
	}
	argmin, argmax := argNum(cmd.Args)
	r := templateArg{
		Imports:        imports(cmd),
		Usage:          genUsage(cmd),
		ShortHelp:      cmd.ShortHelp,
		LongHelp:       cmd.LongHelp,
		Example:        cmd.Example,
		Hidden:         cmd.Hidden,
		DontHideCursor: cmd.DontHideCursor,
		Args:           cmd.Args,
		ArgsNumMin:     argmin,
		ArgsNumMax:     argmax,
		Flags:          cmd.Flags,
		Name:           cmdname,
		RunCMD:         runcmd,
		ParentCmd:      parent,

		ArgType:        argType,
		ArgConvertType: argConvertType,
		FlagType:       flagType,
		Underline:      underline,
	}
	return r
}

func genUsage(cmd command.Command) string {
	args := []string{}
	for _, arg := range cmd.Args {
		if arg.IsOption() {
			args = append(args, "["+arg.GetName()+"]")
		} else {
			args = append(args, "<"+arg.GetName()+">")
		}
	}
	return fmt.Sprintf("%s %s", cmd.Name, strings.Join(args, " "))
}
func argNum(args []command.Arg) (int, int) {
	var min, max int
	for _, arg := range args {
		if !arg.IsOption() {
			min++
		}
		max++
	}
	return min, max
}
func argConvertType(arg command.Arg) string {
	return arg.ConvertType()
}
func argType(arg command.Arg) string {
	return reflect.TypeOf(arg).String()
}

func flagType(flag command.Flag) string {
	switch flag.(type) {
	case command.IntFlag:
		return "int"
	case command.StringFlag:
		return "string"
	case command.BoolFlag:
		return "bool"
	case command.FloatFlag:
		return "float64"
	case command.IPFlag:
		return "net.IP"
	case command.StringListFlag:
		return "[]string"
	default:
		panic("should not be here")
	}

}

func imports(cmd command.Command) []string {
	r := []string{
		"github.com/spf13/cobra",
		"github.com/erda-project/erda/tools/cli/command",
	}
	if needNet(cmd) {
		r = append(r, "net")
	}
	if needCmd(cmd) {
		r = append(r, "github.com/erda-project/erda/tools/cli/cmd")
		r = append(r, "github.com/erda-project/erda/tools/cli/translate")
		r = append(r, "github.com/erda-project/erda/pkg/terminal/color_str")
		r = append(r, "fmt")
	}
	return r
}

type templateArg struct {
	Imports        []string
	Usage          string
	Hidden         bool
	DontHideCursor bool
	ShortHelp      string
	LongHelp       string
	Example        string
	Args           []command.Arg
	ArgsNumMin     int
	ArgsNumMax     int
	Flags          []command.Flag
	Name           string
	RunCMD         string
	ParentCmd      string

	ArgType        func(command.Arg) string
	ArgConvertType func(command.Arg) string
	FlagType       func(command.Flag) string
	Underline      func(string) string
}

var CMDtemplate = template.Must(template.New("spec").Parse(`// GENERATED FILE, DO NOT EDIT
package cmd

import (
{{- range $_, $v := .Imports }}
	"{{$v}}"
{{- end}}
)

var {{.Name}}Cmd = &cobra.Command{
	Use:   "{{.Usage}}",
	{{- if .ShortHelp}}
	Short: ` + "`{{.ShortHelp}}`" + `,
	{{- end}}
	{{- if .LongHelp}}
	Long: ` + "`{{.LongHelp}}`" + `,
	{{- end}}
	Example: ` + "`{{.Example}}`" + `,
	{{- if .RunCMD}}
	Args:  cobra.RangeArgs({{$.ArgsNumMin}}, {{$.ArgsNumMax}}),
	Hidden: {{.Hidden}},
	RunE: func(_ *cobra.Command, args []string) error {
		var err error
		defer func(){
			if err != nil {
				fmt.Println(color_str.Red("✗ "), translate.Translate(err))
			}
		}()
		{{.Name}}ctx = command.GetContext()
		{{- range $idx, $v := .Args}}
		if len(args)-1 >= {{$idx}} {
			if err = ({{call $.ArgType $v}}{}).Validate({{$idx}}, args[{{$idx}}]); err != nil { return err }
			{{$.Name}}Arg{{$idx}}, _ = (({{call $.ArgType $v}}{}).Convert(args[{{$idx}}])).({{call $.ArgConvertType $v}})
		}
		{{- end}}
		{{- if $.DontHideCursor}}
		command.Tput("cvvis")
		defer command.Tput("civis")
		{{- end}}
		err = {{.RunCMD}}({{.Name}}ctx{{range $idx, $v := .Args}}, {{$.Name}}Arg{{$idx}}{{end}}{{range $_, $v := .Flags}}, {{$.Name}}{{call $.Underline $v.Name}}Flag{{end}})
		return err
	},
	{{- end}}
}

var (
	{{.Name}}ctx *command.Context

{{- range $_, $v := .Flags}}
	{{$.Name}}{{call $.Underline $v.Name}}Flag {{(call $.FlagType $v)}}
{{- end}}

{{- range $idx, $v := .Args}}
	{{$.Name}}Arg{{$idx}} {{call $.ArgConvertType $v}}
{{- end}}
)

func init() {
{{- range $_, $v := .Flags}}
	{{- if eq (call $.FlagType $v) "string"}}
	{{$.Name}}Cmd.Flags().StringVarP(&{{$.Name}}{{call $.Underline $v.Name}}Flag, "{{$v.Name}}", "{{$v.Short}}", "{{$v.DefaultValue}}", "{{$v.Doc}}")
	{{- else if eq (call $.FlagType $v) "bool"}}
	{{$.Name}}Cmd.Flags().BoolVarP(&{{$.Name}}{{call $.Underline $v.Name}}Flag, "{{$v.Name}}", "{{$v.Short}}", {{$v.DefaultValue}}, "{{$v.Doc}}")
	{{- else if eq (call $.FlagType $v) "int"}}
	{{$.Name}}Cmd.Flags().IntVarP(&{{$.Name}}{{call $.Underline $v.Name}}Flag, "{{$v.Name}}", "{{$v.Short}}", {{$v.DefaultValue}}, "{{$v.Doc}}")
	{{- else if eq (call $.FlagType $v) "float64"}}
	{{$.Name}}Cmd.Flags().Float64VarP(&{{$.Name}}{{call $.Underline $v.Name}}Flag, "{{$v.Name}}", "{{$v.Short}}", {{$v.DefaultValue}}, "{{$v.Doc}}")
	{{- else if eq (call $.FlagType $v) "net.IP"}}
	{{$.Name}}Cmd.Flags().IPVarP(&{{$.Name}}{{call $.Underline $v.Name}}Flag, "{{$v.Name}}", "{{$v.Short}}", net.IP([]byte("{{$v.DefaultValue}}")), "{{$v.Doc}}")
	{{- else if eq (call $.FlagType $v) "[]string"}}
	{{$.Name}}Cmd.Flags().StringSliceVarP(&{{$.Name}}{{call $.Underline $v.Name}}Flag, "{{$v.Name}}", "{{$v.Short}}", {{$v.DefaultV}}, "{{$v.Doc}}")
	{{- end}}
{{end}}
	{{.ParentCmd}}.AddCommand({{.Name}}Cmd)
}

`))

func needNet(cmd command.Command) bool {
	for _, arg := range cmd.Args {
		if _, ok := arg.(command.IPArg); ok {
			return true
		}
	}
	for _, flag := range cmd.Flags {
		if _, ok := flag.(command.IPFlag); ok {
			return true
		}
	}
	return false
}

func needCmd(cmd command.Command) bool {
	return cmd.Run != nil
}

func underline(s string) string {
	return strings.Replace(s, "-", "_", -1)
}
