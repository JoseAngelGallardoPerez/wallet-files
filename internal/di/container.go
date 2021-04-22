package di

import (
	"fmt"
	"log"
	"os"

	"github.com/Confialink/wallet-pkg-env_config"
	"github.com/Confialink/wallet-pkg-env_mods"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/inconshreveable/log15"
	"github.com/jinzhu/gorm"
	"github.com/kildevaeld/go-acl"

	"github.com/Confialink/wallet-files/internal/auth"
	"github.com/Confialink/wallet-files/internal/config"
	"github.com/Confialink/wallet-files/internal/database"
	"github.com/Confialink/wallet-files/internal/policy"
	"github.com/Confialink/wallet-files/internal/service"
	"github.com/Confialink/wallet-files/internal/storage"
	files "github.com/Confialink/wallet-files/rpc"
)

var Container *container

type container struct {
	appConfig           config.Config
	dbConnection        *gorm.DB
	repository          *database.Repository
	authService         auth.ServiceInterface
	acl                 *acl.ACL
	storageS3           *storage.S3
	storageLocal        *storage.Local
	storageService      *service.StorageService
	s3Uploader          *s3manager.Uploader
	s3Downloader        *s3manager.Downloader
	s3                  *s3.S3
	awsCredentials      *credentials.Credentials
	awsSession          *session.Session
	awsConfig           *aws.Config
	pbServer            files.PbServerInterface
	permissionsService  *policy.PermissionsService
	usersService        *service.Users
	serviceLogger       log15.Logger
}

func init() {
	Container = &container{}
	readConfig(&Container.appConfig)
	validateConfig(Container.Config(), Container.ServiceLogger().New("service", "configReader"))
}

// Config returns config
func (c *container) Config() *config.Config {
	return &c.appConfig
}

// DbConnection creates new DB connection if not exists and return
func (c *container) DbConnection() *gorm.DB {
	var err error
	if nil == c.dbConnection {
		c.dbConnection, err = database.NewConnection(c.Config().Db)
		if nil != err {
			fmt.Printf("Can't establish connection to backend: %v", err)
			log.Fatalf("Can't establish connection to backend: %v", err)
		}
	}
	return c.dbConnection
}

// Repository creates new repository if not exists and return
func (c *container) Repository() *database.Repository {
	if nil == c.repository {
		c.repository = database.NewRepository(c.DbConnection())
	}

	return c.repository
}

// StorageService creates new storage service if not exists and return
func (c *container) StorageService() *service.StorageService {
	if c.storageService == nil {
		pool := map[string]storage.Storage{
			storage.StorageS3:    c.StorageS3(),
			storage.StorageLocal: c.StorageLocal(),
		}

		c.storageService = service.NewStorageService(pool, c.Config(), c.Repository())
	}

	return c.storageService
}

// StorageLocal creates new s3 storage service if not exists and return
func (c *container) StorageLocal() *storage.Local {
	if nil == c.storageLocal {
		c.storageLocal = storage.NewLocal(
			c.Repository(),
		)
	}

	return c.storageLocal
}

// StorageS3 creates new s3 storage service if not exists and return
func (c *container) StorageS3() *storage.S3 {
	if nil == c.storageS3 {
		c.storageS3 = storage.NewS3(
			c.S3Uploader(),
			c.S3Downloader(),
			c.S3(),
			c.Config().AwsConfig,
			c.Repository(),
		)
	}

	return c.storageS3
}

// S3Uploader creates new s3 uploader if not exists and return
func (c *container) S3Uploader() *s3manager.Uploader {
	if nil == c.s3Uploader {
		c.s3Uploader = s3manager.NewUploader(c.AwsSession())
	}

	return c.s3Uploader
}

// S3Downloader creates new s3 downloader if not exists and return
func (c *container) S3Downloader() *s3manager.Downloader {
	if nil == c.s3Downloader {
		c.s3Downloader = s3manager.NewDownloader(c.AwsSession())
	}

	return c.s3Downloader
}

