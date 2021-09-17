package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"strings"
)

type BarkConfig struct {
	Server string
	Token  string
	//
	Content string
	Title   string
	//
	AutoCopy bool
	Archive  bool
	//
	Sound string
	Group string
	Url   string
	Copy  string
}

func (b *BarkConfig) GetUrl() *url.URL {
	if b.Server == "" {
		b.Server = "https://api.day.app"
	}
	var path string
	if b.Token == "" {
		log.Fatalf("token is empty")
	}
	if b.Title != "" {
		path = fmt.Sprintf("%s/%s/%s", b.Token, b.Title, b.Content)
	} else {
		path = fmt.Sprintf("%s/%s", b.Token, b.Content)
	}
	parsedUrl, err := url.Parse(b.Server)
	if err != nil {
		log.Fatalf("parse url fail %+v", err)
	}
	if !strings.HasSuffix(parsedUrl.Path, "/") {
		parsedUrl.Path = parsedUrl.Path + "/"
	}
	parsedUrl.Path = parsedUrl.Path + path
	query := parsedUrl.Query()
	if b.Archive {
		query.Set("isArchive", "1")
	}
	if b.AutoCopy {
		query.Set("autoCopy", "1")
	}
	if b.Sound != "" {
		query.Set("sound", b.Sound)
	}
	if b.Group != "" {
		query.Set("group", b.Group)
	}
	if b.Url != "" {
		query.Set("url", b.Url)
	}
	if b.Copy != "" {
		query.Set("copy", b.Copy)
	}
	parsedUrl.RawQuery = query.Encode()
	return parsedUrl
}

func getConfigFromEnv(data interface{}) {
	dataType := reflect.TypeOf(data).Elem()
	v := reflect.ValueOf(data).Elem()
	for i := 0; i < dataType.NumField(); i++ {
		field := dataType.Field(i)
		envName := fmt.Sprintf("PLUGIN_%s", strings.ToUpper(field.Name))
		//fmt.Println(envName)
		env, found := os.LookupEnv(envName)
		if !found {
			continue
		}
		switch field.Type.String() {
		case "string":
			v.Field(i).Set(reflect.ValueOf(env))
		case "bool":
			parseBool, err := strconv.ParseBool(env)
			if err == nil {
				v.Field(i).Set(reflect.ValueOf(parseBool))
			}
		default:
			log.Printf("unknow type %s with env %v", field.Type.String(), env)
		}
	}
}

func main() {
	config := BarkConfig{}
	getConfigFromEnv(&config)
	resp, err := http.Get(config.GetUrl().String())
	if err != nil {
		log.Fatalf("request bark server error %+v", err)
	}
	if !(resp.StatusCode >= 200 && resp.StatusCode < 300) {
		log.Printf("bark response %+v", resp)
	} else {
		log.Println("request bark success")
	}
}
