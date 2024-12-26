package main

import (
	"html/template"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

type ConfigDoc struct {
	Host string
}

type ConfigServer struct {
	Url string
}

type Package struct {
	Repo        string
	Description string
}

type Config struct {
	Godoc   ConfigDoc
	Server  ConfigServer
	Package map[string]Package
}

type VanityData struct {
	GoImportContent  string
	GoRefreshContent string
	RepoLink         string
}

func main() {
	fileData, err := os.ReadFile("./vanity.toml")
	if err != nil {
		log.Fatalf("Unable to read vanity.toml with error %v", err)
	}

	var cfg Config
	err = toml.Unmarshal(fileData, &cfg)
	if err != nil {
		log.Fatalf("Failed to read config with error %v", err)
	}

	templ := template.Must(
		template.ParseFiles("./template.html"))

	os.MkdirAll("./dist", os.ModePerm)

	for pkgName, packageDetails := range cfg.Package {
		fileName := filepath.Join("./dist", pkgName+".html")
		fd, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE, os.ModePerm)
		if err != nil {
			log.Fatalf("Failed to write file %v with error %v", fileName, err)
		}
		defer fd.Close()
		repoUrlwithProto := urlWithProto(packageDetails.Repo)
		templ.Execute(
			fd, VanityData{
				GoImportContent:  constructImportContent(cfg, pkgName, repoUrlwithProto),
				GoRefreshContent: constructRefreshContent(cfg, pkgName),
				RepoLink:         repoUrlwithProto,
			},
		)
	}
}

func constructRefreshContent(cfg Config, pkgName string) string {
	var sb strings.Builder
	sb.WriteString("0;URL='")
	path, err := url.JoinPath(cfg.Godoc.Host, cfg.Server.Url, pkgName)
	if err != nil {
		log.Printf("Failed to create refreshContent url for %v", pkgName)
		return ""
	}
	usableUrl, err := url.Parse(path)
	if err != nil {
		log.Printf("Failed to create refreshContent url for %v", pkgName)
		return ""
	}
	usableUrl.Scheme = "https"
	sb.WriteString(usableUrl.String())
	return sb.String()
}

func constructImportContent(cfg Config, pkgName string, repoUrl string) string {
	var sb strings.Builder
	urlPath, _ := url.JoinPath(cfg.Server.Url, pkgName)
	sb.WriteString(urlPath)
	sb.WriteString(" git ")
	sb.WriteString(repoUrl)
	return sb.String()
}

func urlWithProto(urlString string) string {
	parsed, _ := url.Parse(urlString)
	parsed.Scheme = "http"
	return parsed.String()
}
