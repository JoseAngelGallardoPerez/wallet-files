package syssettings

import (
	"context"

	"github.com/alecthomas/units"

	pb "github.com/Confialink/wallet-settings/rpc/proto/settings"

	"github.com/Confialink/wallet-files/internal/service/syssettings/connection"
)

type UserFilesStorageLimits struct {
	TotalLimitBytes    int64
	FileSizeLimitBytes int64
}

// GetTimeSettings returns new TimeSettings from settings service or err if can not get it
func GetUserFilesStorageLimits() (*UserFilesStorageLimits, error) {
	settings := UserFilesStorageLimits{}
	client, err := connection.GetSystemSettingsClient()
	if err != nil {
		return &settings, err
	}

	response, err := client.List(context.Background(), &pb.Request{Path: "regional/general/%"})
	if err != nil {
		return &settings, err
	}

	totalLimitBytes, err := units.ParseBase2Bytes(
		getSettingValue(response.Settings, "regional/general/total_user_files_storage_limit_mb") + "MB")
	if err != nil {
		return &settings, err
	}

	fileSizeLimitBytes, err := units.ParseBase2Bytes(
		getSettingValue(response.Settings, "regional/general/user_file_size_limit_mb") + "MB")
	if err != nil {
		return &settings, err
	}

	settings.TotalLimitBytes = int64(totalLimitBytes)
	settings.FileSizeLimitBytes = int64(fileSizeLimitBytes)

	return &settings, nil
}

func getSettingValue(settings []*pb.Setting, path string) string {
	for _, v := range settings {
		if v.Path == path {
			return v.Value
		}
	}
	return ""
}
