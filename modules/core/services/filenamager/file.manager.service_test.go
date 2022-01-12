package filemanager

import (
	"reflect"
	"testing"

	"github.com/erda-project/erda-proto-go/core/services/filemanager/pb"
)

// import (
// 	context "context"
// 	servicehub "github.com/erda-project/erda-infra/base/servicehub"
// 	pb "github.com/erda-project/erda-proto-go/core/services/filemanager/pb"
// 	reflect "reflect"
// 	testing "testing"
// )

// func Test_fileManagerService_ListFiles(t *testing.T) {
// 	type args struct {
// 		ctx context.Context
// 		req *pb.ListFilesRequest
// 	}
// 	tests := []struct {
// 		name     string
// 		service  string
// 		config   string
// 		args     args
// 		wantResp *pb.ListFilesResponse
// 		wantErr  bool
// 	}{
// 		// TODO: Add test cases.
// 		{
// 			"case 1",
// 			"erda.core.services.filemanager.FileManagerService",
// 			`
// erda.core.services.filemanager:
// `,
// 			args{
// 				context.TODO(),
// 				&pb.ListFilesRequest{
// 					// TODO: setup fields
// 				},
// 			},
// 			&pb.ListFilesResponse{
// 				// TODO: setup fields.
// 			},
// 			false,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			hub := servicehub.New()
// 			events := hub.Events()
// 			go func() {
// 				hub.RunWithOptions(&servicehub.RunOptions{Content: tt.config})
// 			}()
// 			err := <-events.Started()
// 			if err != nil {
// 				t.Error(err)
// 				return
// 			}
// 			srv := hub.Service(tt.service).(pb.FileManagerServiceServer)
// 			got, err := srv.ListFiles(tt.args.ctx, tt.args.req)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("fileManagerService.ListFiles() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if !reflect.DeepEqual(got, tt.wantResp) {
// 				t.Errorf("fileManagerService.ListFiles() = %v, want %v", got, tt.wantResp)
// 			}
// 		})
// 	}
// }

// func Test_fileManagerService_ReadFile(t *testing.T) {
// 	type args struct {
// 		ctx context.Context
// 		req *pb.ReadFileRequest
// 	}
// 	tests := []struct {
// 		name     string
// 		service  string
// 		config   string
// 		args     args
// 		wantResp *pb.ReadFileResponse
// 		wantErr  bool
// 	}{
// 		// TODO: Add test cases.
// 		{
// 			"case 1",
// 			"erda.core.services.filemanager.FileManagerService",
// 			`
// erda.core.services.filemanager:
// `,
// 			args{
// 				context.TODO(),
// 				&pb.ReadFileRequest{
// 					// TODO: setup fields
// 				},
// 			},
// 			&pb.ReadFileResponse{
// 				// TODO: setup fields.
// 			},
// 			false,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			hub := servicehub.New()
// 			events := hub.Events()
// 			go func() {
// 				hub.RunWithOptions(&servicehub.RunOptions{Content: tt.config})
// 			}()
// 			err := <-events.Started()
// 			if err != nil {
// 				t.Error(err)
// 				return
// 			}
// 			srv := hub.Service(tt.service).(pb.FileManagerServiceServer)
// 			got, err := srv.ReadFile(tt.args.ctx, tt.args.req)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("fileManagerService.ReadFile() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if !reflect.DeepEqual(got, tt.wantResp) {
// 				t.Errorf("fileManagerService.ReadFile() = %v, want %v", got, tt.wantResp)
// 			}
// 		})
// 	}
// }

// func Test_fileManagerService_WriteFile(t *testing.T) {
// 	type args struct {
// 		ctx context.Context
// 		req *pb.WriteFileRequest
// 	}
// 	tests := []struct {
// 		name     string
// 		service  string
// 		config   string
// 		args     args
// 		wantResp *pb.WriteFileResponse
// 		wantErr  bool
// 	}{
// 		// TODO: Add test cases.
// 		{
// 			"case 1",
// 			"erda.core.services.filemanager.FileManagerService",
// 			`
// erda.core.services.filemanager:
// `,
// 			args{
// 				context.TODO(),
// 				&pb.WriteFileRequest{
// 					// TODO: setup fields
// 				},
// 			},
// 			&pb.WriteFileResponse{
// 				// TODO: setup fields.
// 			},
// 			false,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			hub := servicehub.New()
// 			events := hub.Events()
// 			go func() {
// 				hub.RunWithOptions(&servicehub.RunOptions{Content: tt.config})
// 			}()
// 			err := <-events.Started()
// 			if err != nil {
// 				t.Error(err)
// 				return
// 			}
// 			srv := hub.Service(tt.service).(pb.FileManagerServiceServer)
// 			got, err := srv.WriteFile(tt.args.ctx, tt.args.req)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("fileManagerService.WriteFile() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if !reflect.DeepEqual(got, tt.wantResp) {
// 				t.Errorf("fileManagerService.WriteFile() = %v, want %v", got, tt.wantResp)
// 			}
// 		})
// 	}
// }

