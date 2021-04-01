// +build !default

package dao

// func newDBClient2() (*DBClient, error) {
// 	os.Setenv("MYSQL_HOST", "192.168.1.104")
// 	os.Setenv("MYSQL_PORT", "3306")
// 	os.Setenv("MYSQL_USERNAME", "root")
// 	os.Setenv("MYSQL_PASSWORD", "test123")
// 	os.Setenv("MYSQL_DATABASE", "dice")
// 	os.Setenv("COLONY_OFFICER_ADDR", "dice.officer.marathon.l4lb.thisdcos.directory:9029")
// 	os.Setenv("DICE_HUB_ADDR", "dicehub.marathon.l4lb.thisdcos.directory:10000")
// 	os.Setenv("ADDON_PLATFORM_ADDR", "addon-platform.marathon.l4lb.thisdcos.directory:8080")
// 	os.Setenv("KAFKA_BROKERS", "kafka1.marathon.l4lb.thisdcos.directory:9092")
// 	os.Setenv("CMDB_HOST_TOPIC", "cmdb_metaserver_host")
// 	os.Setenv("CMDB_CONTAINER_TOPIC", "cmdb_metaserver_container")
// 	os.Setenv("CMDB_GROUP", "cmdb_group")
// 	os.Setenv("EVENTBOX_ADDR", "192.168.0.1:8080")
// 	os.Setenv("SELF_ADDR", "127.0.0.1:80")
//
// 	defer func() {
// 		os.Unsetenv("MYSQL_HOST")
// 		os.Unsetenv("MYSQL_PORT")
// 		os.Unsetenv("MYSQL_USERNAME")
// 		os.Unsetenv("MYSQL_PASSWORD")
// 		os.Unsetenv("MYSQL_DATABASE")
// 		os.Unsetenv("COLONY_OFFICER_ADDR")
// 		os.Unsetenv("DICE_HUB_ADDR")
// 		os.Unsetenv("ADDON_PLATFORM_ADDR")
// 		os.Unsetenv("EVENTBOX_ADDR")
// 		os.Unsetenv("SELF_ADDR")
// 	}()
//
// 	//if err := conf.Parse(); err != nil {
// 	//	return nil, err
// 	//}
// 	//
// 	//return NewDBClient(conf.DbDial)
//
// 	return &DBClient{}, nil
// }
//
// func TestInsertContainer(t *testing.T) {
// 	db, err := newDBClient2()
// 	assert.Nil(t, err)
//
// 	c := &types.CmContainer{
// 		ID:     "test-123",
// 		Status: "Stopped",
// 	}
// 	err = db.InsertContainer(context.Background(), c)
// }
//
// func TestAllContainersOnlyByService(t *testing.T) {
// 	db, err := newDBClient2()
// 	assert.Nil(t, err)
//
// 	runtime := "1214"
// 	services := []string{
// 		"blog-service",
// 		"showcase-front",
// 	}
//
// 	summary, err := db.AllContainersByService(context.Background(), runtime, services)
// 	assert.Nil(t, err)
//
// 	t.Logf("TestAllContainersOnlyByService result: %+v", summary)
// }
//
// func TestDeleteStoppedContainersByPeriod(t *testing.T) {
// 	db, err := newDBClient2()
// 	assert.Nil(t, err)
//
// 	period := 7 * 24 * time.Hour
//
// 	err = db.DeleteStoppedContainersByPeriod(context.Background(), period)
// 	assert.Nil(t, err)
// }
//
// func TestQueryContainer(t *testing.T) {
// 	db, err := newDBClient2()
// 	assert.Nil(t, err)
//
// 	_, err = db.QueryContainer(context.Background(), "xxx", "123456789012")
// 	assert.EqualError(t, err, types.NotFound)
// }
//
// func TestUpdateContainerByPrimaryKeyID(t *testing.T) {
// 	db, err := newDBClient2()
// 	assert.Nil(t, err)
//
// 	c := &types.CmContainer{
// 		ID: "lbl-test2",
// 	}
//
// 	c.ModelHeader.ID = 11
//
// 	err = db.UpdateContainerByPrimaryKeyID(context.Background(), c)
// 	assert.Nil(t, err)
// }
