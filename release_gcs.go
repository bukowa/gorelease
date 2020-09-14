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

// GCSRelease is Google Cloud Storage ReleaseFunc
func GCSRelease(bucket string) ReleaseFunc {
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
			object := bck.Object(f.Name())

			// write file to object
			writer := object.NewWriter(ctx)
			if _, err = writer.Write(b); err != nil {
				handle(err)
			}

			closeErr = writer.Close()
			return nil
		})
	}
}
