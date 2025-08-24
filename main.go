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

	"github.com/bingoohuang/ngg/ss"
	"github.com/mitchellh/go-homedir"
)

// nolint gochecknoglobals
var (
	pkgName      string
	targetDir    string
	disableCache bool
)

// nolint
func init() {
	version := false

	flag.BoolVar(&disableCache, "disableCache", false, "disable cache of go-starter project downloading")
	flag.BoolVar(&version, "v", false, "show version")
	flag.StringVar(&targetDir, "dir", ".", "target directory")
	flag.StringVar(&pkgName, "pkg", "", "package name, default to last element of target directory")
	flag.Parse()

	if version {
		fmt.Println("v2020-04-28")
		os.Exit(0)
	}

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

type fileInterceptor interface {
	match(filename string) bool
	intercept(src []byte) []byte
}

type gogoInterceptor struct {
	PkgName      []byte
	PkgSnakeCase []byte
}

var _ fileInterceptor = (*gogoInterceptor)(nil)

func (r gogoInterceptor) match(filename string) bool {
	switch filepath.Ext(filename) {
	case ".go", ".md", ".mod", ".html":
		return true
	default:
		return false
	}
}

func (r gogoInterceptor) intercept(src []byte) []byte {
	bs := bytes.Replace(src, []byte("github.com/bingoohuang/gostarter"), r.PkgName, -1)
	bs = bytes.Replace(bs, []byte("gostarter"), r.PkgName, -1)

	return bytes.Replace(bs, []byte("GOSTARTER"), r.PkgSnakeCase, -1)
}

func main() {
	zipFile := download()
	interceptor := gogoInterceptor{
		PkgName:      []byte(pkgName),
		PkgSnakeCase: []byte(ss.ToSnakeUpper(pkgName)),
	}

	if err := Unzip(zipFile, targetDir, interceptor); err != nil {
		log.Fatal(err)
	}

	fmt.Println(pkgName + " created successfully in " + targetDir + "!")
}

const starterZipURL = "https://github.com/bingoohuang/gostarter/archive/master.zip"

func download() string {
	cacheZip, _ := homedir.Expand("~/.gostarter/master.zip")
	cacheExists := false

	if !disableCache {
		if st, err := os.Stat(cacheZip); err == nil {
			cacheExists = true
			// cache will be expired in 10 days
			if st.ModTime().Add(240 * time.Hour).After(time.Now()) { // nolint gomnd
				fmt.Printf("cache %s found!\n", cacheZip)
				return cacheZip
			}

			fmt.Printf("cache %s expired in 10 days!\n", cacheZip)
		}
	}

	cacheZipDir, _ := homedir.Expand("~/.gostarter/")
	_ = os.MkdirAll(cacheZipDir, os.ModePerm)

	fmt.Printf("start to download %s\n", starterZipURL)

	if err := DownloadFile(cacheZip, starterZipURL); err != nil {
		if !cacheExists {
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
	resp, err := http.Get(url) // nolint gosec
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
func Unzip(src, dest string, interceptor fileInterceptor) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		fn := strings.TrimPrefix(f.Name, "gostarter-master/")
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
			_ = os.MkdirAll(fpath, os.ModePerm)
			continue
		}

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

		if interceptor.match(fn) {
			buf := new(bytes.Buffer)
			_, _ = buf.ReadFrom(rc)
			src := interceptor.intercept(buf.Bytes())
			_, _ = outFile.Write(src)
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
