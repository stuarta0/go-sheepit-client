package storage

import (
    "archive/zip"
    "errors"
    "fmt"
    "os"
    "io"
    "io/ioutil"
    "path/filepath"
    "crypto/md5"
    "strings"
)


func CreateWorkingDirectory(path string) error {
    if err := os.MkdirAll(path, 0755); err != nil {
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
                            os.Remove(fpath)
                        }
                    }
                }
            }
        }
    }

    return nil
}

// http://stackoverflow.com/a/24792688
func Extract(src, dest string) error {
    r, err := zip.OpenReader(src)
    if err != nil {
        return err
    }
    defer func() {
        if err := r.Close(); err != nil {
            panic(err)
        }
    }()

    os.MkdirAll(dest, 0755)

    // Closure to address file descriptors issue with all the deferred .Close() methods
    extractAndWriteFile := func(f *zip.File) error {
        rc, err := f.Open()
        if err != nil {
            return err
        }
        defer func() {
            if err := rc.Close(); err != nil {
                panic(err)
            }
        }()

        path := filepath.Join(dest, f.Name)

        if f.FileInfo().IsDir() {
            os.MkdirAll(path, f.Mode())
        } else {
            os.MkdirAll(filepath.Dir(path), f.Mode())
            f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
            if err != nil {
                return err
            }
            defer func() {
                if err := f.Close(); err != nil {
                    panic(err)
                }
            }()

            _, err = io.Copy(f, rc)
            if err != nil {
                return err
            }
        }
        return nil
    }

    for _, f := range r.File {
        err := extractAndWriteFile(f)
        if err != nil {
            return err
        }
    }

    return nil
}