package yakgrpc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"github.com/yaklang/yaklang/common/consts"
	"github.com/yaklang/yaklang/common/go-funk"
	"github.com/yaklang/yaklang/common/log"
	"github.com/yaklang/yaklang/common/utils"
	"github.com/yaklang/yaklang/common/yak/yaklib/codec"
	"github.com/yaklang/yaklang/common/yakgrpc/yakit"
	"github.com/yaklang/yaklang/common/yakgrpc/ypb"
	"google.golang.org/protobuf/encoding/protowire"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var (
	currentProjectMutex = new(sync.Mutex)
)

func (s *Server) SetCurrentProject(ctx context.Context, req *ypb.SetCurrentProjectRequest) (*ypb.Empty, error) {
	currentProjectMutex.Lock()
	defer currentProjectMutex.Unlock()
	if req.GetId() > 0 {
		db := s.GetProfileDatabase()
		proj, err := yakit.GetProjectById(db, req.GetId(), yakit.TypeProject)
		if err != nil {
			err := yakit.InitializingProjectDatabase()
			if err != nil {
				log.Errorf("init db failed: %s", err)
			}
			return &ypb.Empty{}, nil
		}
		err = yakit.SetCurrentProjectById(db, req.GetId())
		if err != nil {
			err := yakit.InitializingProjectDatabase()
			if err != nil {
				log.Errorf("init db failed: %s", err)
			}
			return &ypb.Empty{}, nil
		}
		// is not the default database and does not need to generate files
		if CheckDefault(proj.ProjectName, proj.Type, proj.FolderID, proj.ChildFolderID) == nil {
			old, err := os.Open(proj.DatabasePath)
			if err != nil {
				return nil, utils.Errorf("Local database file not found: %s", err)
			}
			old.Close()
		}

		projectDatabase, err := gorm.Open("sqlite3", proj.DatabasePath)
		if err != nil {
			return nil, utils.Errorf("open project database failed: %s", err)
		}
		projectDatabase.AutoMigrate(yakit.ProjectTables...)
		consts.SetDefaultYakitProjectDatabaseName(proj.DatabasePath)
		consts.SetGormProjectDatabase(projectDatabase)
		return &ypb.Empty{}, nil
	}
	return nil, utils.Errorf("params is empty")
}

func (s *Server) GetProjects(ctx context.Context, req *ypb.GetProjectsRequest) (*ypb.GetProjectsResponse, error) {
	paging, data, err := yakit.QueryProject(s.GetProfileDatabase(), req)
	if err != nil {
		return nil, err
	}
	total, _ := yakit.QueryProjectTotal(s.GetProfileDatabase(), req)
	return &ypb.GetProjectsResponse{
		Projects: funk.Map(data, func(i *yakit.Project) *ypb.ProjectDescription {
			return i.ToGRPCModel()
		}).([]*ypb.ProjectDescription),
		Pagination:   req.GetPagination(),
		Total:        int64(paging.TotalRecord),
		TotalPage:    int64(paging.Page),
		ProjectToTal: int64(total.TotalRecord),
	}, nil
}

var projectNameRe = regexp.MustCompile(`(?i)[_a-z0-9\p{Han}][-_0-9a-z \p{Han}]*`)

func projectNameToFileName(s string) string {
	s = strings.ReplaceAll(s, "-", "_")
	return strings.Join(projectNameRe.FindAllString(s, -1), "_")
}

var encryptProjectMagic = []byte{0xff, 0xff, 0xff, 0xff}

