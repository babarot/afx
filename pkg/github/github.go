package github

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"

	"github.com/b4b4r07/afx/pkg/errors"
	"github.com/mholt/archiver"
)

// Release represents a GitHub release and its client
// A difference from Release is whether a client or not
type Release struct {
	Client *http.Client

	Name   string
	Assets Assets

	// Filename is used for specifying a filename directly in release assets.
	// Normally it requires to filter release assets based on OS/Arch information.
	// But by doing this field, don't need to filter assets.
	Filename string
}

// Asset represents GitHub release's asset.
// Basically this means one archive file attached in a release
type Asset struct {
	Name string
	Home string
	Path string
	URL  string
}

type Assets []Asset

func (as *Assets) filter(fn func(Asset) bool) *Assets {
	var assets Assets
	if len(*as) < 2 {
		// no more need to filter
		log.Printf("[DEBUG] assets.filter: finished filtering because length of assets is less than two")
		return as
	}

	for _, asset := range *as {
		if fn(asset) {
			assets = append(assets, asset)
		}
	}

	// logging if assets are changed by filter
	if len(*as) != len(assets) {
		log.Printf("[DEBUG] assets.filter: filtered: %#v", getAssetKeys(assets))
	}

	*as = assets
	return as
}

// getAssetKeys just returns list of asset.Name
func getAssetKeys(assets []Asset) []string {
	var names []string
	for _, asset := range assets {
		names = append(names, asset.Name)
	}
	return names
}

func (r *Release) GetAsset() (Asset, error) {
	log.Printf("[DEBUG] assets: %#v\n", getAssetKeys(r.Assets))

	if len(r.Assets) == 0 {
		return Asset{}, fmt.Errorf("%s: no assets found", r.Name)
	}

	if r.Filename != "" {
		log.Printf("[DEBUG] asset: found filename %q is specified in config", r.Filename)
		for _, asset := range r.Assets {
			if asset.Name == r.Filename {
				log.Printf("[DEBUG] asset: filename %q is matched with assets", r.Filename)
				return asset, nil
			}
		}
		return Asset{}, fmt.Errorf("%s: no matched in assets", r.Filename)
	}

	assets := *r.Assets.
		filter(func(asset Asset) bool {
			expr := `.*\.sbom`
			// filter out
			return !regexp.MustCompile(expr).MatchString(asset.Name)
		}).
		filter(func(asset Asset) bool {
			expr := ".*(sha256sum|checksum).*"
			// filter out
			return !regexp.MustCompile(expr).MatchString(asset.Name)
		}).
		filter(func(asset Asset) bool {
			expr := ""
			switch runtime.GOOS {
			case "darwin":
				expr += ".*(apple|darwin|Darwin|osx|mac|macos|macOS).*"
			case "linux":
				expr += ".*(linux|hoe).*"
			}
			return regexp.MustCompile(expr).MatchString(asset.Name)
		}).
		filter(func(asset Asset) bool {
			expr := ""
			switch runtime.GOARCH {
			case "amd64":
				expr += ".*(amd64|64).*"
			case "386":
				expr += ".*(386|86).*"
			}
			return regexp.MustCompile(expr).MatchString(asset.Name)
		})

	switch len(assets) {
	case 0:
		return Asset{}, errors.New("asset not found after filtered")
	case 1:
		return assets[0], nil
	default:
		log.Printf("[WARN] %d assets found: %#v", len(assets), getAssetKeys(assets))
		log.Printf("[WARN] first one %q will be used", assets[0].Name)
		return assets[0], nil
	}
}

// Download is
func (r *Release) Download(ctx context.Context) (Asset, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	asset, err := r.GetAsset()
	if err != nil {
		return asset, err
	}

	log.Printf("[DEBUG] asset: %#v", asset)

	req, err := http.NewRequest(http.MethodGet, asset.URL, nil)
	if err != nil {
		return asset, err
	}

	client := new(http.Client)
	resp, err := client.Do(req.WithContext(ctx))
	if err != nil {
		return asset, err
	}
	defer resp.Body.Close()

	os.MkdirAll(asset.Home, os.ModePerm)
	file, err := os.Create(asset.Path)
	if err != nil {
		return asset, errors.Wrapf(err, "%s: failed to create file", asset.Path)
	}
	defer file.Close()
	_, err = io.Copy(file, resp.Body)

	return asset, err
}

// Unarchive is
func (r *Release) Unarchive(asset Asset) error {
	uaIface, err := archiver.ByExtension(asset.Path)
	if err != nil {
		// err: this will be an error of format unrecognized by filename
		// but in this case, maybe not archived file: e.g. tigrawap/slit
		//
		log.Printf("[WARN] archiver.ByExtension(): %v", err)

		// TODO: remove this logic?
		// thanks to this logic, we don't need to specify this statement to link.from
		//
		//   command:
		//     link:
		//     - from: '*jq*'
		//
		// because this logic renames a binary of 'jq-1.6' to 'jq'
		//
		target := filepath.Join(asset.Home, r.Name)
		if _, err := os.Stat(target); err != nil {
			log.Printf("[DEBUG] renamed from %s to %s", asset.Path, target)
			os.Rename(asset.Path, target)
			os.Chmod(target, 0755)
		}
		return nil
	}

	tar := &archiver.Tar{
		OverwriteExisting:      true,
		MkdirAll:               false,
		ImplicitTopLevelFolder: false,
		ContinueOnError:        false,
	}
	switch v := uaIface.(type) {
	case *archiver.Rar:
		v.OverwriteExisting = true
	case *archiver.Zip:
		v.OverwriteExisting = true
	case *archiver.TarBz2:
		v.Tar = tar
	case *archiver.TarGz:
		v.Tar = tar
	case *archiver.TarLz4:
		v.Tar = tar
	case *archiver.TarSz:
		v.Tar = tar
	case *archiver.TarXz:
		v.Tar = tar
	case *archiver.Gz,
		*archiver.Bz2,
		*archiver.Lz4,
		*archiver.Snappy,
		*archiver.Xz:
		// nothing to customise
	}

	u, ok := uaIface.(archiver.Unarchiver)
	if !ok {
		return errors.New("cannot type assertion with archiver.Unarchiver")
	}

	if err := u.Unarchive(asset.Path, asset.Home); err != nil {
		return errors.Wrapf(err, "%s: failed to unarchive", r.Name)
	}

	log.Printf("[DEBUG] removed archive file: %s", asset.Path)
	os.Remove(asset.Path)

	return nil
}
