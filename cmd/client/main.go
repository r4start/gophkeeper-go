package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"os"

	"github.com/r4start/goph-keeper/cmd/client/cfg"
	"github.com/r4start/goph-keeper/cmd/client/cmd"
	"github.com/r4start/goph-keeper/internal/client/storage"
)

//go:generate sh -c "git branch --show-current > branch.txt"
//go:generate sh -c "printf %s $(git rev-parse HEAD) > commit.txt"
//go:generate sh -c "sh -c 'date +%Y-%m-%dT%H:%M:%S' > date.txt"

const (
	_cfgDirName         = ".gophkeeper"
	_cfgFileName        = "cfg.json"
	_cfgStorageFileName = "main.db"
)

//go:embed *
var buildInfo embed.FS

func main() {
	if err := errMain(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func errMain() error {
	c, err := prepareConfig()
	if err != nil {
		return err
	}

	ls, err := storage.NewLocalStorage(c.StoragePath)
	if err != nil {
		return err
	}

	registerCmd, err := cmd.NewRegisterCommand(c, ls)
	if err != nil {
		return err
	}

	authCmd, err := cmd.NewAuthCommand(c, ls)
	if err != nil {
		return err
	}

	storeCmd, err := cmd.NewStoreCommand(c, ls)
	if err != nil {
		return err
	}

	syncCmd, err := cmd.NewSyncCommand(c, ls)
	if err != nil {
		return err
	}

	delCmd, err := cmd.NewDeleteCommand(c, ls)
	if err != nil {
		return err
	}

	listCmd, err := cmd.NewListCommand(c, ls)
	if err != nil {
		return err
	}

	rootCmd := cmd.NewRootCommand()
	rootCmd.AddCommand(registerCmd.Command)
	rootCmd.AddCommand(authCmd.Command)
	rootCmd.AddCommand(storeCmd.Command)
	rootCmd.AddCommand(syncCmd.Command)
	rootCmd.AddCommand(delCmd.Command)
	rootCmd.AddCommand(listCmd.Command)

	rootCmd.Version = generateVersion()

	if err := rootCmd.Execute(); err != nil {
		return err
	}

	if err := c.Save(); err != nil {
		return err
	}
	return nil
}

func prepareConfig() (*cfg.Config, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(configDir)
	if err != nil {
		return nil, err
	}

	dirFound := false
	for _, entry := range entries {
		if entry.Name() == _cfgDirName && entry.IsDir() {
			dirFound = true
			break
		}
	}

	if err := os.Chdir(configDir); err != nil {
		return nil, err
	}

	if !dirFound {
		if err := os.Mkdir(_cfgDirName, 0700); err != nil {
			return nil, err
		}
	}

	entries, err = os.ReadDir(_cfgDirName)
	if err != nil {
		return nil, err
	}

	dirFound = false
	for _, entry := range entries {
		if entry.Name() == _cfgFileName {
			dirFound = true
			break
		}
	}

	if err := os.Chdir(_cfgDirName); err != nil {
		return nil, err
	}

	var config *cfg.Config = nil
	if dirFound {
		config, err = cfg.NewConfigFromFile(_cfgFileName)
		if err != nil {
			return nil, err
		}
	} else {
		config = cfg.NewConfig()
		data, err := json.Marshal(config)
		if err != nil {
			return nil, err
		}

		if err := os.WriteFile(_cfgFileName, data, 0700); err != nil {
			return nil, err
		}
	}

	if len(config.StoragePath) == 0 {
		config.StoragePath = _cfgStorageFileName
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	if len(config.SyncDirectory) == 0 {
		config.SyncDirectory = fmt.Sprintf("%s%c%s", homeDir, os.PathSeparator, _cfgDirName)
	}

	entries, err = os.ReadDir(homeDir)
	if err != nil {
		return nil, err
	}
	dirFound = false
	for _, entry := range entries {
		if entry.Name() == _cfgDirName {
			dirFound = true
			break
		}
	}

	if !dirFound {
		if err := os.Mkdir(config.SyncDirectory, 0700); err != nil {
			return nil, err
		}
	}

	return config, nil
}

func generateVersion() string {
	buildVersion := "N/A\n"
	buildDate := buildVersion
	buildCommit := buildVersion

	if data, err := buildInfo.ReadFile("branch.txt"); err == nil {
		buildVersion = string(data)
	}

	if data, err := buildInfo.ReadFile("commit.txt"); err == nil {
		buildCommit = string(data)
	}

	if data, err := buildInfo.ReadFile("date.txt"); err == nil {
		buildDate = string(data)
	}

	return fmt.Sprintf("\nBuild version: %sBuild date: %sBuild commit: %s",
		buildVersion, buildDate, buildCommit)
}