func (s *Server) NewProject(ctx context.Context, req *ypb.NewProjectRequest) (*ypb.NewProjectResponse, error) {
	if req.Type == "" {
		return nil, utils.Errorf("type is empty")
	}
	name := req.GetProjectName()
	if !projectNameRe.MatchString(name) {
		return nil, utils.Errorf("create project by name failed! name should match %v", projectNameRe.String())
	}
	var pathName string
	// The project needs to be saved, the folder is not required
	if req.Type == yakit.TypeProject {
		databaseName := fmt.Sprintf("yakit-project-%v-%v.sqlite3.db", projectNameToFileName(name), time.Now().Unix())
		pathName = filepath.Join(consts.GetDefaultYakitProjectsDir(), databaseName)
		if ok, _ := utils.PathExists(pathName); ok {
			return nil, utils.Errorf("path: %v is not existed", pathName)
		}
	}
	projectData := &yakit.Project{
		ProjectName:   req.GetProjectName(),
		Description:   req.GetDescription(),
		DatabasePath:  pathName,
		Type:          req.Type,
		FolderID:      req.FolderId,
		ChildFolderID: req.ChildFolderId,
	}
	pro, _ := yakit.GetProjectByWhere(s.GetProfileDatabase(), req.GetProjectName(), req.GetFolderId(), req.GetChildFolderId(), req.GetType(), req.GetId())
	if pro != nil {
		return nil, utils.Errorf("Same Project or file names cannot be repeated in the same level directory")
	}

	if req.GetId() > 0 {
		oldPro, err := yakit.GetProjectByID(s.GetProfileDatabase(), req.GetId())
		if err != nil {
			return nil, utils.Errorf("update project not exist %v", err.Error())
		}

		err = os.Rename(oldPro.DatabasePath, pathName)
		if err != nil {
			return nil, errors.Errorf(" oldfile=%v rename newname=%v fail=%v", oldPro.DatabasePath, pathName, err)
		}
		err = yakit.UpdateProject(s.GetProfileDatabase(), req.GetId(), *projectData)
		if err != nil {
			return nil, utils.Errorf("update project failed!")
		}

		return &ypb.NewProjectResponse{Id: req.GetId(), ProjectName: req.GetProjectName()}, nil

	} else {
		if CheckDefault(req.GetProjectName(), req.GetType(), req.GetFolderId(), req.GetChildFolderId()) != nil {
			return nil, utils.Errorf("cannot use this name: %s, %v is for buildin", yakit.INIT_DATABASE_RECORD_NAME, yakit.INIT_DATABASE_RECORD_NAME)
		}
	}

	/*pro, _ := yakit.GetProjectByWhere(s.GetProfileDatabase(), req.GetProjectName(), req.FolderId, req.ChildFolderId, req.Type, 0)
	if pro != nil {
		return nil, utils.Errorf("File/Folder name cannot be repeated")
	}*/
	db := s.GetProfileDatabase()
	if db = db.Create(&projectData); db.Error != nil {
		return nil, db.Error
	}
	// created library
	projectDatabase, err := gorm.Open("sqlite3", pathName)
	if err != nil {
		return nil, utils.Errorf("open project database failed: %s", err)
	}
	projectDatabase.AutoMigrate(yakit.ProjectTables...)
	projectDatabase.Close()

	return &ypb.NewProjectResponse{Id: int64(projectData.ID), ProjectName: req.GetProjectName()}, nil
}

/*func (s *Server) RemoveProject(ctx context.Context, req *ypb.RemoveProjectRequest) (*ypb.Empty, error) {
	if req.GetProjectName() == yakit.INIT_DATABASE_RECORD_NAME {
		return nil, utils.Error("[default] cannot be deleted")
	}

	err := yakit.DeleteProjectByProjectName(s.GetProfileDatabase(), req.GetProjectName())
	if err != nil {
		return nil, err
	}
	return &ypb.Empty{}, nil
}*/

func (s *Server) IsProjectNameValid(ctx context.Context, req *ypb.IsProjectNameValidRequest) (*ypb.Empty, error) {
	if req.GetType() == "" {
		return nil, utils.Error("type is empty")
	}
	if CheckDefault(req.GetProjectName(), req.GetType(), req.GetFolderId(), req.GetChildFolderId()) != nil {
		return nil, utils.Error("[default] cannot be user's db name")
	}
	proj, _ := yakit.GetProject(consts.GetGormProfileDatabase(), req)
	if proj != nil {
		return nil, utils.Errorf("project name: %s is existed", req.GetProjectName())
	}

	if !projectNameRe.MatchString(req.GetProjectName()) {
		return nil, utils.Errorf("validate project by name failed! name should match %v", projectNameRe.String())
	}

	return &ypb.Empty{}, nil
}

