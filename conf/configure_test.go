package conf

import (
	"embed"
	"strings"
	"testing"
)

//go:embed monitor/monitor/chartview
var systemChartviewFS embed.FS

//go:embed monitor/monitor/notify
var notifyChartviewFS embed.FS

func TestConfigure_JsonReader(t *testing.T) {
	type fields struct {
		FS      embed.FS
		Dirname string
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{name: "case1", fields: fields{
			FS:      systemChartviewFS,
			Dirname: "monitor/monitor/chartview",
		}, want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Configure{
				FS:      tt.fields.FS,
				Dirname: tt.fields.Dirname,
			}
			got := c.JsonReader()
			if len(*got) > 0 != tt.want {
				t.Errorf("JsonReader() = %v, want %v", got, tt.want)
			}
			for _, file := range *got {
				if !strings.Contains(JsonFileExtension, file.Extension) {
					t.Errorf("JsonReader() read unknow file extension %s, %s", file.Extension, file.Filename)
				}
			}
		})
	}
}

func TestConfigure_YamlOrYmlReader(t *testing.T) {
	type fields struct {
		FS      embed.FS
		Dirname string
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{name: "case1", fields: fields{
			FS:      notifyChartviewFS,
			Dirname: "monitor/monitor/notify",
		}, want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Configure{
				FS:      tt.fields.FS,
				Dirname: tt.fields.Dirname,
			}
			got := c.YamlOrYmlReader()
			if len(*got) > 0 != tt.want {
				t.Errorf("YamlOrYmlReader() = %v, want %v", got, tt.want)
			}
			for _, file := range *got {
				if !strings.Contains(YamlOrYmlFileExtension, file.Extension) {
					t.Errorf("YamlOrYmlReader() read unknow file extension %s, %s", file.Extension, file.Filename)
				}
			}
		})
	}
}

func Test_reader(t *testing.T) {
	type args struct {
		fs            embed.FS
		dirname       string
		fileExtension string
		files         *ConfigurationSearcher
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "case1",
			args: args{
				fs:            systemChartviewFS,
				dirname:       "monitor/monitor/chartview",
				fileExtension: JsonFileExtension,
				files:         &ConfigurationSearcher{},
			},
		},
		{
			name: "case2",
			args: args{
				fs:            notifyChartviewFS,
				dirname:       "monitor/monitor/notify",
				fileExtension: YamlOrYmlFileExtension,
				files:         &ConfigurationSearcher{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader(tt.args.fs, tt.args.dirname, tt.args.fileExtension, tt.args.files)
			files := *tt.args.files
			if len(files) <= 0 {
				t.Errorf("reader() fail.")
			}
		})
	}
}
