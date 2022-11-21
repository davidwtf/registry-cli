package action

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"registry-cli/pkg/client"
	"registry-cli/pkg/option"
	"strings"

	"github.com/opencontainers/go-digest"
)

func Blob(blobID string, opts *option.Options) error {
	cli, err := client.NewClient(opts)
	if err != nil {
		opts.WriteDebug("init client", err)
		return err
	}
	repo, err := cli.NewRepository(opts.Repositiory, client.PullAction)
	if err != nil {
		opts.WriteDebug("init repository service", err)
		return err
	}

	if !strings.HasPrefix(blobID, "sha256:") {
		blobID = "sha256:" + blobID
	}
	dgst := digest.Digest(blobID)
	if err := dgst.Validate(); err != nil {
		opts.WriteDebug(fmt.Sprintf(`validate blob digest "%s"`, dgst.String()), err)
		return err
	}

	reader, err := repo.Blobs(opts.Ctx).Open(opts.Ctx, dgst)
	if err != nil {
		opts.WriteDebug("open blob", err)
		return err
	}
	defer reader.Close()

	dst := opts.Destination
	dst, err = filepath.Abs(dst)
	if err != nil {
		opts.WriteDebug(fmt.Sprintf(`get absolute path for "%s"`, dst), err)
		return err
	}
	if _, err := os.Stat(dst); err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(dst, os.FileMode(0755)); err != nil {
				opts.WriteDebug(fmt.Sprintf(`make destionation directory "%s"`, dst), err)
				return err
			}
		} else {
			opts.WriteDebug(fmt.Sprintf(`check destionation directory "%s"`, dst), err)
			return err
		}
	}

	fn := filepath.Join(dst, blobID)
	writer, err := os.OpenFile(fn, os.O_CREATE|os.O_WRONLY, os.FileMode(0640))
	if err != nil {
		opts.WriteDebug(fmt.Sprintf(`create "%s"`, fn), err)
		return err
	}

	defer writer.Close()

	n, err := io.Copy(writer, reader)
	if err != nil {
		opts.WriteDebug(fmt.Sprintf(`copy blob "%s" to file "%s"`, blobID, fn), err)
		return err
	}
	opts.WriteDebug(fmt.Sprintf(`write "%s" %d bytes`, fn, n), nil)
	return nil
}
