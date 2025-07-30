// service/mcServerManager.go
package service

import (
	"encoding/json"
	"fmt"
	"go-backend/common"
	"net/http"
	"os"
	"path/filepath"
)

type CreateServerRequest struct {
	ServerType      string `json:"server_type"`
	ServerVer       string `json:"server_ver"`
	FabricLoader    string `json:"fabric_loader"`
	FabricInstaller string `json:"fabric_installer"`
	DisplayName     string `json:"display_name"`
}

type GameVersion struct {
	Version string `json:"version"`
	Stable  bool   `json:"stable"`
}

func ErrorFileClear(path string) error {
	if err := os.RemoveAll(path); err != nil {
		return fmt.Errorf("cleanup failed for %s: %w", path, err)
	}
	return nil
}

func CreateServer(ownerID string, serverType string, serverVer string, fabricLoader string, fabricInstaller string) (string, error) {
	var idPerFix, fURL, vURL string
	var err error

	if fabricLoader == "" {
		fabricLoader = common.LatestFabricLoaderVersion // 預設值
	}

	if fabricInstaller == "" {
		fabricInstaller = common.LatestFabricInstallerVersion
	}

	switch serverType {
	case "Fabric":
		idPerFix = "mcsfv-"
		fURL = fmt.Sprintf(
			"https://meta.fabricmc.net/v2/versions/loader/%s/%s/%s/server/jar",
			serverVer, fabricLoader, fabricInstaller,
		)
	case "Vanilla":
		if idPerFix == "" {
			idPerFix = "mcsvv-"
		}
		url, ok := common.VanillaServerUrl[serverVer]
		if !ok {
			return "", fmt.Errorf("unsupported server version: %s", serverVer)
		}
		vURL = url
	default:
		return "", fmt.Errorf("unsupported server type: %s", serverType)
	}

	uid := common.GetRandomIntString(4)
	serverID := idPerFix + serverVer + "-" + uid + "-" + "OID-" + ownerID

	sysPath := filepath.Join(common.MinecraftServerPath, serverID)
	// defer 一個清理機制：若後續 err != nil，就把 sysPath 刪掉
	defer func() {
		if err != nil {
			if clearErr := ErrorFileClear(sysPath); clearErr != nil {
				msg := fmt.Sprintf("warning: %v", clearErr)
				common.SysLog(msg)
			}
		}
	}()

	if err = os.MkdirAll(sysPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create server directory %s: %w", sysPath, err)
	}

	if vURL != "" {
		vanillaJarPath := filepath.Join(sysPath, "server.jar")
		if err = common.DownloadFile(vanillaJarPath, vURL); err != nil {
			return "", fmt.Errorf("failed to download vanilla server jar: %w", err)
		}
	}

	if fURL != "" {
		fabricInstallerPath := filepath.Join(sysPath, "fabric-erver.jar")
		if err = common.DownloadFile(fabricInstallerPath, fURL); err != nil {
			return "", fmt.Errorf("failed to download fabric installer: %w", err)
		}
	}

	eulaPath := filepath.Join(sysPath, "eula.txt")
	eulaContent := []byte("eula=true\n")
	if err = os.WriteFile(eulaPath, eulaContent, 0644); err != nil {
		return "", fmt.Errorf("failed to write eula.txt: %w", err)
	}

	return serverID, nil
}

func GetAllFabricVersions() ([]string, error) {
	resp, err := http.Get("https://meta.fabricmc.net/v2/versions/game")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get fabric versions, status code: %d", resp.StatusCode)
	}

	var gv []GameVersion
	if err := json.NewDecoder(resp.Body).Decode(&gv); err != nil {
		return nil, err
	}

	versions := make([]string, len(gv))
	for i, v := range gv {
		versions[i] = v.Version
	}
	return versions, nil
}

func GetAllVanillaVersions() (map[string]string, error) {
	all := common.VanillaServerUrl
	return all, nil
}

func StartServer(serverID string) error {
	return nil
}
