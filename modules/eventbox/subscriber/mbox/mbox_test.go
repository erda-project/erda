package sms

// func TestMBoxSend(t *testing.T) {

// 	subscriber := New(bundle.New(bundle.WithCMDB()))
// 	params := map[string]string{
// 		"notice": "notice",
// 		"link":   "link",
// 		"title":  "机器告警通知",
// 	}
// 	request := map[string]interface{}{
// 		"template": "测试测试",
// 		"params":   params,
// 		"orgID":    1,
// 		"module":   "monitor",
// 	}

// 	ids := []string{"1", "2"}
// 	dest, _ := json.Marshal(ids)
// 	msg := &types.Message{
// 		Content: request,
// 		Time:    time.Now().Unix(),
// 	}
// 	content, _ := json.Marshal(request)
// 	errs := subscriber.Publish(string(dest), string(content), msg.Time, msg)
// 	if len(errs) > 0 {
// 		panic(errs[0])
// 	}
// }