func (s *Server) GetCurrentProject(ctx context.Context, _ *ypb.Empty) (*ypb.ProjectDescription, error) {
	currentProjectMutex.Lock()
	defer currentProjectMutex.Unlock()

	db := s.GetProfileDatabase()
	proj, err := yakit.GetCurrentProject(db)
	if err != nil {
		return nil, utils.Errorf("cannot fetch current project")
	}
	return proj.ToGRPCModel(), nil
}

func (s *Server) ExportProject(req *ypb.ExportProjectRequest, stream ypb.Yak_ExportProjectServer) error {
	var outputFile string
	feedProgress := func(verbose string, progress float64) {
		stream.Send(&ypb.ProjectIOProgress{
			TargetPath: outputFile,
			Percent:    progress,
			Verbose:    verbose,
		})
	}
	feedProgress("Start exporting", 0.1)

	/*path := consts.GetDefaultYakitProjectDatabase(consts.GetDefaultYakitBaseDir())
	if !utils.IsFile(path) {
		feedProgress("Export failed -"+"Database does not exist:"+path, 0.9)
		return utils.Errorf("cannot found database file in: %s", path)
	}*/
	proj, err := yakit.GetProjectById(s.GetProfileDatabase(), req.GetId(), yakit.TypeProject)
	if err != nil {
		feedProgress("Export failed -"+"Database does not exist:", 0.9)
		return utils.Errorf("cannot found database file in: %s", err.Error())
	}
	feedProgress("Find data files", 0.3)
	fp, err := os.Open(proj.DatabasePath)
	if err != nil {
		feedProgress("cannot be found Database file"+err.Error(), 0.4)
		return utils.Errorf("open database failed: %s", err)
	}
	defer fp.Close()

	/*db := s.GetProfileDatabase()
	proj, err := yakit.GetCurrentProject(db)
	if err != nil {
		feedProgress("cannot find the current database:"+err.Error(), 0.5)
		return err
	}*/

	suffix := ""
	if req.GetPassword() != "" {
		suffix = ".enc"
	}
	outputFile = filepath.Join(consts.GetDefaultYakitProjectsDir(), "project-"+projectNameToFileName(proj.ToGRPCModel().GetProjectName())+".yakitproject"+suffix)
	outFp, err := os.OpenFile(outputFile, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		feedProgress("failed to open output file!", 0.5)
		return err
	}
	defer outFp.Close()

	feedProgress("Start exporting basic project data", 0.6)

	var ret []byte
	ret = protowire.AppendString(ret, proj.ProjectName)
	ret = protowire.AppendString(ret, proj.Description)
	params := map[string]interface{}{
		"allowPassword": req.GetPassword() != "",
	}
	raw, _ := json.Marshal(params)
	ret = protowire.AppendBytes(ret, raw)
	feedProgress("Successfully exported project basic data, started to export project database", 0.65)

	ctx, cancel := context.WithCancel(context.Background())
	var finished = false
	go func() {
		defer func() {
			finished = true
		}()
		var percent float64 = 0.65
		var count = 0
		for {
			count++
			select {
			case <-ctx.Done():
				return
			default:
				nowPercent := percent + float64(count)*0.01
				if nowPercent > 0.93 {
					return
				}
				feedProgress("", nowPercent)
				time.Sleep(time.Second)
			}
		}
	}()
	var buf bytes.Buffer
	buf.Write(ret)
	io.Copy(&buf, fp)

	var results []byte = buf.Bytes()
	if req.GetPassword() != "" {
		feedProgress("Start encrypting the database... SM4-GCM", 0)
		encData, err := codec.SM4GCMEnc(codec.PKCS7Padding([]byte(req.GetPassword())), results, nil)
		if err != nil {
			feedProgress("Failed to encrypt the database:"+err.Error(), 0.97)
			cancel()
			return err
		}
		results = encData
	}

	feedProgress("Start compressing the database", 0)
	results, err = utils.GzipCompress(results)
	if err != nil {
		feedProgress("Export project failed: GZIP Compression failed: "+err.Error(), 0.97)
		cancel()
		return err
	}

	if req.GetPassword() != "" {
		feedProgress("Start writing encrypted data, please keep the password properly", 0.94)
	}

	if req.GetPassword() != "" {
		outFp.Write(encryptProjectMagic)
	}
	outFp.Write(results)
	cancel()
	for !finished {
		time.Sleep(300 * time.Millisecond)
	}
	feedProgress("exported successfully, exported project size:"+utils.ByteSize(uint64(len(results))), 1.0)
	return nil
}

