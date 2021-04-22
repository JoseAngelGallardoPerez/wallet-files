package files

import (
	"github.com/Confialink/wallet-files/internal/service"
	"context"
	"fmt"
	"net/http"

	"github.com/Confialink/wallet-files/internal/config"
	"github.com/Confialink/wallet-files/internal/database"
	pb "github.com/Confialink/wallet-files/rpc/files"
)

type PbServerInterface interface {
	Start()
	GetFile(ctx context.Context, req *pb.FileReq) (*pb.FileResp, error)
}

type pbServer struct {
	repo    *database.Repository
	config  *config.Config
	storage *service.StorageService
}

func NewPbServer(repo *database.Repository, config *config.Config, storage *service.StorageService) *pbServer {
	return &pbServer{repo, config, storage}
}

func (s *pbServer) Start() {
	twirpHandler := pb.NewServiceFilesServer(s, nil)
	mux := http.NewServeMux()
	mux.Handle(pb.ServiceFilesPathPrefix, twirpHandler)
	go http.ListenAndServe(fmt.Sprintf(":%s", s.config.ProtoBufPort), mux)
}

func (s *pbServer) GetFile(ctx context.Context, req *pb.FileReq) (*pb.FileResp, error) {
	if file, err := s.repo.FindByID(req.Id); err != nil {
		return nil, err
	} else {
		return &pb.FileResp{Id: file.ID}, nil
	}
}

func (s *pbServer) UserHasFiles(ctx context.Context, req *pb.UserHasFilesReq) (resp *pb.UserHasFilesResp, err error) {
	resp = &pb.UserHasFilesResp{}
	files, err := s.repo.FindAdminVisibleByUID(req.Uid, req.ExcludeCategories)
	if err != nil {
		return
	}

	if len(files) > 0 {
		resp.FilesExist = true
	}

	return
}

func (s *pbServer) UploadFile(ctx context.Context, req *pb.UploadFileReq) (resp *pb.UploadFileResp, err error) {
	resp = &pb.UploadFileResp{}
	var cat *string
	if req.Category != "" {
		cat = &req.Category
	}
	_, err = s.storage.UploadBytes(req.Bytes, req.FileName, req.Uid, req.AdminOnly, req.Private, cat)
	if err != nil {
		return
	}
	return
}

func (s *pbServer) DownloadFile(_ context.Context, req *pb.FileReq) (resp *pb.BinaryFileResp, err error) {
	file, err := s.repo.GetByID(req.Id)
	if err != nil {
		return nil, err
	}
	return &pb.BinaryFileResp{
		Data:        s.storage.Download(file),
		Size:        file.Size,
		ContentType: file.ContentType,
	}, nil
}
