package gorelease

import (
	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/storage"
	"context"
	"fmt"
	"github.com/pkg/errors"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// local path: url
type GCSResult map[string]string

// GCSRelease is Google Cloud Storage ReleaseFunc
func GCSRelease(bucket string, result GCSResult) ReleaseFunc {
	// create default context
	ctx := context.Background()

	// get api credentials
	creds, err := google.FindDefaultCredentials(ctx, secretmanager.DefaultAuthScopes()...)
	if err != nil {
		log.Fatal(errors.Wrap(err, "while finding default credentials"))
	}

	// create new client
	client, err := storage.NewClient(ctx, option.WithCredentials(creds))
	if err != nil {
		log.Fatal(errors.Wrap(err, "while creating new client"))
	}

	// get bucket
	bck := client.Bucket(bucket)

	return func(r *Release) error {
		return r.ForEachTargetBuild(func(target *Target, build *FileBuild) error {

			// handles error
			var handle = func(err error) {
				log.Fatal(errors.Wrapf(err, "while releasing %s", build.BinPath))
			}

			// open release file
			f, err := os.Open(build.BinPath)
			if err != nil {
				handle(err)
			}

			// read file
			b, err := ioutil.ReadAll(f)
			if err != nil {
				handle(err)
			}

			// close file
			if err = f.Close(); err != nil {
				handle(err)
			}

			// create new object
			obj := bck.Object(build.BinPath)

			// todo check if exists
			// write file to object
			w := obj.NewWriter(ctx)
			log.Printf("writing to gcs object %s", build.BinPath)
			if _, err = w.Write(b); err != nil {
				handle(err)
			}

			// close writer
			err = w.Close()
			if err != nil {
				handle(err)
			}
			_, err = obj.Attrs(ctx)
			if err != nil {
				handle(err)
			}
			result[build.BinPath] = makeObjectURL(bucket, *build)
			return nil
		})
	}
}

// convertMediaLink converts object MediaLink to clean url
func makeObjectURL(bucket string, f FileBuild) string {
	p := filepath.ToSlash(f.BinPath)
	p = strings.TrimPrefix(p, "/")
	return fmt.Sprintf("https://storage.googleapis.com/%s/%s", bucket, p)
}
