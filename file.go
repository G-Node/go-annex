package gannex

import (
	"os"
	"path/filepath"
	"strings"
	"io"
	"fmt"
	"log"
	"regexp"
)

type AFile struct {
	Filepath  string
	OFilename string
	Info      os.FileInfo
}

type AnnexFileNotFound struct {
	error
}

func NewAFile(annexpath, repopath, Ofilename string, APFileC []byte) (*AFile, error) {
	nAF := &AFile{OFilename: Ofilename}
	secPar := regexp.MustCompile(`(\.\.)`)
	if secPar.Match(APFileC){
		return nil, fmt.Errorf("Path not allowed")
	}
	// see https://regexper.com/#%5B%5C%5C%5C%2F%5Dannex%5B%5C%5C%5C%2F%5D(%5B%5E%5C.%5D%2B(%5C.%5Cw%2B))%3F
	aFPattern := regexp.MustCompile(`[\\\/]annex[\\\/](.+)`)
	matches := aFPattern.FindStringSubmatch(string(APFileC))
	log.Printf("matched: %v", matches)
	if matches != nil && len(matches) > 1 {
		filepath := strings.Replace(matches[1], "\\", "/", 0)
		filepath = fmt.Sprintf("%s/annex/%s", repopath, filepath)
		log.Printf("Filepath: %s", filepath)
		info, err := os.Stat(filepath)
		if err == nil {
			nAF.Filepath = filepath
			nAF.Info = info
			return nAF, nil
		}
	}

	pathParts := strings.SplitAfter(string(APFileC), string(os.PathSeparator))
	filename := strings.TrimSpace(pathParts[len(pathParts)-1])
	// lets find the annex file
	filepath.Walk(filepath.Join(annexpath, repopath), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("%v", err)
			return filepath.SkipDir
		}
		if info.IsDir() {
			return nil
		} else if info.Name() == filename {
			nAF.Filepath = path
			nAF.Info = info
			return io.EOF
		}
		return nil
	})
	if nAF.Filepath != "" {
		return nAF, nil
	} else {
		return nil, AnnexFileNotFound{error: fmt.Errorf("Could not find File: %s anywhere below: %s", filename,
			filepath.Join(annexpath, repopath))}
	}

}

func (af *AFile) Open() (*os.File, error) {
	fp, err := os.Open(af.Filepath)
	if err != nil {
		return nil, err
	}
	return fp, nil

}
