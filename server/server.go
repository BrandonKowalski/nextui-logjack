package server

import (
	"context"
	"embed"
	"fmt"
	"html/template"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"nextui-logjack/internal"
	"nextui-logjack/utils"
)

//go:embed templates/*.html
var templatesFS embed.FS

const ServerPort = 8080

type Server struct {
	httpServer  *http.Server
	listener    net.Listener
	url         string
	directories map[string]string
}

type FileInfo struct {
	Name    string
	Size    string
	ModTime string
	IsDir   bool
	IsText  bool
	Path    string
}

func New(config *internal.Config) *Server {
	dirs := make(map[string]string)
	for _, d := range config.Directories {
		dirs[d.Name] = d.Path
	}
	return &Server{directories: dirs}
}

func (s *Server) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/delete/", s.handleDelete)
	mux.HandleFunc("/view/", s.handleView)
	mux.HandleFunc("/download/", s.handleDownload)
	mux.HandleFunc("/", s.handleRequest)

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", ServerPort))
	if err != nil {
		return fmt.Errorf("failed to start listener: %w", err)
	}

	s.listener = listener
	s.httpServer = &http.Server{
		Handler: mux,
	}

	localIP := utils.GetLocalIP()
	s.url = fmt.Sprintf("http://%s:%d", localIP, ServerPort)

	go s.httpServer.Serve(listener)

	return nil
}

func (s *Server) Stop() error {
	if s.httpServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return s.httpServer.Shutdown(ctx)
	}
	return nil
}

func (s *Server) URL() string {
	return s.url
}

func (s *Server) handleRequest(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/")

	if path == "" {
		s.handleRoot(w, r)
		return
	}

	parts := strings.SplitN(path, "/", 2)
	dirName := parts[0]
	subPath := ""
	if len(parts) > 1 {
		subPath = parts[1]
	}

	basePath, ok := s.directories[dirName]
	if !ok {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	fullPath := filepath.Join(basePath, subPath)

	if !strings.HasPrefix(filepath.Clean(fullPath), filepath.Clean(basePath)) {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	info, err := os.Stat(fullPath)
	if err != nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	if !info.IsDir() {
		http.ServeFile(w, r, fullPath)
		return
	}

	s.handleDirectory(w, dirName, basePath, subPath, fullPath)
}

func (s *Server) handleRoot(w http.ResponseWriter, _ *http.Request) {
	if len(s.directories) == 1 {
		for name, basePath := range s.directories {
			s.handleDirectory(w, name, basePath, "", basePath)
			return
		}
	}

	var files []FileInfo
	for name := range s.directories {
		files = append(files, FileInfo{
			Name:  name,
			IsDir: true,
			Path:  "/" + name,
		})
	}

	sort.Slice(files, func(i, j int) bool {
		return strings.ToLower(files[i].Name) < strings.ToLower(files[j].Name)
	})

	data := struct {
		Path     string
		Files    []FileInfo
		Parent   string
		HasBack  bool
		IsRoot   bool
		ShowPath bool
	}{
		Path:     "/",
		Files:    files,
		HasBack:  false,
		IsRoot:   true,
		ShowPath: true,
	}

	tmpl := template.Must(template.ParseFS(templatesFS, "templates/index.html"))
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tmpl.Execute(w, data)
}

func (s *Server) handleDirectory(w http.ResponseWriter, dirName, _, subPath, fullPath string) {
	entries, err := os.ReadDir(fullPath)
	if err != nil {
		http.Error(w, "Failed to read directory", http.StatusInternalServerError)
		return
	}

	var files []FileInfo
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}

		filePath := "/" + dirName
		if subPath != "" {
			filePath = filePath + "/" + subPath
		}
		filePath = filePath + "/" + entry.Name()

		files = append(files, FileInfo{
			Name:    entry.Name(),
			Size:    formatSize(info.Size()),
			ModTime: info.ModTime().Format("2006-01-02 15:04"),
			IsDir:   entry.IsDir(),
			IsText:  isTextFile(entry.Name()),
			Path:    filePath,
		})
	}

	sort.Slice(files, func(i, j int) bool {
		if files[i].IsDir != files[j].IsDir {
			return files[i].IsDir
		}
		return strings.ToLower(files[i].Name) < strings.ToLower(files[j].Name)
	})

	currentPath := "/" + dirName
	if subPath != "" {
		currentPath = currentPath + "/" + subPath
	}

	parent := filepath.Dir(currentPath)
	if parent == "/" {
		parent = "/"
	}

	hasBack := true
	showPath := true
	if len(s.directories) == 1 && subPath == "" {
		hasBack = false
		showPath = false
	}

	data := struct {
		Path     string
		Files    []FileInfo
		Parent   string
		HasBack  bool
		IsRoot   bool
		ShowPath bool
	}{
		Path:     currentPath,
		Files:    files,
		Parent:   parent,
		HasBack:  hasBack,
		IsRoot:   false,
		ShowPath: showPath,
	}

	tmpl := template.Must(template.ParseFS(templatesFS, "templates/index.html"))
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tmpl.Execute(w, data)
}

