# proto

## Libraries

Current supported languages are:

* Go , see [erda-proto-go](https://github.com/erda-project/erda-proto-go)

## Scripts

### proto_fetcher.sh

`proto_fetcher.sh` is a shell script for fetching and processing external protocol buffer files into the desired format.

General Function:

| Method                          | Description                                                                                                                                                                                   |
|---------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `fetch_code()`                  | Fetch code using git. <br/> Parameters: <ul><li>`--repo`: repository URL.</li><li>`--mirror`: repository mirror.</li><li>`--commit-id`: commit ID.</li><li>`--branch`: branch name.</li></ul> |
| `cleanup_external_repo()`       | Clean up external repo directory.                                                                                                                                                             |
| `move_proto_files`              | Moves all .proto files from a source dir to a target dir                                                                                                                                      | 
| `add_skip_go_from_annotation()` | Add `// +SKIP_GO-FORM` annotation to proto files. <br/> Parameters: <ul><li>`target_path`: the directory where proto files are located.</li></ul>                                             |
