package gorelease

import (
	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/storage"
	"context"
	"github.com/pkg/errors"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"io/ioutil"
	"log"
	"os"
)

// filename: url
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
				log.Fatal(errors.Wrapf(err, "while releasing %s", build))
			}

			// handle writer close
			var closeErr error
			defer func() {
				if closeErr != nil {
					handle(err)
				}
			}()

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
			err = f.Close()
			if closeErr != nil {
				handle(err)
			}

			// create new object
			obj := bck.Object(f.Name())

			// todo check if exists
			// write file to object
			w := obj.NewWriter(ctx)
			if _, err = w.Write(b); err != nil {
				handle(err)
			}

			// close writer
			closeErr = w.Close()

			attrs, err := obj.Attrs(ctx)
			if err != nil {
				handle(err)
			}
			result[build.Name] = attrs.MediaLink
			return nil
		})
	}
}
