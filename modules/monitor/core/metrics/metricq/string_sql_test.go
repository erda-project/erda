package metricq

import "testing"

func TestInterpolateParams(t *testing.T) {
	type args struct {
		query string
		args  []interface{}
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "1", args: args{"SELECT * FROM ? WHERE name='?'", []interface{}{"user", "Harry"}}, want: "SELECT * FROM user WHERE name='Harry'"},
		{name: "2", args: args{"SELECT * FROM ? Limit ?", []interface{}{"user", 10}}, want: "SELECT * FROM user Limit 10"},
		{name: "3", args: args{"SELECT sum(errors_sum),min(start_time_min),max(end_time_max),last(labels_distinct),trace_id::tag FROM trace WHERE terminus_key::tag=$terminus_key GROUP BY trace_id::tag ORDER BY max(start_time_min::field) LIMIT ?",
			[]interface{}{10}},
			want: "SELECT sum(errors_sum),min(start_time_min),max(end_time_max),last(labels_distinct),trace_id::tag FROM trace WHERE terminus_key::tag=$terminus_key GROUP BY trace_id::tag ORDER BY max(start_time_min::field) LIMIT 10"},
		{name: "4", args: args{"SELECT terminus_key::tag FROM ? WHERE terminus_key='?' AND service_name='?' LIMIT 1", []interface{}{"user", "test_tk", "test_service_id"}}, want: "SELECT terminus_key::tag FROM user WHERE terminus_key='test_tk' AND service_name='test_service_id' LIMIT 1"},
		{name: "5", args: args{"SELECT service_instance_id::tag,if(gt(now()-timestamp,300000000000),'false','true') FROM application_service_node WHERE terminus_key=$terminus_key AND service_name=$service_name AND service_id=$service_id GROUP BY service_instance_id::tag",
			[]interface{}{}},
			want: "SELECT service_instance_id::tag,if(gt(now()-timestamp,300000000000),'false','true') FROM application_service_node WHERE terminus_key=$terminus_key AND service_name=$service_name AND service_id=$service_id GROUP BY service_instance_id::tag"},
		{name: "6", args: args{"SELECT sum(count_sum),sum(elapsed_sum)/sum(count_sum) FROM ? WHERE ?='?' AND ?='?' AND ?='?'",
			[]interface{}{"user", "terminus_key", "test_tk", "service_name", "test_service_name", "service_id", "test_service_id"}},
			want: "SELECT sum(count_sum),sum(elapsed_sum)/sum(count_sum) FROM user WHERE terminus_key='test_tk' AND service_name='test_service_name' AND service_id='test_service_id'"},
		{name: "7", args: args{"SELECT sum(elapsed_count::field) FROM application_?_slow", []interface{}{"http"}}, want: "SELECT sum(elapsed_count::field) FROM application_http_slow"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := BuildStatement(tt.args.query, tt.args.args...)
			if err != nil {
				t.Errorf("BuildStatement() error = %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("BuildStatement() got = %v, want %v", got, tt.want)
			}
		})
	}
}
