package main

import (
	"io"
	"os"
	"path"
)

type renameFile struct {
	*os.File
	filename string
}

func (f *renameFile) Close() error {
	f.File.Sync()
	
	if err := f.File.Close(); err != nil {
		return err
	}
	return os.Rename(f.File.Name(), f.filename)
}

func create(filename string) (io.WriteCloser, error) {
	tmpFilename := filename + ".tmp"
	
	if err := EnsureBaseDir(tmpFilename);err != nil{
		return nil, err
	}
	
	f, err := os.Create(tmpFilename)
	if err != nil {
		return nil, err
	}
	
	rf := &renameFile{
		File:     f,
		filename: filename,
	}
	return rf, nil
}

func EnsureBaseDir(fpath string) error {
	baseDir := path.Dir(fpath)
	info, err := os.Stat(baseDir)
	if err == nil && info.IsDir() {
		return nil
	}
	return os.MkdirAll(baseDir, 0755)
}