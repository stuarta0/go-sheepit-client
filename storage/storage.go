package storage

import (
    "fmt"
    "errors"
    "os"
    "io/ioutil"
    "path/filepath"
    "crypto/md5"
    "strings"
)


func CreateWorkingDirectory(path string) error {
    if err := os.MkdirAll(path, os.ModeDir); err != nil {
        return err
    } else {
        //fmt.Println("created working directory", path)
    }
    return cleanDirectory(path)
}

func removeWorkingDirectory(path string) error {
    return errors.New("not implemented")
    // if (this.userSpecifiedACacheDir == true) {
    //      this.cleanWorkingDirectory();
    //  }
    //  else {
    //      Utils.delete(this.workingDirectory);
    //  }
}

// Only keeps zips whose MD5 hash matches their filename
// All directories removed and all other files removed
func cleanDirectory(path string) error {
    if _, err := os.Stat(path); err != nil {
        return nil
    }

    if files, err := ioutil.ReadDir(path); err != nil {
        return err
    } else {
        for _, file := range files {
            fpath := filepath.Join(path, file.Name())
            if file.IsDir() {
                os.Remove(fpath)
            } else {
                if filepath.Ext(fpath) == ".zip" {
                    if contents, err := ioutil.ReadFile(fpath); err == nil {
                        if hash := fmt.Sprintf("%x", md5.Sum(contents)); hash != strings.TrimSuffix(file.Name(), ".zip") {
                            fmt.Println("does not match hash", hash)
                            os.Remove(fpath)
                        }
                    }
                }
            }
        }
    }

    return nil
}