func (s *Server) MigrateLegacyDatabase(ctx context.Context, req *ypb.Empty) (*ypb.Empty, error) {
	err := yakit.MigrateLegacyDatabase()
	if err != nil {
		return nil, err
	}
	return &ypb.Empty{}, nil
}

func (s *Server) ImportProject(req *ypb.ImportProjectRequest, stream ypb.Yak_ImportProjectServer) error {
	feedProgress := func(verbose string, progress float64) {
		stream.Send(&ypb.ProjectIOProgress{
			TargetPath: req.GetProjectFilePath(),
			Percent:    progress,
			Verbose:    verbose,
		})
	}

	feedProgress("Start importing the project: "+req.GetLocalProjectName(), 0.1)
	path := req.GetProjectFilePath()
	if !utils.IsFile(path) {
		return utils.Errorf("cannot find local project path: %s", path)
	}

	feedProgress("Open the local file of the project:"+req.GetProjectFilePath(), 0.2)
	fp, err := os.Open(req.GetProjectFilePath())
	if err != nil {
		feedProgress("failed to open the local file of the project:"+err.Error(), 0.9)
		return err
	}
	defer fp.Close()

	feedProgress("in the same level directory Reading the project file", 0.3)
	raw, err := ioutil.ReadAll(fp)
	if err != nil {
		feedProgress("failed to read the project file:"+err.Error(), 0.9)
		return err
	}

	if bytes.HasPrefix(raw, encryptProjectMagic) {
		if req.GetPassword() != "" {
			raw = raw[len(encryptProjectMagic):]
		} else {
			feedProgress("requires a password Decryption of project data", 0.99)
			return utils.Error("requires password to decrypt")
		}
	}

	feedProgress("Decompressing the database", 0.4)
	bytes, err := utils.GzipDeCompress(raw)
	if err != nil {
		return err
	}

	feedProgress("Decompression completed, decryption of database", 0.43)
	if req.GetPassword() != "" {
		decData, err := codec.SM4GCMDec(codec.PKCS7Padding([]byte(req.GetPassword())), bytes, nil)
		if err != nil {
			feedProgress("Decryption failed!", 0.99)
			return utils.Error("Decryption failed!")
		}
		bytes = decData
	}

	feedProgress("Read basic project information", 0.45)
	projectName, n := protowire.ConsumeString(bytes)
	bytes = bytes[n:]
	description, n := protowire.ConsumeString(bytes)
	bytes = bytes[n:]
	paramsBytes, n := protowire.ConsumeBytes(bytes)
	bytes = bytes[n:]

	var params = make(map[string]interface{})
	json.Unmarshal(paramsBytes, &params)
	if params != nil && len(params) > 0 {
		// handle params
	}

	feedProgress(fmt.Sprintf(
		"read basic project information, original project name %v, description information: %v",
		projectName, description,
	), 0.5)

	if req.GetLocalProjectName() != "" {
		projectName = req.GetLocalProjectName()
	}

	if projectName == "[default]" {
		projectName = "_default_"
	}

	_, err = s.IsProjectNameValid(stream.Context(), &ypb.IsProjectNameValidRequest{ProjectName: projectName, Type: yakit.TypeProject})
	if err != nil {
		projectName = projectName + fmt.Sprintf("_%v", utils.RandStringBytes(6))
		_, err := s.IsProjectNameValid(stream.Context(), &ypb.IsProjectNameValidRequest{ProjectName: projectName})
		if err != nil {
			feedProgress("failed to create new project:"+projectName+"："+err.Error(), 0.9)
			return utils.Errorf("cannot valid project name: %s", err)
		}
	}
	feedProgress("Create a new project:"+projectName, 0.6)
	databaseName := fmt.Sprintf("yakit-%v-%v.sqlite3.db", projectNameToFileName(projectName), time.Now().Unix())
	fileName := filepath.Join(consts.GetDefaultYakitProjectsDir(), databaseName)
	err = os.WriteFile(
		fileName,
		bytes,
		0666,
	)
	if err != nil {
		feedProgress("Failed to create a new database:"+err.Error(), 0.9)
		return err
	}

	feedProgress("Create the project:"+projectName, 0.7)
	proj := &yakit.Project{
		ProjectName:   projectName,
		Description:   description,
		DatabasePath:  fileName,
		FolderID:      req.FolderId,
		ChildFolderID: req.GetChildFolderId(),
		Type:          "project",
	}
	err = yakit.CreateOrUpdateProject(s.GetProfileDatabase(), projectName, req.FolderId, req.ChildFolderId, "project", proj)
	if err != nil {
		feedProgress("failed to create project data:"+err.Error(), 0.9)
		return err
	}
	feedProgress("imported the project successfully.", 1.0)
	return nil
}

