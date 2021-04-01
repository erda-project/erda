package dao

var dbClient *DBClient

// func init() {
// 	os.Setenv("MYSQL_HOST", "localhost")
// 	os.Setenv("MYSQL_PORT", "3306")
// 	os.Setenv("MYSQL_USERNAME", "")
// 	os.Setenv("MYSQL_PASSWORD", "")
// 	os.Setenv("MYSQL_DATABASE", "")
// 	url := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=%s",
// 		os.Getenv("MYSQL_USERNAME"), os.Getenv("MYSQL_PASSWORD"), os.Getenv("MYSQL_HOST"),
// 		os.Getenv("MYSQL_PORT"), os.Getenv("MYSQL_DATABASE"), "Local")
//
// 	logrus.Debugf("Initialize db with %s, url: %s", DIALECT, url)
//
// 	db, err := gorm.Open(DIALECT, url)
// 	if err != nil {
// 		logrus.Fatal(err)
// 	}
// 	dbClient = &DBClient{
// 		db,
// 	}
// }

// func TestDBClient_ListNexusRepositories(t *testing.T) {
// 	repos, err := dbClient.ListNexusRepositories(apistructs.NexusRepositoryListRequest{
// 		Formats: []nexus.RepositoryFormat{nexus.RepositoryFormatNpm},
// 	})
// 	assert.NoError(t, err)
// 	for _, repo := range repos {
// 		b, _ := json.MarshalIndent(repo, "", "  ")
// 		fmt.Println(string(b))
// 	}
//
// 	fmt.Println("=============")
//
// 	repos, err = dbClient.ListNexusRepositories(apistructs.NexusRepositoryListRequest{
// 		NameContains: []string{"-proxy-org", "maven2-proxy-org-", "maven2-hosted-"},
// 	})
// 	assert.NoError(t, err)
// 	for _, repo := range repos {
// 		b, _ := json.MarshalIndent(repo, "", "  ")
// 		fmt.Println(string(b))
// 	}
// }