// S3 creates new s3 instance if not exists and return
func (c *container) S3() *s3.S3 {
	if nil == c.s3 {
		c.s3 = s3.New(c.AwsSession())
	}

	return c.s3
}

// AwsCredentials creates new aws credentials if not exists and return
func (c *container) AwsCredentials() *credentials.Credentials {
	if nil == c.awsCredentials {
		c.awsCredentials = credentials.NewEnvCredentials()
	}

	return c.awsCredentials
}

// AwsSession creates new aws session if not exists and return
func (c *container) AwsSession() *session.Session {
	if nil == c.awsSession {
		c.awsSession, _ = session.NewSession(c.AwsConfig())
	}

	return c.awsSession
}

func (c *container) AwsConfig() *aws.Config {
	if nil == c.awsConfig {
		c.awsConfig = &aws.Config{
			Region:      aws.String(c.Config().AwsConfig.Region),
			Credentials: c.AwsCredentials(),
		}
	}
	return c.awsConfig
}

// PbServer creates new proto buf server if not exists and return
func (c *container) PbServer() files.PbServerInterface {
	if nil == c.pbServer {
		c.pbServer = files.NewPbServer(c.Repository(), c.Config(), c.StorageService())
	}

	return c.pbServer
}

// AuthService creates new auth service if not exists and return
func (c *container) AuthService() auth.ServiceInterface {
	if nil == c.authService {
		c.authService = auth.NewService(c.Acl(), c.PermissionsService())
	}

	return c.authService
}

func (c *container) PermissionsService() *policy.PermissionsService {
	if nil == c.permissionsService {
		c.permissionsService = policy.NewPermissionsService(c.UsersService(), c.ServiceLogger().New("service", "PermissionsService"))
	}

	return c.permissionsService
}

func (c *container) UsersService() *service.Users {
	if nil == c.usersService {
		c.usersService = service.NewUsers()
	}

	return c.usersService
}

// Acl creates new acl if not exists and return
func (c *container) Acl() *acl.ACL {
	if nil == c.acl {
		c.acl = acl.New(acl.NewMemoryStore())
	}

	return c.acl
}

func (c *container) ServiceLogger() log15.Logger {
	if c.serviceLogger == nil {
		c.serviceLogger = log15.New("service", "files")
	}
	return c.serviceLogger
}

// readConfig reads configs from ENV variables
func readConfig(cfg *config.Config) {
	cfg.Port = os.Getenv("VELMIE_WALLET_FILES_PORT")
	cfg.ProtoBufPort = os.Getenv("VELMIE_WALLET_FILES_PROTO_BUF_PORT")
	cfg.Env = env_config.Env("ENV", env_mods.Development)
	cfg.Storage = os.Getenv("VELMIE_WALLET_FILES_STORAGE")
	cfg.AwsConfig = readAwsConfig()

	defaultConfigReader := env_config.NewReader("files")
	cfg.Cors = defaultConfigReader.ReadCorsConfig()
	cfg.Db = defaultConfigReader.ReadDbConfig()
}

func validateConfig(cfg *config.Config, logger log15.Logger) {
	validator := env_config.NewValidator(logger)
	validator.ValidateCors(cfg.Cors, logger)
	validator.ValidateDb(cfg.Db, logger)
	validator.CriticalIfEmpty(cfg.Port, "VELMIE_WALLET_FILES_PORT", logger)
	validator.CriticalIfEmpty(cfg.ProtoBufPort, "VELMIE_WALLET_FILES_PROTO_BUF_PORT", logger)
	validator.CriticalIfEmpty(cfg.Storage, "VELMIE_WALLET_FILES_STORAGE", logger)
}

// readAwsConfig reads AWS configs from ENV variables
func readAwsConfig() config.AwsConfig {
	awsConfig := config.AwsConfig{
		S3Bucket: os.Getenv("VELMIE_WALLET_FILES_AWS_S3_BUCKET"),
		Region:   os.Getenv("VELMIE_WALLET_FILES_AWS_REGION"),
	}
	return awsConfig
}
