package webhook

// func TestCreateBuiltinHooks(t *testing.T) {
// 	impl, err := NewWebHookImpl()
// 	assert.Nil(t, err)
// 	assert.Nil(t, createIfNotExist(impl, &CreateHookRequest{
// 		Name:   "test-builtin-hook",
// 		Events: []string{"test-event"},
// 		URL:    "http://xxx/a",
// 		Active: true,
// 		HookLocation: HookLocation{
// 			Org:         "-1",
// 			Project:     "-1",
// 			Application: "-1",
// 		},
// 	}))
// }

// func TestMakeSureBuiltinHooks(t *testing.T) {
// 	impl, err := NewWebHookImpl()
// 	assert.Nil(t, err)
// 	assert.Nil(t, MakeSureBuiltinHooks(impl))
// }
