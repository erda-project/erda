package gitmodule

var Setting = struct {
	MaxGitDiffLines          int
	MaxGitDiffLineCharacters int
	MaxGitDiffFiles          int
	MaxGitDiffSize           int
	ContextLineCount         int
	RepoStatsCache           Cache
	RepoSubmoduleCache       Cache
	PathCommitCache          Cache
	CommitCache              Cache
	ObjectSizeCache          Cache
}{
	MaxGitDiffLines:          200,
	MaxGitDiffLineCharacters: 1500, //diff line最大长度
	MaxGitDiffFiles:          100,
	MaxGitDiffSize:           5000, //大于指定大小就当二进制文件不做diff处理,单位Byte,单文件diff不做限制
	ContextLineCount:         3,
	RepoStatsCache:           NewMemCache(1000, "Repo-"),
	RepoSubmoduleCache:       NewMemCache(1000, "Repo-Submodule-"),
	PathCommitCache:          NewMemCache(5000, "path_commit_"),
	CommitCache:              NewMemCache(5000, "commit_"),
	ObjectSizeCache:          NewMemCache(10000, "object_size_"),
}
