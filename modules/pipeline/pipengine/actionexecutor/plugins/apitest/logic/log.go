package logic

import (
	"os"
	"reflect"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
)

var log = newLogger()

type actionLogFormatter struct {
	logrus.TextFormatter
}

func (f *actionLogFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	_bytes, err := f.TextFormatter.Format(entry)
	if err != nil {
		return nil, err
	}
	return append([]byte("[Action Log] "), _bytes...), nil
}

func newLogger() *logrus.Logger {
	// set logrus
	l := logrus.New()
	l.SetFormatter(&actionLogFormatter{
		logrus.TextFormatter{
			ForceColors:            true,
			DisableTimestamp:       true,
			DisableLevelTruncation: true,
		},
	})
	l.SetOutput(os.Stdout)
	return l
}

func addLineDelimiter(prefix ...string) {
	var _prefix string
	if len(prefix) > 0 {
		_prefix = prefix[0]
	}
	log.Printf("%s==========", _prefix)
}

func addNewLine(num ...int) {
	_num := 1
	if len(num) > 0 {
		_num = num[0]
	}
	if _num <= 0 {
		_num = 1
	}
	for i := 0; i < _num; i++ {
		log.Println()
	}
}

func printOriginalAPIInfo(api *apistructs.APIInfo) {
	if api == nil {
		return
	}
	log.Printf("Original API Info:")
	defer addNewLine()
	// name
	if api.Name != "" {
		log.Printf("name: %s", api.Name)
	}
	// url
	log.Printf("url: %s", api.URL)
	// method
	log.Printf("method: %s", api.Method)
	// headers
	if len(api.Headers) > 0 {
		log.Printf("headers:")
		for _, h := range api.Headers {
			log.Printf("  key: %s", h.Key)
			log.Printf("  value: %s", h.Value)
			if h.Desc != "" {
				log.Printf("  desc: %s", h.Desc)
			}
			addLineDelimiter("  ")
		}
	}
	// params
	if len(api.Params) > 0 {
		log.Printf("params:")
		for _, p := range api.Params {
			log.Printf("  key: %s", p.Key)
			log.Printf("  value: %s", p.Value)
			if p.Desc != "" {
				log.Printf("  desc: %s", p.Desc)
			}
			addLineDelimiter("  ")
		}
	}
	// request body
	if api.Body.Type != "" {
		log.Printf("request body:")
		log.Printf("  type: %s", api.Body.Type.String())
		log.Printf("  content: %s", jsonOneLine(api.Body.Content))
	}
	// out params
	if len(api.OutParams) > 0 {
		log.Printf("out params:")
		for _, out := range api.OutParams {
			log.Printf("  arg: %s", out.Key)
			log.Printf("  source: %s", out.Source.String())
			if out.Expression != "" {
				log.Printf("  expr: %s", out.Expression)
			}
			addLineDelimiter("  ")
		}
	}
	// asserts
	if len(api.Asserts) > 0 {
		log.Printf("asserts:")
		for _, group := range api.Asserts {
			for _, assert := range group {
				log.Printf("  key: %s", assert.Arg)
				log.Printf("  operator: %s", assert.Operator)
				log.Printf("  value: %s", assert.Value)
				addLineDelimiter("  ")
			}
		}
	}
}

func printGlobalAPIConfig(cfg *apistructs.APITestEnvData) {
	if cfg == nil {
		return
	}
	log.Printf("Global API Config:")
	defer addNewLine()

	// name
	if cfg.Name != "" {
		log.Printf("name: %s", cfg.Name)
	}
	// domain
	log.Printf("domain: %s", cfg.Domain)
	// headers
	if len(cfg.Header) > 0 {
		log.Printf("headers:")
		for k, v := range cfg.Header {
			log.Printf("  key: %s", k)
			log.Printf("  value: %s", v)
			addLineDelimiter("  ")
		}
	}
	// global
	if len(cfg.Global) > 0 {
		log.Printf("global configs:")
		for key, item := range cfg.Global {
			log.Printf("  key: %s", key)
			log.Printf("  value: %s", item.Value)
			log.Printf("  type: %s", item.Type)
			if item.Desc != "" {
				log.Printf("  desc: %s", item.Desc)
			}
			addLineDelimiter("  ")
		}
	}
}

func printRenderedHTTPReq(req *apistructs.APIRequestInfo) {
	if req == nil {
		return
	}
	log.Printf("Rendered HTTP Request:")
	defer addNewLine()

	// url
	log.Printf("url: %s", req.URL)
	// method
	log.Printf("method: %s", req.Method)
	// headers
	if len(req.Headers) > 0 {
		log.Printf("headers:")
		for key, values := range req.Headers {
			log.Printf("  key: %s", key)
			if len(values) == 1 {
				log.Printf("  value: %s", values[0])
			} else {
				log.Printf("  values: %v", values)
			}
			addLineDelimiter("  ")
		}
	}
	// params
	if len(req.Params) > 0 {
		log.Printf("params:")
		for key, values := range req.Params {
			log.Printf("  key: %s", key)
			if len(values) == 1 {
				log.Printf("  value: %s", values[0])
			} else {
				log.Printf("  values: %v", values)
			}
			addLineDelimiter("  ")
		}
	}
	// body
	if req.Body.Type != "" {
		log.Printf("request body:")
		log.Printf("  type: %s", req.Body.Type.String())
		log.Printf("  content: %s", req.Body.Content)
	}
}

func printHTTPResp(resp *apistructs.APIResp) {
	if resp == nil {
		return
	}
	log.Printf("HTTP Response:")
	defer addNewLine()

	// status
	log.Printf("http status: %d", resp.Status)
	// headers
	if len(resp.Headers) > 0 {
		log.Printf("response headers:")
		for key, values := range resp.Headers {
			log.Printf("  key: %s", key)
			if len(values) == 1 {
				log.Printf("  value: %s", values[0])
			} else {
				log.Printf("  values: %v", values)
			}
			addLineDelimiter("  ")
		}
	}
	// response body
	if resp.BodyStr != "" {
		log.Printf("response body: %s", resp.BodyStr)
	}
}

func printOutParams(outParams map[string]interface{}, meta *Meta) {
	if len(outParams) == 0 {
		return
	}
	log.Printf("Out Params:")
	defer addNewLine()

	// 按定义顺序返回
	for _, define := range meta.OutParamsDefine {
		k := define.Key
		v, ok := outParams[k]
		if !ok {
			continue
		}
		meta.OutParamsResult[k] = v
		log.Printf("  arg: %s", k)
		log.Printf("  source: %s", define.Source.String())
		if define.Expression != "" {
			log.Printf("  expr: %s", define.Expression)
		}
		log.Printf("  value: %s", jsonOneLine(v))
		var vtype string
		if v == nil {
			vtype = "nil"
		} else {
			vtype = reflect.TypeOf(v).String()
		}
		log.Printf("  type: %s", vtype)
		addLineDelimiter("  ")
	}
}

func printAssertResults(success bool, results []*apistructs.APITestsAssertData) {
	log.Printf("Assert Result: %t", success)
	defer addNewLine()

	log.Printf("Assert Detail:")
	for _, result := range results {
		log.Printf("  arg: %s", result.Arg)
		log.Printf("  operator: %s", result.Operator)
		log.Printf("  value: %s", result.Value)
		log.Printf("  actualValue: %s", jsonOneLine(result.ActualValue))
		log.Printf("  success: %t", result.Success)
		if result.ErrorInfo != "" {
			log.Printf("  errorInfo: %s", result.ErrorInfo)
		}
		addLineDelimiter("  ")
	}
}
