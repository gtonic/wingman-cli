package fs

import (
	"os"
	"path/filepath"
	"time"
)

func New(root string) *FileSystem {
	os.MkdirAll(root, 0755)

	return &FileSystem{
		root: root,
	}
}

type FileSystem struct {
	root string
}

type FileInfo struct {
	Path string `json:"path"`
	Name string `json:"name"`

	Size      int64     `json:"size"`
	Timestamp time.Time `json:"timestamp"`
}

func (fs *FileSystem) ListFiles() ([]FileInfo, error) {
	var result []FileInfo

	filepath.Walk(fs.root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if info.IsDir() {
			return nil
		}

		path, _ = filepath.Rel(fs.root, path)
		name := filepath.Base(path)

		if path == "." {
			return nil
		}

		file := FileInfo{
			Path: path,
			Name: name,

			Size:      info.Size(),
			Timestamp: info.ModTime(),
		}

		result = append(result, file)

		return nil
	})

	return result, nil
}

func (fs *FileSystem) CreateFile(path, content string) error {
	path = filepath.Join(fs.root, path)

	os.MkdirAll(filepath.Dir(path), 0755)
	return os.WriteFile(path, []byte(content), 0644)
}

func (fs *FileSystem) ReadFile(path string) (string, error) {
	path = fs.resolvePath(path)

	data, err := os.ReadFile(path)

	if err != nil {
		return "", err
	}

	return string(data), nil
}

func (fs *FileSystem) DeleteFile(path string) error {
	path = fs.resolvePath(path)

	os.Remove(path)

	for dir := filepath.Dir(path); dir != fs.root; dir = filepath.Dir(dir) {
		if err := os.Remove(dir); err != nil {
			if !os.IsNotExist(err) && !os.IsExist(err) {
				continue
			}

			break
		}
	}

	return nil
}

func (fs *FileSystem) CreateDir(path string) error {
	path = fs.resolvePath(path)
	return os.MkdirAll(path, 0755)
}

func (fs *FileSystem) DeleteDir(path string) error {
	path = fs.resolvePath(path)
	return os.RemoveAll(path)
}

func (fs *FileSystem) resolvePath(path string) string {
	return filepath.Join(fs.root, filepath.FromSlash(path))
}
