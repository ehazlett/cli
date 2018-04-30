package liballoy

import (
	"archive/tar"
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/docker/docker/api/types"
)

func (a *Alloy) Run(version string) error {
	variant, ok := a.variants[version]
	if !ok {
		return fmt.Errorf("variant not found for version %s", version)
	}

	sum := sha256.Sum256([]byte(variant.Image))
	dest := filepath.Join(a.cacheDir, fmt.Sprintf("%x", sum))
	if err := os.MkdirAll(dest, 0755); err != nil {
		return err
	}

	entrypoint := filepath.Join(dest, variant.Entrypoint)

	// get variant if entrypoint does not exist
	if _, err := os.Stat(entrypoint); err != nil {
		if !os.IsNotExist(err) {
			return err
		}

		fmt.Printf("alloy: getting version %s\n", version)

		ctx := context.Background()

		if _, err := a.client.ImagePull(ctx, variant.Image, types.ImagePullOptions{}); err != nil {
			return err
		}

		r, err := a.client.ImageSave(ctx, []string{variant.Image})
		if err != nil {
			return err
		}

		// extract
		tr := tar.NewReader(r)
		for {
			h, err := tr.Next()
			if err == io.EOF {
				break
			}

			if err != nil {
				return err
			}

			if filepath.Base(h.Name) == "layer.tar" {
				tf, err := ioutil.TempFile("", "alloy-")
				if err != nil {
					return err
				}
				defer os.Remove(tf.Name())

				if _, err := io.Copy(tf, tr); err != nil {
					return err
				}
				tf.Close()

				if err := extract(tf.Name(), dest); err != nil {
					return err
				}
				break
			}
		}
	}

	cmd := exec.Command(entrypoint, os.Args[1:]...)
	cmd.Env = append(cmd.Env, "ALLOY_ENV=exec")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func extract(src, dest string) error {
	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer f.Close()
	tr := tar.NewReader(f)
	for {
		h, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		d := filepath.Join(dest, h.Name)
		if h.FileInfo().IsDir() {
			if err := os.MkdirAll(d, 0755); err != nil {
				return err
			}
			continue
		}

		f, err := os.Create(d)
		if err != nil {
			return err
		}

		if _, err := io.Copy(f, tr); err != nil {
			return err
		}
		f.Close()

		if err := os.Chmod(d, h.FileInfo().Mode()); err != nil {
			return err
		}
	}

	return nil
}
