package bin

/*
The bin package is involved in program setup, teardown, and the movement
and conversion of bin/xml files
*/

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	workspace      string = `tempFolder\`
	sceneryFolder  string = `Scenery\`
	networksFolder string = `Networks\`
	backupFolder   string = `AssetBackup\`
)

var (
	ErrCopyMismatch error = errors.New("error: file copy size mismatch")
	ErrNoOverwrite  error = errors.New("Overwrite of existing backups declined")
)

// Setup backs up files, moves the bin's to the temp folder, then converts bin to xml
func Setup(routeFolder string) error {
	fmt.Println("Running setup")

	if _, err := os.Stat(routeFolder + backupFolder); !os.IsNotExist(err) {
		overwrite := 'n'
		fmt.Printf("Existing backup folder found, overwrite? (y, n): ")
		for {
			fmt.Scanf("%c\n", &overwrite)
			if overwrite == 'y' || overwrite == 'Y' {
				os.RemoveAll(routeFolder + backupFolder)
				err := backupAssets(routeFolder, routeFolder+backupFolder)
				if err != nil {
					return err
				}
				break
			} else if overwrite == 'n' || overwrite == 'N' {
				fmt.Println("Existing backup retained")
				break
			}
			fmt.Printf("Please enter y or n: ")
		}
	}

	/*
		//Make directory for working on xml files
		if err := os.Mkdir(workspace, 0755); err != nil {
			return err
		}
		if err := os.Mkdir(workspace+sceneryFolder, 0755); err != nil {
			return err
		}

		// Make directory to copy all the backup .bin files to
		if err := os.Mkdir(backupFolder, 0755); err != nil {
			if e, ok := err.(*os.PathError); ok { // If err is the special error type Mkdir can return
				if os.IsExist(e) { // if the directory already exists
					overwrite := 'y'
					fmt.Println("Backup directory already exists, overwrite?")
					fmt.Scanf("%c\n", &overwrite)
					if overwrite == 'n' || overwrite == 'N' {
						Teardown(backupFolder, true)
						return ErrNoOverwrite
					}
				}
			} else {
				return err //Return the not-special error type
			}
		}

		if err := backupAssets(binFolder, backupFolder); err != nil {
			return err
		}

		if err := MoveAssetFiles(binFolder, workspace, ".bin"); err != nil {
			Teardown(backupFolder, true)
			return err
		}

		if err := SerzConvert(".bin"); err != nil {
			Teardown(backupFolder, true)
			return err
		}
	*/
	return nil
}

func Teardown(backupFolder string, removeBackups bool) {
	os.RemoveAll("tempFiles")

	if removeBackups == true {
		err := os.RemoveAll(backupFolder)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func backupAssets(srcFolder, dstFolder string) error {
	fmt.Printf("Backing up bin files")
	err := os.Mkdir(dstFolder, 0755)
	overwrite := 'n'
	if e, ok := err.(*os.PathError); ok { // If err is the special error type Mkdir can return
		if os.IsExist(e) { // if the directory already exists
			for overwrite != 'y' {
				fmt.Println("Backup directory already exists, overwrite?")
				fmt.Scanf("%c\n", &overwrite)
				if overwrite == 'n' || overwrite == 'N' {
					Teardown(backupFolder, true)
					return ErrNoOverwrite
				} else if overwrite == 'y' || overwrite == 'Y' {
					err = os.RemoveAll(dstFolder)
					if err != nil {
						return err
					}
					os.Mkdir(dstFolder, 0755)
				}
			}
		}
	}
	dotCounter := 0
	if err != nil {
		return err
	}
	err = filepath.Walk(srcFolder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(srcFolder, path) //Path relative to source folder
		if err != nil {
			return err
		}
		// Create directories, but don't copy the backup directory
		if info.IsDir() == true && !strings.Contains(strings.Trim(backupFolder, `\`), info.Name()) {
			os.Mkdir(dstFolder+relPath, 0755)
		}
		return nil
	})

	err = filepath.Walk(srcFolder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(srcFolder, path)
		if err != nil {
			return err
		}
		// Copy bin files, except ones already in the backup directory
		if info.IsDir() != true && filepath.Ext(path) == ".bin" && !strings.Contains(path, strings.Trim(backupFolder, `\`)) {
			origFile, err := os.Open(path)
			if err != nil {
				return err
			}
			defer origFile.Close()

			newFile, err := os.Create(dstFolder + relPath)
			if err != nil {
				return err
			}
			writ, err := io.Copy(newFile, origFile)
			if err != nil {
				return err
			}
			if writ != info.Size() {
				return ErrCopyMismatch
			}
			if dotCounter%50 == 0 {
				fmt.Printf(".")
			}
			dotCounter++
			return newFile.Close()
		}
		return nil
	})
	fmt.Printf("\n")
	return err
}

// Revert copies all the backed-up bin files back to the scenery directory
func Revert(routeFolder, backupFolder string) error {
	binFolder := routeFolder + sceneryFolder
	return backupAssets(backupFolder, binFolder)
}

// SerzConvert uses the DTG serz.exe to convert .bin to .xml
// ext controls the filetype to convert FROM
// Defaults to converting
func SerzConvert(ext string) error {
	fmt.Printf("Converting files")
	err := filepath.Walk(workspace, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() != true && filepath.Ext(path) == ext {
			cmd := exec.Command("./serz.exe", path)
			if err := cmd.Run(); err != nil {
				return err
			}
			err := os.Remove(path)
			if err != nil {
				return err
			}
		}
		fmt.Printf(".")
		return nil
	})
	if err != nil {
		return err
	}
	fmt.Printf("\n")
	return nil
}

// MoveAssetFiles moves .bin or .xml files form oldLoc TO newLoc, ignoring files that do not
// have the extension passed by ext
func MoveAssetFiles(oldLoc, newLoc, ext string) error {
	err := filepath.Walk(oldLoc, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() != true && filepath.Ext(path) == ext {
			newPath := newLoc + filepath.Base(path)
			err = os.Rename(path, newPath)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}