func isTextFile(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	switch ext {
	case ".txt", ".log", ".json", ".xml", ".ini", ".cfg", ".conf", ".md", ".sh":
		return true
	}
	return false
}

func (s *Server) handleView(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/view/")

	parts := strings.SplitN(path, "/", 2)
	if len(parts) < 2 {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	dirName := parts[0]
	subPath := parts[1]

	basePath, ok := s.directories[dirName]
	if !ok {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	fullPath := filepath.Join(basePath, subPath)

	if !strings.HasPrefix(filepath.Clean(fullPath), filepath.Clean(basePath)) {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	info, err := os.Stat(fullPath)
	if err != nil || info.IsDir() {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	content, err := os.ReadFile(fullPath)
	if err != nil {
		http.Error(w, "Failed to read file", http.StatusInternalServerError)
		return
	}

	parent := filepath.Dir("/" + path)

	data := struct {
		Name    string
		Path    string
		Parent  string
		Content string
	}{
		Name:    filepath.Base(fullPath),
		Path:    "/" + path,
		Parent:  parent,
		Content: string(content),
	}

	tmpl := template.Must(template.ParseFS(templatesFS, "templates/view.html"))
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tmpl.Execute(w, data)
}

func (s *Server) handleDownload(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/download/")

	parts := strings.SplitN(path, "/", 2)
	if len(parts) < 2 {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	dirName := parts[0]
	subPath := parts[1]

	basePath, ok := s.directories[dirName]
	if !ok {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	fullPath := filepath.Join(basePath, subPath)

	if !strings.HasPrefix(filepath.Clean(fullPath), filepath.Clean(basePath)) {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	info, err := os.Stat(fullPath)
	if err != nil || info.IsDir() {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filepath.Base(fullPath)))
	http.ServeFile(w, r, fullPath)
}

func (s *Server) handleDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/delete/")

	parts := strings.SplitN(path, "/", 2)
	if len(parts) < 2 {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	dirName := parts[0]
	subPath := parts[1]

	basePath, ok := s.directories[dirName]
	if !ok {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	fullPath := filepath.Join(basePath, subPath)

	if !strings.HasPrefix(filepath.Clean(fullPath), filepath.Clean(basePath)) {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	info, err := os.Stat(fullPath)
	if err != nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	if info.IsDir() {
		err = os.RemoveAll(fullPath)
	} else {
		err = os.Remove(fullPath)
	}

	if err != nil {
		http.Error(w, "Failed to delete", http.StatusInternalServerError)
		return
	}

	parent := "/" + dirName
	if idx := strings.LastIndex(subPath, "/"); idx > 0 {
		parent = "/" + dirName + "/" + subPath[:idx]
	}

	http.Redirect(w, r, parent+"?delete=1", http.StatusSeeOther)
}

func formatSize(size int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case size >= GB:
		return fmt.Sprintf("%.1f GB", float64(size)/GB)
	case size >= MB:
		return fmt.Sprintf("%.1f MB", float64(size)/MB)
	case size >= KB:
		return fmt.Sprintf("%.1f KB", float64(size)/KB)
	default:
		return fmt.Sprintf("%d B", size)
	}
}
