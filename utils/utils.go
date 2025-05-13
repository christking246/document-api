package utils

import (
	"crypto/rand"
	"documentApi/data"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

// GetFiles returns a list of files and directories in the specified directory path.
//
// It filters the results based on the provided extensions, and whether to include folders or files.
// The function can also search recursively in subdirectories if the deep parameter is set to true.
//
// Parameters:
// - dirPath: The directory path to search in.
// - exts: A slice of file extensions to filter the results.
// - folders: A boolean indicating whether to include directories in the results.
// - files: A boolean indicating whether to include files in the results.
// - deep: A boolean indicating whether to search recursively in subdirectories.
//
// Returns:
// - A slice of FileMetaData representing the files and directories found.
// - An error if any occurred during the operation.
func GetFiles(dirPath string, exts []string, folders bool, files bool, deep bool) ([]data.FileMetaData, error) {
	var list = []data.FileMetaData{}

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return list, err
	}

	for _, entry := range entries {
		if !entry.IsDir() && files { // if not dir assume file for now
			// if no extension are provided, return all files
			if len(exts) == 0 {
				list = append(list, data.FileMetaData{Name: entry.Name(), Path: path.Join(dirPath, entry.Name())})
			} else {
				for _, ext := range exts {
					if len(entry.Name()) < len(ext) {
						continue
					}
					// check if the file has the correct extension
					if entry.Name()[len(entry.Name())-len(ext):] == ext {
						list = append(list, data.FileMetaData{Name: entry.Name(), Path: path.Join(dirPath, entry.Name())})
					}
				}
			}
		}
		if entry.IsDir() && folders {
			list = append(list, data.FileMetaData{Name: entry.Name(), Path: path.Join(dirPath, entry.Name())})
		}
		if entry.IsDir() && deep {
			subEntries, err := GetFiles(path.Join(dirPath, entry.Name()), exts, folders, files, deep)
			if err != nil { // maybe a sub folder is not accessible, should we hard fail?
				return list, err
			}
			list = append(list, subEntries...)
		}
	}

	return list, nil
}

// getNextArchiveNumber returns the next available archive number as a string.
//
// It does this by checking the existence of files in the log directory with a specific naming pattern.
// The function starts with an initial archive number of 1 and increments it until it finds a number
// that does not correspond to an existing file in the log directory.
//
// Return:
// - The next available archive number as a string.
func getNextArchiveNumber(logName string) string {
	current := 1
	logDir := filepath.Join(os.Getenv("LOG_DIR"), logName+".")

	for {
		filePath := logDir + strconv.Itoa(current)
		_, err := os.Stat(filePath)
		if os.IsNotExist(err) {
			break
		}
		current++
	}

	return strconv.Itoa(current)
}

// InitDir initializes a directory by creating it if it does not exist.
//
// It takes a directory path and a logger as parameters.
// If the directory does not exist, it attempts to create it.
func InitDir(dir string, logger *logrus.Logger) bool {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err2 := os.MkdirAll(dir, os.ModePerm)
		if err2 != nil {
			logger.Error("Error creating directory: " + dir + " - " + err2.Error())
			return false
		} else {
			logger.Debug("Directory created: " + dir)
			return true
		}
	}
	return true
}

// SetupLogger sets up the logger based on the environment.
//
// It loads the environment from a .env file, sets the log level based on the environment,
// and configures the logger to output to either stdout or a file depending on the environment.
//
// Returns a file pointer and a logrus logger instance.
func SetupLogger(logName string) (*os.File, *logrus.Logger) {
	// Load env
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
	}

	var env = os.Getenv("ENV")
	if env == "" {
		env = "dev"
	}

	if env == "test" {
		// do we actually want to log, if in test mode?
		var testLogger = logrus.New()
		testLogger.SetLevel(logrus.InfoLevel)
		testLogger.SetOutput(os.Stdout)
		testLogger.SetFormatter(&logrus.TextFormatter{
			ForceColors:   true,
			FullTimestamp: true,
		})
		return nil, testLogger
	}

	// create dir for the log file if it doesn't yet exist
	if _, err := os.Stat(os.Getenv("LOG_DIR")); os.IsNotExist(err) {
		err2 := os.MkdirAll(os.Getenv("LOG_DIR"), os.ModePerm)
		if err2 != nil {
			fmt.Println("Error creating log directory: " + os.Getenv("LOG_DIR"))
			fmt.Println(err)
		} else {
			fmt.Println("Created log directory")
		}
	}

	// setup logger
	var Logger = logrus.New()
	var isProd = env == "prod" || env == "production"
	if isProd {
		Logger.SetLevel(logrus.InfoLevel)
	} else {
		Logger.SetLevel(logrus.DebugLevel)
	}
	Logger.SetFormatter(&logrus.TextFormatter{
		ForceColors:   true, // !isProd,
		FullTimestamp: true,
	})

	// archive current log files if exist
	logFilePath := filepath.Join(os.Getenv("LOG_DIR"), logName)
	_, statErr := os.Stat(logFilePath)
	if statErr == nil {
		fmt.Println("Found existing log file: " + logFilePath)
		var archiveNumber = getNextArchiveNumber(logName)
		renameError := os.Rename(logFilePath, logFilePath+"."+archiveNumber)
		if renameError != nil {
			fmt.Println("Error renaming existing log file: " + renameError.Error())
		}
	}

	// create log file
	file, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY, 0644)

	// setup outputs for the logger
	if isProd && err == nil {
		Logger.SetOutput(file)
		return file, Logger
	} else if isProd && err != nil {
		Logger.SetOutput(os.Stdout)
		Logger.Error("Failed to setup logger with file: " + err.Error())
		return nil, Logger
	} else if !isProd && err == nil {
		// Log to the file in addition to stdout
		Logger.SetOutput(io.MultiWriter(os.Stdout, file))
		return file, Logger
	} else {
		Logger.SetOutput(os.Stdout)
		Logger.Warn(err)
		Logger.Warn("Unable to setup logger with a file")
		return nil, Logger
	}
}

// Base returns the last element of a path p.
// This is a wrapper for the standard library's path.Base()
// since it only handles unix type paths
func Base(p string) string {
	if p == "" {
		return "."
	}

	// handle windows paths by switching the slashes
	if os.PathSeparator == '\\' {
		p = filepath.ToSlash(p)
	}

	// should I swap back the slashes?
	return path.Base(p)
}

// Dir returns all but the last element of path, typically the path's directory.
// This is a wrapper for the standard library's path.Dir()
// since it only handles unix type paths
func Dir(p string) string {
	if p == "" {
		return "."
	}

	// handle windows paths by switching the slashes
	if os.PathSeparator == '\\' {
		p = filepath.ToSlash(p)
	}

	// should I swap back the slashes?
	return path.Dir(p)
}

func GenerateId() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x-%x-%x-%x-%x%x-%d", b[0:4], b[4:6], b[6:8], b[8:10], b[10:12], b[12:], time.Now().UnixMilli())
}
