package configuration

import (
	"context"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"cloud.google.com/go/storage"
)

var loc string

// LoadConfig TODO
func LoadConfig() *Config {
	loc = os.Getenv("GW_CONFIG_LOCATION")
	c := &(Config{})
	if strings.HasPrefix(loc, "gs://") {
		return c.Load(fromCloud)
	} else if strings.HasPrefix(loc, "./") {
		return c.Load(fromLocal)
	}
	return c.Load(fromCloud, fromLocal)
}

func fromCloud() []byte {
	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	if projectID == "" {
		return nil
	}

	gsExp := regexp.MustCompile(`gs://(?P<bucket>[^/]+)/(?P<object>.+)`)
	match := gsExp.FindStringSubmatch(loc)

	var bucket, object string

	for i, name := range gsExp.SubexpNames() {
		if i != 0 {
			if name == "bucket" {
				bucket = match[i]
				if object != "" {
					break
				}
			} else if name == "object" {
				object = match[i]
				if bucket != "" {
					break
				}
			}
		}
	}

	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatal(err)
	}

	data, err := readGCS(client, bucket, object)
	if err != nil {
		log.Fatalf("Cannot read object: %v", err)
	}

	return data
}

func readGCS(client *storage.Client, bucket, object string) ([]byte, error) {
	ctx := context.Background()
	// [START download_file]
	rc, err := client.Bucket(bucket).Object(object).NewReader(ctx)
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	data, err := ioutil.ReadAll(rc)
	if err != nil {
		return nil, err
	}
	return data, nil
	// [END download_file]
}

func fromLocal() []byte {
	folder, err := filepath.Abs(filepath.Dir("."))
	if err != nil {
		log.Println("loadBinaryFromLocal err", err)
		return nil
	}
	data, err := ioutil.ReadFile(folder + strings.TrimPrefix(loc, "."))
	if err != nil {
		log.Println("loadBinaryFromLocal err", err)
		return nil
	}
	return data
}