func CheckDefault(ProjectName, Type string, FolderId, ChildFolderId int64) error {
	if ProjectName == yakit.INIT_DATABASE_RECORD_NAME && Type == yakit.TypeProject && FolderId == 0 && ChildFolderId == 0 {
		return utils.Error("[default] cannot be deleted")
	}
	return nil
}

func (s *Server) DeleteProject(ctx context.Context, req *ypb.DeleteProjectRequest) (*ypb.Empty, error) {
	if req.GetId() > 0 {
		db := s.GetProfileDatabase()
		db = db.Where(" id = ? or folder_id = ? or child_folder_id = ? ", req.GetId(), req.GetId(), req.GetId())
		projects := yakit.YieldProject(db, ctx)
		if projects == nil {
			return nil, utils.Error("Delete project does not exist")
		}
		proj, err := yakit.GetDefaultProject(s.GetProfileDatabase())
		if err != nil {
			return nil, utils.Errorf("open project database failed: %s", err)
		}
		err = yakit.SetCurrentProjectById(s.GetProfileDatabase(), int64(proj.ID))
		if err != nil {
			return nil, utils.Errorf("open project database failed: %s", err)
		}

		for k := range projects {
			if CheckDefault(k.ProjectName, k.Type, k.FolderID, k.ChildFolderID) != nil {
				log.Info("[default] cannot be deleted")
				break
			}
			if req.IsDeleteLocal {
				consts.GetGormProjectDatabase().Close()
				err := os.RemoveAll(k.DatabasePath)
				if err != nil {
					log.Error("failed to delete the local database:" + err.Error())
				}
			}
			defaultDb, err := gorm.Open("sqlite3", proj.DatabasePath)
			if err != nil {
				log.Errorf("failed to switch default database %s", err)
			}
			defaultDb.AutoMigrate(yakit.ProjectTables...)
			consts.SetDefaultYakitProjectDatabaseName(proj.DatabasePath)
			consts.SetGormProjectDatabase(defaultDb)

			err = yakit.DeleteProjectById(s.GetProfileDatabase(), int64(k.ID))
			if err != nil {
				log.Error("Failed to delete the project:" + err.Error())
			}
		}
		return &ypb.Empty{}, nil
	}
	return &ypb.Empty{}, nil
}

func (s *Server) GetDefaultProject(ctx context.Context, req *ypb.Empty) (*ypb.ProjectDescription, error) {
	proj, err := yakit.GetDefaultProject(s.GetProfileDatabase())
	if err != nil {
		return nil, utils.Errorf("cannot fetch default project")
	}
	return proj.ToGRPCModel(), nil
}

func (s *Server) QueryProjectDetail(ctx context.Context, req *ypb.QueryProjectDetailRequest) (*ypb.ProjectDescription, error) {
	var proj *ypb.ProjectDescription
	if req.GetId() > 0 {
		proj, err := yakit.GetProjectDetail(s.GetProfileDatabase(), req.GetId())
		if err != nil {
			return nil, utils.Errorf("cannot fetch project")
		}
		return proj.BackGRPCModel(), nil
	}
	return proj, nil
}

func (s *Server) GetTemporaryProject(ctx context.Context, req *ypb.Empty) (*ypb.ProjectDescription, error) {
	proj, err := yakit.GetTemporaryProject(s.GetProfileDatabase())
	if err != nil {
		return nil, utils.Errorf("cannot fetch temporary project")
	}
	return proj.ToGRPCModel(), nil
}
