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
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"github.com/mitchellh/mapstructure"
	toml "github.com/pelletier/go-toml"
	"github.com/spf13/cobra"
	yaml "gopkg.in/yaml.v2"
)

var keyReg, _ = regexp.Compile("{{[a-zA-Z0-9_ \t]+}}")

func findKeys(title, content string) []string {
	keySet := map[string]struct{}{}
	for _, key := range keyReg.FindAllString(title, -1) {
		keySet[strings.TrimSpace(key)] = struct{}{}
	}
	for _, key := range keyReg.FindAllString(content, -1) {
		keySet[strings.TrimSpace(key)] = struct{}{}
	}
	var keys []string
	for k := range keySet {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

var apiCmd = &cobra.Command{
	Use:   "api",
	Short: "convert to api http file",
	Run: func(cmd *cobra.Command, args []string) {
		ext := "json"
		dir, _ := cmd.Flags().GetString("dir")
		if strings.TrimSpace("/") == dir {
			fmt.Println("dir must not be /") // 保护跟目录
			os.Exit(1)
		}
		if ext != "" {
			ext = "." + ext
		}
		envsLines := readFileLines(".env")
		envs := map[string]string{}
		for _, line := range envsLines {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) < 2 {
				continue
			}
			envs[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
		type Group struct {
			Output *os.File
			List   []map[string]interface{}
		}
		groups := map[string]*Group{}
		defer func() {
			for _, g := range groups {
				if g.Output != nil {
					g.Output.Close()
				}
			}
		}()
		filepath.Walk(dir, func(p string, info os.FileInfo, err error) error {
			if info.IsDir() || err != nil {
				return nil
			}
			if strings.Contains(p, "vendor") || strings.Contains(p, ".git") {
				return nil
			}
			if !strings.HasSuffix(info.Name(), ext) {
				return nil
			}
			alert_type, alert_index := getAlertTypeIndexFromPath(p)
			group, _ := groups[alert_type]
			if group == nil {
				group = &Group{}
				groups[alert_type] = group
			}

			if group.Output == nil {
				outfile := path.Join(filepath.Dir(p), "api-test.http")
				output, err := os.Create(outfile)
				if err != nil {
					fmt.Printf("fail to create output file %s : %s\n", path.Join(dir, "api-test.http"), err)
					os.Exit(1)
				}
				group.Output = output
			}
			rows, err := getDataList(p)
			if err != nil {
				fmt.Println(p, err)
				os.Exit(1)
			}
			for _, row := range rows {
				row["alert_index"] = alert_index
				group.List = append(group.List, row)
			}
			return nil
		})
		type Templete struct {
			Window int `json:"window"`
			Funcs  []struct {
				Aggregator  string      `json:"aggregator"`
				Field       string      `json:"field"`
				FieldScript string      `json:"field_script"`
				Operator    string      `json:"operator"`
				Value       interface{} `json:"value"`
			} `json:"functions"`
		}
		for alert_type, group := range groups {

			groupId := envs["notify_group_id"]
			apiURL := envs["api_url"]
			dingding := envs["url"]
			clusterName := envs["cluster_name"]

			org_id := envs["org_id"]
			user_id := envs["user_id"]
			tenant_group := envs["tenant_group"]
			domain := envs["domain"]

			notify_group_id := 1
			if groupId != "" {
				notify_group_id, _ = strconv.Atoi(groupId)
			}

			output := group.Output
			output.WriteString("@url = ")
			output.WriteString(apiURL)
			output.WriteString("\n")

			output.WriteString("@notify_group_id = ")
			output.WriteString(groupId)
			output.WriteString("\n")

			output.WriteString("@dingding = ")
			output.WriteString(dingding)
			output.WriteString("\n")

			output.WriteString("@org_id = ")
			output.WriteString(org_id)
			output.WriteString("\n")

			output.WriteString("@user_id = ")
			output.WriteString(user_id)
			output.WriteString("\n")

			output.WriteString("@domain = ")
			output.WriteString(domain)
			output.WriteString("\n")

			output.WriteString("\n\n### 删除 alert\n")
			output.WriteString("DELETE {{url}}/api/alerts/<id>")
			output.WriteString("\n")

			for _, row := range group.List {
				if row["template"] == nil || fmt.Sprint(row["template"]) == "" {
					continue
				}
				alertScope := fmt.Sprint(row["alert_scope"])
				json.Unmarshal([]byte(alertScope), &alertScope)
				alertIndex := fmt.Sprint(row["alert_index"])
				// json.Unmarshal([]byte(alertIndex), &alertIndex)
				template := fmt.Sprint(row["template"])
				err := json.Unmarshal([]byte(template), &template)
				if err != nil {
					fmt.Println(alert_type, alertIndex, err)
					return
				}
				var temp Templete
				err = json.Unmarshal([]byte(template), &temp)
				if err != nil {
					fmt.Println(alert_type, alertIndex, err)
					return
				}
				var funcs []map[string]interface{}
				for _, fn := range temp.Funcs {
					if fn.Operator == "" {
						continue
					}
					funcs = append(funcs, map[string]interface{}{
						"field":      fn.Field,
						"aggregator": fn.Aggregator,
						"operator":   fn.Operator,
						"value":      fn.Value,
					})
				}

				name := fmt.Sprint(row["name"])
				json.Unmarshal([]byte(name), &name)
				body := map[string]interface{}{
					"name":        name,
					"clusterName": clusterName,
					"alertScope":  alertScope,
					"enable":      true,
					"rules": []interface{}{
						map[string]interface{}{
							"alertIndex": alertIndex,  // 表达式定义
							"window":     temp.Window, // 持续周期
							"functions":  funcs,
						},
					},
				}
				if alertScope == "org" {
					body["alertScopeId"] = org_id
					body["domain"] = domain
					body["notifies"] = []interface{}{
						map[string]interface{}{
							"type":      "notify_group",  // 类型，现支持notify_group（通知组）和dingding（钉钉）
							"groupId":   notify_group_id, // 通知组 ID
							"groupType": "dingding",      // 通知类型，与通知组的targets中的type一致
						},
					}
				} else {
					body["alertScopeId"] = tenant_group
					body["notifies"] = []interface{}{
						map[string]interface{}{
							"type":        "dingding",
							"dingdingUrl": "{{dingding}}", // 钉钉地址
						},
					}
				}

				output.WriteString("\n\n### ")
				output.WriteString(name)
				output.WriteString("\n# @ ")
				output.WriteString(alert_type)
				output.WriteString("/")
				output.WriteString(alertIndex)
				if alertScope == "org" {
					output.WriteString("\nPOST {{url}}/api/orgs/alerts\nContent-Type: application/json\nOrg-ID: {{org_id}}\nUser-ID: {{user_id}}\n\n")
				} else {
					output.WriteString("\nPOST {{url}}/api/alerts\nContent-Type: application/json\n\n")
				}

				output.WriteString(MarshalAndIntend(body))
				output.WriteString("\n\n")
			}
			fmt.Printf("write %s ok . --------------------------------- \n", group.Output.Name())
		}
		fmt.Println("ok")
	},
}

var httpCmd = &cobra.Command{
	Use:   "http",
	Short: "convert to http file",
	Run: func(cmd *cobra.Command, args []string) {
		ext := "yml"
		target := "dingding"
		dir, _ := cmd.Flags().GetString("dir")
		if strings.TrimSpace("/") == dir {
			fmt.Println("dir must not be /") // 保护跟目录
			os.Exit(1)
		}
		if ext != "" {
			ext = "." + ext
		}
		envsLines := readFileLines(".env")
		envs := map[string]string{}
		for _, line := range envsLines {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) < 2 {
				continue
			}
			envs[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
		type Group struct {
			Output *os.File
			List   []map[string]interface{}
			KeySet map[string]struct{}
		}
		groups := map[string]*Group{}
		defer func() {
			for _, g := range groups {
				if g.Output != nil {
					g.Output.Close()
				}
			}
		}()
		filepath.Walk(dir, func(p string, info os.FileInfo, err error) error {
			if info.IsDir() || err != nil {
				return nil
			}
			if strings.Contains(p, "vendor") || strings.Contains(p, ".git") {
				return nil
			}
			if !strings.HasSuffix(info.Name(), ext) {
				return nil
			}
			alert_type, alert_index := getAlertTypeIndexFromPath(p)
			group, _ := groups[alert_type]
			if group == nil {
				group = &Group{
					KeySet: map[string]struct{}{},
				}
				groups[alert_type] = group
			}

			if group.Output == nil {
				outfile := path.Join(filepath.Dir(p), target+".http")
				output, err := os.Create(outfile)
				if err != nil {
					fmt.Printf("fail to create output file %s : %s\n", path.Join(dir, target+".http"), err)
					os.Exit(1)
				}
				group.Output = output
			}
			rows, err := getDataList(p)
			if err != nil {
				fmt.Println(p, err)
				os.Exit(1)
			}
			trimKeyFunc := func(r rune) bool {
				ok := unicode.IsSpace(r)
				if ok {
					return ok
				}
				return r == '{' || r == '}'
			}
			for _, row := range rows {
				if !strings.Contains(fmt.Sprint(row["target"]), target) {
					continue
				}
				title := fmt.Sprint(row["title"])
				template := fmt.Sprint(row["template"])
				for _, key := range keyReg.FindAllString(title, -1) {
					group.KeySet[strings.TrimFunc(key, trimKeyFunc)] = struct{}{}
				}
				for _, key := range keyReg.FindAllString(template, -1) {
					group.KeySet[strings.TrimFunc(key, trimKeyFunc)] = struct{}{}
				}
				row["alert_index"] = alert_index
				group.List = append(group.List, row)
			}
			return nil
		})
		for alert_type, group := range groups {
			var keys []string
			for k := range group.KeySet {
				if k != "" {
					keys = append(keys, k)
				}
			}
			sort.Strings(keys)

			output := group.Output
			output.WriteString("@url = ")
			output.WriteString(envs["url"])
			output.WriteString("\n")
			for _, key := range keys {
				if key == "url" {
					continue
				}
				output.WriteString("@")
				output.WriteString(key)
				output.WriteString(" = ")
				output.WriteString(envs[key])
				output.WriteString("\n")
			}

			for _, row := range group.List {
				name := decodeString(fmt.Sprint(row["name"]))
				alert_index := fmt.Sprint(row["alert_index"])
				trigger := decodeString(fmt.Sprint(row["trigger"]))
				template := decodeString(fmt.Sprint(row["template"]))
				title := decodeString(fmt.Sprint(row["title"]))

				output.WriteString("\n\n### ")
				output.WriteString(name)
				output.WriteString("\n# @ ")
				output.WriteString(alert_type)
				output.WriteString("/")
				output.WriteString(alert_index)
				output.WriteString("/")
				output.WriteString(trigger)
				output.WriteString("\nPOST {{url}}\nContent-Type: application/json\n\n")

				msg := map[string]interface{}{
					"msgtype": "markdown",
					"markdown": map[string]interface{}{
						"title": title,
						"text":  template,
					},
				}
				output.WriteString(MarshalAndIntend(msg))
				output.WriteString("\n")
			}
			fmt.Printf("write %s ok . --------------------------------- \n", group.Output.Name())
		}
		fmt.Println("ok")
	},
}

var sqlCmd = &cobra.Command{
	Use:   "sql",
	Short: "convert json/yaml/toml to insert sql",
	Run:   buildSQLCommand,
}

func buildSQLCommand(cmd *cobra.Command, args []string) {
	builder := newSQLBuilder(cmd)
	createDistDirIfNotExists("dist")
	builder.Build()
}

type sqlBuilder struct {
	extension   string
	directory   string
	filepath    string
	tableName   string
	alertEnable bool
}

func (bc *sqlBuilder) Build() {
	if bc.fileMode() {
		bc.buildWithFile()
	} else if bc.dirMode() {
		bc.buildWithDirectory()
	} else {
		log.Println("you must specify directory or file")
	}
	fmt.Println("build sql successfully! Please check in dist/")
}

func (bc *sqlBuilder) fileMode() bool {
	return bc.filepath != "" && bc.directory == ""
}

func (bc *sqlBuilder) dirMode() bool {
	return bc.filepath == "" && bc.directory != ""
}

func (bc *sqlBuilder) createSqlFile() *os.File {
	output, err := os.Create(path.Join("dist", bc.tableName+".sql"))
	if err != nil {
		log.Printf("fail to create output file %s : %s\n", path.Join(bc.directory, bc.tableName+".sql"), err)
		os.Exit(1)
	}
	return output
}

func (bc *sqlBuilder) buildWithFile() {
	output := bc.createSqlFile()
	defer output.Close()

	output.WriteString("SET NAMES utf8mb4;\n")
	output.WriteString("BEGIN;\n")
	output.WriteString("UPDATE `" + bc.tableName + "` SET `enable` = 0;\n")
	sqls, err := convertToSQL(bc.filepath, bc.tableName, bc.alertEnable)
	if err != nil {
		log.Printf("convert failed. err: %s, filepath: %s\n", err, bc.filepath)
		os.Exit(1)
	}
	for _, sql := range sqls {
		fmt.Print(sql)
		output.WriteString(sql)
	}
	output.WriteString("COMMIT;\n")
}

func (bc *sqlBuilder) buildWithDirectory() {
	output := bc.createSqlFile()
	defer output.Close()

	output.WriteString("SET NAMES utf8mb4;\n")
	output.WriteString("BEGIN;\n")
	output.WriteString("UPDATE `" + bc.tableName + "` SET `enable` = 0;\n")

	filepath.Walk(bc.directory, func(p string, info os.FileInfo, err error) error {
		if info.IsDir() || err != nil || strings.Contains(p, "vendor") {
			return nil
		}
		if !strings.HasSuffix(info.Name(), bc.extension) {
			return nil
		}
		sqls, err := convertToSQL(p, bc.tableName, bc.alertEnable)
		if err != nil {
			log.Printf("convert failed. err: %s, filepath: %s\n", err, bc.filepath)
			os.Exit(1)
		}
		for _, sql := range sqls {
			output.WriteString(sql)
		}
		return nil
	})
	output.WriteString("COMMIT;\n")
}

func newSQLBuilder(cmd *cobra.Command) *sqlBuilder {
	ext, _ := cmd.Flags().GetString("ext")
	dir, _ := cmd.Flags().GetString("dir")
	fp, _ := cmd.Flags().GetString("file")
	table, _ := cmd.Flags().GetString("table")
	alert, _ := cmd.Flags().GetBool("alert")

	// 保护跟目录
	if strings.TrimSpace(dir) == "/" {
		fmt.Println("dir must not be /")
		os.Exit(1)
	}
	if ext != "" {
		ext = "." + ext
	}
	return &sqlBuilder{
		extension:   ext,
		directory:   dir,
		filepath:    fp,
		tableName:   table,
		alertEnable: alert,
	}
}

var rootCmd = &cobra.Command{
	Use:   "convert",
	Short: "Convert json/yaml/toml to insert sql",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func convertVal(data interface{}) string {
	if data == nil {
		return ""
	}
	typ := reflect.TypeOf(data)
	val := reflect.ValueOf(data)
	switch typ.Kind() {
	case reflect.Struct, reflect.Slice, reflect.Array, reflect.Map:
		data = toJsonData(data)
		byts, _ := json.Marshal(data)
		byts, _ = json.Marshal(string(byts))
		return string(byts)
	case reflect.String:
		byts, _ := json.Marshal(data)
		return string(byts)
	case reflect.Ptr:
		return convertVal(val.Elem().Interface())
	}
	return fmt.Sprint(data)
}

type Values []reflect.Value

func (s Values) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s Values) Len() int      { return len(s) }
func (s Values) Less(i, j int) bool {
	return strings.Compare(fmt.Sprint(s[i]), fmt.Sprint(s[j])) < 0
}

func convertToSQL(file, table string, alert bool) ([]string, error) {
	rows, err := getDataList(file)
	if err != nil {
		return nil, err
	}
	// force alert_type alert_index from path
	var (
		alert_type, alert_index string
	)
	if alert {
		alert_type, alert_index = getAlertTypeIndexFromPath(file)
	}

	var output []string
	write := func(keys []reflect.Value, value reflect.Value) {
		buf1 := &bytes.Buffer{}
		buf2 := &bytes.Buffer{}
		for _, key := range keys {
			val := value.MapIndex(key)
			buf1.WriteString("`" + fmt.Sprint(key.Interface()) + "`")
			buf1.WriteString(",")
			buf2.WriteString(fmt.Sprint(val))
			buf2.WriteString(",")
		}
		fields := string(buf1.Bytes()[0 : buf1.Len()-1])
		values := string(buf2.Bytes()[0 : buf2.Len()-1])
		output = append(output, fmt.Sprintf("INSERT `%s`(%s) VALUES(%s);\n", table, fields, values))
	}
	for _, row := range rows {
		if len(row) < 0 {
			continue
		}
		if alert {
			// force alert_type alert_index from path
			row["alert_type"] = "\"" + alert_type + "\""
			row["alert_index"] = "\"" + alert_index + "\""
		}
		value := reflect.ValueOf(row)
		keys := value.MapKeys()
		sort.Sort(Values(keys))

		if target, ok := row["target"]; ok {
			targets := strings.Split(fmt.Sprint(target), ",")
			if len(targets) > 1 {
				for _, tar := range strings.Split(strings.Join(targets, "\",\""), ",") {
					row["target"] = tar
					write(keys, value)
				}
				continue
			}
		}
		write(keys, value)
	}
	return output, nil
}

func getDataList(file string) ([]map[string]interface{}, error) {
	convert := func(data interface{}) map[string]interface{} {
		value := reflect.ValueOf(data)
		if value.Kind() != reflect.Map {
			fmt.Printf("invalid data type %v \n", reflect.TypeOf(data))
			os.Exit(1)
		}

		keys := value.MapKeys()
		if len(keys) == 0 {
			return nil
		}
		out := map[string]interface{}{}
		for _, key := range keys {
			val := value.MapIndex(key)
			out[fmt.Sprint(key.Interface())] = convertVal(val.Interface())
		}
		return out
	}
	data := readFile(file)
	if data == nil {
		return nil, nil
	}
	var output []map[string]interface{}
	switch data := data.(type) {
	case []interface{}:
		for _, item := range data {
			row := convert(item)
			if row != nil {
				output = append(output, row)
			}
		}
	case map[string]interface{}:
		row := convert(data)
		if row != nil {
			output = append(output, row)
		}
	default:
		fmt.Printf("invalid file format: %s, %v\n", file, data)
		os.Exit(1)
	}
	return output, nil
}

func bindDefaultSelectPair(data map[string]interface{}) {
	var expressionObj interface{}
	if v, ok := data["expression"]; ok {
		expressionObj = v
	} else if v, ok := data["template"]; ok {
		expressionObj = v
	} else {
		return
	}

	selectObj := expressionObj.(map[string]interface{})["select"]
	selectMap := selectObj.(map[string]interface{})
	selectMap["_meta"] = "#_meta"
	selectMap["_metric_scope"] = "#_metric_scope"
	selectMap["_metric_scope_id"] = "#_metric_scope_id"
	selectMap["org_name"] = "#org_name"
}

// 解决 map[interface{}]interface{} 无法json序列化的问题
func toJsonData(data interface{}) interface{} {
	if data == nil {
		return data
	}
	typ := reflect.TypeOf(data)
	val := reflect.ValueOf(data)
	switch typ.Kind() {
	case reflect.Map:
		out := make(map[string]interface{})
		keys := val.MapKeys()
		for _, key := range keys {
			strKey := fmt.Sprint(key.Interface())
			value := val.MapIndex(key)
			out[strKey] = toJsonData(value.Interface())
		}
		return out
	case reflect.Slice:
		len := val.Len()
		var slice []interface{}
		for i := 0; i < len; i++ {
			slice = append(slice, toJsonData(val.Index(i).Interface()))
		}
		return slice
	case reflect.Array:
		len := val.Len()
		typ := reflect.ArrayOf(len, typ.Elem())
		array := reflect.New(typ)
		array = array.Elem()
		for i := 0; i < len; i++ {
			array.Index(i).Set(reflect.ValueOf(toJsonData(val.Index(i).Interface())))
		}
		return array.Interface()
	case reflect.Ptr:
		return toJsonData(val.Elem().Interface())
	case reflect.Struct:
		var out map[string]interface{}
		decoder, _ := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
			Metadata:         nil,
			Result:           &out,
			WeaklyTypedInput: true,
			TagName:          "json",
			DecodeHook: mapstructure.ComposeDecodeHookFunc(
				mapstructure.StringToTimeDurationHookFunc(),
				mapstructure.StringToSliceHookFunc(","),
			),
		})
		decoder.Decode(data)
		return toJsonData(out)
	}
	return val.Interface()
}

func readFile(file string) interface{} {
	exts := []string{"json", "yaml", "yml", "toml"}
	for _, ext := range exts {
		if !strings.HasSuffix(file, ext) {
			continue
		}
		reader, ok := fileReaders[ext]
		if !ok {
			fmt.Printf("not exit %s file readder \n", ext)
			os.Exit(1)
		}
		_, err := os.Stat(file)
		if err != nil {
			fmt.Println("get file failed. err: ", err)
			continue
		}
		byts, err := ioutil.ReadFile(file)
		if err != nil {
			fmt.Println(file, err)
			os.Exit(1)
		}
		data, err := reader(byts)
		if err != nil {
			fmt.Println(file, err)
			os.Exit(1)
		}
		return data
	}
	return nil
}

var fileReaders = map[string]func([]byte) (interface{}, error){
	"json": func(byts []byte) (interface{}, error) {
		var out interface{}
		err := json.Unmarshal(byts, &out)
		return out, err
	},
	"yaml": readYaml,
	"yml":  readYaml,
	"toml": func(byts []byte) (interface{}, error) {
		var out interface{}
		err := toml.Unmarshal(byts, &out)
		return out, err
	},
}

func getAlertTypeIndexFromPath(p string) (string, string) {
	parts := strings.Split(p, "/")
	if len(parts) < 2 {
		fmt.Printf("invalid path %s\n", p)
		os.Exit(1)
	}
	idx := strings.LastIndex(parts[len(parts)-1], ".")
	if idx < 0 {
		fmt.Printf("invalid path %s\n", p)
		os.Exit(1)
	}
	return parts[len(parts)-2], parts[len(parts)-1][0:idx]
}

func readYaml(byts []byte) (interface{}, error) {
	var out interface{}
	err := yaml.Unmarshal(byts, &out)
	return out, err
}

func readFileLines(name string) []string {
	byts, err := ioutil.ReadFile(name)
	if err != nil {
		return nil
	}
	text := strings.TrimSpace(string(byts))
	if len(text) == 0 {
		return nil
	}
	return strings.Split(text, "\n")
}

func decodeString(text string) string {
	var out string
	json.Unmarshal([]byte(text), &out)
	return out
}

func MarshalAndIntend(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	var out bytes.Buffer
	err = json.Indent(&out, b, "", "\t")
	if err != nil {
		panic(err)
	}
	return out.String()
}

func createDistDirIfNotExists(dist string) {
	_, err := os.Stat(dist)
	if os.IsNotExist(err) {
		os.MkdirAll(dist, os.ModePerm)
	}
}

func init() {
	rootCmd.AddCommand(sqlCmd)
	rootCmd.AddCommand(httpCmd)
	rootCmd.AddCommand(apiCmd)
	sqlCmd.Flags().String("ext", "yml", "convert .yml file")
	sqlCmd.Flags().String("table", "tablename", "table for sql output")
	sqlCmd.Flags().String("dir", "", "directory to convert")
	sqlCmd.Flags().String("file", "", "the file to convert")
	sqlCmd.Flags().Bool("alert", false, "generate alertType and alertIndex")
	sqlCmd.Flags().String("type", ".", "alert type to set enable 0")

	httpCmd.Flags().String("dir", ".", "directory to convert")
	apiCmd.Flags().String("dir", ".", "directory to convert")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