// func Test_fileManagerService_MakeDirectory(t *testing.T) {
// 	type args struct {
// 		ctx context.Context
// 		req *pb.MakeDirectoryRequest
// 	}
// 	tests := []struct {
// 		name     string
// 		service  string
// 		config   string
// 		args     args
// 		wantResp *pb.MakeDirectoryResponse
// 		wantErr  bool
// 	}{
// 		// TODO: Add test cases.
// 		{
// 			"case 1",
// 			"erda.core.services.filemanager.FileManagerService",
// 			`
// erda.core.services.filemanager:
// `,
// 			args{
// 				context.TODO(),
// 				&pb.MakeDirectoryRequest{
// 					// TODO: setup fields
// 				},
// 			},
// 			&pb.MakeDirectoryResponse{
// 				// TODO: setup fields.
// 			},
// 			false,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			hub := servicehub.New()
// 			events := hub.Events()
// 			go func() {
// 				hub.RunWithOptions(&servicehub.RunOptions{Content: tt.config})
// 			}()
// 			err := <-events.Started()
// 			if err != nil {
// 				t.Error(err)
// 				return
// 			}
// 			srv := hub.Service(tt.service).(pb.FileManagerServiceServer)
// 			got, err := srv.MakeDirectory(tt.args.ctx, tt.args.req)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("fileManagerService.MakeDirectory() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if !reflect.DeepEqual(got, tt.wantResp) {
// 				t.Errorf("fileManagerService.MakeDirectory() = %v, want %v", got, tt.wantResp)
// 			}
// 		})
// 	}
// }

// func Test_fileManagerService_MoveFile(t *testing.T) {
// 	type args struct {
// 		ctx context.Context
// 		req *pb.MoveFileRequest
// 	}
// 	tests := []struct {
// 		name     string
// 		service  string
// 		config   string
// 		args     args
// 		wantResp *pb.MoveFileResponse
// 		wantErr  bool
// 	}{
// 		// TODO: Add test cases.
// 		{
// 			"case 1",
// 			"erda.core.services.filemanager.FileManagerService",
// 			`
// erda.core.services.filemanager:
// `,
// 			args{
// 				context.TODO(),
// 				&pb.MoveFileRequest{
// 					// TODO: setup fields
// 				},
// 			},
// 			&pb.MoveFileResponse{
// 				// TODO: setup fields.
// 			},
// 			false,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			hub := servicehub.New()
// 			events := hub.Events()
// 			go func() {
// 				hub.RunWithOptions(&servicehub.RunOptions{Content: tt.config})
// 			}()
// 			err := <-events.Started()
// 			if err != nil {
// 				t.Error(err)
// 				return
// 			}
// 			srv := hub.Service(tt.service).(pb.FileManagerServiceServer)
// 			got, err := srv.MoveFile(tt.args.ctx, tt.args.req)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("fileManagerService.MoveFile() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if !reflect.DeepEqual(got, tt.wantResp) {
// 				t.Errorf("fileManagerService.MoveFile() = %v, want %v", got, tt.wantResp)
// 			}
// 		})
// 	}
// }

func Test_parseFileList(t *testing.T) {
	tests := []struct {
		name    string
		text    string
		want    *pb.FileDirectory
		wantErr bool
	}{
		{
			text: `/root
total 208
drwxr-xr-x   1 root root   4096 2021-12-14T16:55:23.218360893 .
drwxr-xr-x   1 root root   4096 2021-12-14T16:55:23.218360893 ..
-rw-r--r--   1 root root  12123 2019-10-01T09:16:32.000000000 anaconda-post.log
drwxr-xr-x   3 root root   4096 2021-09-23T14:46:27.000000000 app
-rw-r--r--   1 root root 137382 2021-09-23T14:45:23.000000000 arthas-boot.jar
lrwxrwxrwx   1 root root      7 2019-10-01T09:15:19.000000000 bin -> usr/bin`,
			want: &pb.FileDirectory{
				Directory: "/root",
				Files: []*pb.FileInfo{
					{
						Name:      ".",
						Mode:      "drwxr-xr-x",
						Size:      4096,
						HardLinks: 1,
						ModTime:   1639472123218,
						User:      "root",
						UserGroup: "root",
					},
					{
						Name:      "..",
						Mode:      "drwxr-xr-x",
						Size:      4096,
						HardLinks: 1,
						ModTime:   1639472123218,
						User:      "root",
						UserGroup: "root",
					},
					{
						Name:      "anaconda-post.log",
						Mode:      "-rw-r--r--",
						Size:      12123,
						HardLinks: 1,
						ModTime:   1569892592000,
						User:      "root",
						UserGroup: "root",
					},
					{
						Name:      "app",
						Mode:      "drwxr-xr-x",
						Size:      4096,
						HardLinks: 3,
						ModTime:   1632379587000,
						User:      "root",
						UserGroup: "root",
					},
					{
						Name:      "arthas-boot.jar",
						Mode:      "-rw-r--r--",
						Size:      137382,
						HardLinks: 1,
						ModTime:   1632379523000,
						User:      "root",
						UserGroup: "root",
					},
					{
						Name:      "bin",
						Mode:      "lrwxrwxrwx",
						Size:      7,
						HardLinks: 1,
						ModTime:   1569892519000,
						User:      "root",
						UserGroup: "root",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseFileList(tt.text)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseFileList() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseFileList() = %v, want %v", got, tt.want)
			}
		})
	}
}
