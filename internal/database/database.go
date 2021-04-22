package database

import (
	"fmt"

	"github.com/Confialink/wallet-pkg-env_config"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

// CreateConnection creates a new db connection
func NewConnection(config *env_config.Db) (*gorm.DB, error) {
	// initialize a new db connection
	return gorm.Open(
		config.Driver,
		fmt.Sprintf(
			"%s:%s@tcp(%s:%s)/%s?parseTime=true", // username:password@protocol(host)/dbname?param=value
			config.User, config.Password, config.Host, config.Port, config.Schema,
		),
	)
}
