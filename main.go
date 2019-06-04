package main

import (
	"archive/zip"
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mitchellh/go-homedir"
)

var pkgName string
var targetDir string
var disableCache bool

func init() {
	flag.BoolVar(&disableCache, "disableCache", false, "disable cache of go-starter project downloading")
	flag.StringVar(&targetDir, "dir", ".", "target directory")
	flag.StringVar(&pkgName, "pkg", "", "package name, default to last element of target directory")

	flag.Parse()

	if targetDir == "" {
		targetDir, _ = os.Getwd()
	} else {
		targetDir, _ = homedir.Expand(targetDir)
	}

	_ = os.MkdirAll(targetDir, os.ModePerm)

	if pkgName == "" {
		pkgName = filepath.Base(targetDir)
	}
}

type FileInterceptor interface {
	Match(filename string) bool
	Intercept(src []byte) []byte
}

type GogoInterceptor struct {
	PkgName []byte
}

var _ FileInterceptor = (*GogoInterceptor)(nil)

func (r GogoInterceptor) Match(filename string) bool {
	ext := filepath.Ext(filename)
	switch ext {
	case ".go", ".md", ".mod", ".html":
		return true
	default:
		return false
	}
}

func (r GogoInterceptor) Intercept(src []byte) []byte {
	return bytes.Replace(src, []byte("go-starter"), r.PkgName, -1)
}

func main() {
	zipFile := Download()
	interceptor := GogoInterceptor{PkgName: []byte(pkgName)}
	if err := Unzip(zipFile, targetDir, interceptor); err != nil {
		log.Fatal(err)
	}

	fmt.Println(pkgName + " created successfully in " + targetDir + "!")
}

func Download() string {
	cacheZip, _ := homedir.Expand("~/.go-starter/master.zip")
	exists := false
	if !disableCache {
		if st, err := os.Stat(cacheZip); err == nil {
			exists = true
			// cache will be expired in 10 days
			if st.ModTime().Add(time.Duration(240) * time.Hour).After(time.Now()) {
				fmt.Printf("cache %s found!\n", cacheZip)
				return cacheZip
			} else {
				fmt.Printf("cache %s expired in 10 days!\n", cacheZip)
			}
		}
	}

	cacheZipDir, _ := homedir.Expand("~/.go-starter/")
	_ = os.MkdirAll(cacheZipDir, os.ModePerm)
	starterZipURL := "https://github.com/bingoohuang/go-starter/archive/master.zip"
	fmt.Printf("start to download %s\n", starterZipURL)
	if err := DownloadFile(cacheZip, starterZipURL); err != nil {
		if !exists {
			log.Fatal(err)
		}
		fmt.Printf("failed to download %v, use cached instead!\n", err)
	}

	return cacheZip
}

// DownloadFile will download a url to a local file. It's efficient because it will
// write as it downloads and not load the whole file into memory.
func DownloadFile(destFile, url string) error {
	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	// Create the file
	out, err := os.Create(destFile)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}

// Unzip will decompress a zip archive, moving all files and folders
// within the zip file (parameter 1) to an output directory (parameter 2).
func Unzip(src, dest string, interceptor FileInterceptor) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		fn := f.Name
		if strings.HasPrefix(fn, "go-starter-master/") {
			fn = fn[len("go-starter-master/"):]
		}

		if fn == "" {
			continue
		}

		// Store filename/path for returning and using later on
		fpath := filepath.Join(dest, fn)

		// Check for ZipSlip. More Info: http://bit.ly/2MsjAWE
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("%s: illegal file path", fpath)
		}

		if f.FileInfo().IsDir() {
			// Make Folder
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		// Make File
		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			return err
		}

		if interceptor.Match(fn) {
			buf := new(bytes.Buffer)
			buf.ReadFrom(rc)
			src := interceptor.Intercept(buf.Bytes())
			outFile.Write(src)
		} else {
			_, err = io.Copy(outFile, rc)
		}

		// Close the file without defer to close before next iteration of loop
		outFile.Close()
		rc.Close()

		if err != nil {
			return err
		}
	}
	return nil
}
