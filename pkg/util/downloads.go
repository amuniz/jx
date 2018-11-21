package util

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"gopkg.in/AlecAivazis/survey.v1"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/blang/semver"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"

	"github.com/jenkins-x/jx/pkg/jx/cmd"
)

var githubClient *github.Client

// Download a file from the given URL
func DownloadFile(filepath string, url string) (err error) {
	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := GetClientWithTimeout(time.Duration(time.Hour * 2)).Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("download of %s failed with return code %d", url, resp.StatusCode)
		return err
	}

	// Writer the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	// make it executable
	os.Chmod(filepath, 0755)
	if err != nil {
		return err
	}
	return nil
}

func GetLatestVersionFromGitHub(githubOwner, githubRepo string) (semver.Version, error) {
	text, err := GetLatestVersionStringFromGitHub(githubOwner, githubRepo)
	if err != nil {
		return semver.Version{}, err
	}
	if text == "" {
		return semver.Version{}, fmt.Errorf("No version found")
	}
	return semver.Make(text)
}

func GetLatestVersionStringFromGitHub(githubOwner, githubRepo string) (string, error) {
	latestVersionString, err := GetLatestTagFromGitHub(githubOwner, githubRepo)
	if err != nil {
		return "", err
	}
	if latestVersionString != "" {
		return strings.TrimPrefix(latestVersionString, "v"), nil
	}
	return "", fmt.Errorf("Unable to find the latest version for github.com/%s/%s", githubOwner, githubRepo)
}

func GetLatestTagFromGitHub(githubOwner, githubRepo string, options cmd.CreateOptions) (string, error) {
	token := os.Getenv("GH_TOKEN")
	if githubClient == nil {
		var tc *http.Client
		if len(token) > 0 {
			ts := oauth2.StaticTokenSource(
				&oauth2.Token{AccessToken: token},
			)
			tc = oauth2.NewClient(oauth2.NoContext, ts)
		}
		githubClient = github.NewClient(tc)
	}
	client := githubClient
	var (
		release *github.RepositoryRelease
		resp    *github.Response
		err     error
	)

	if len(token) < 0 {
		// anonymous access, check rate limit
		limits, response, e := client.RateLimits(context.Background())
		if e != nil {
			return "", e
		}
		defer response.Body.Close()
		// 5 (out of 60) seems a reasonable threshold to start asking for a token
		if limits.GetCore().Remaining < 5 {
			if options != nil {
				surveyOpts := survey.WithStdio(options.In, options.Out, options.Err)
				prompts := &survey.Input{
					Message: fmt.Sprintf("GitHub API Token:"),
				}
				var token string
				err := survey.AskOne(prompts, &token, nil, surveyOpts)
				if err != nil {
					return "", err
				}
			} else {

			}
		}

	}

	release, resp, err = client.Repositories.GetLatestRelease(context.Background(), githubOwner, githubRepo)
	if err != nil {
		return "", fmt.Errorf("Unable to get latest version for github.com/%s/%s %v", githubOwner, githubRepo, err)
	}
	defer resp.Body.Close()
	latestVersionString := release.TagName
	if latestVersionString != nil {
		return *latestVersionString, nil
	}
	return "", fmt.Errorf("Unable to find the latest version for github.com/%s/%s", githubOwner, githubRepo)
}

// untargz a tarball to a target, from
// http://blog.ralch.com/tutorial/golang-working-with-tar-and-gzipf
func UnTargz(tarball, target string, onlyFiles []string) error {
	zreader, err := os.Open(tarball)
	if err != nil {
		return err
	}
	defer zreader.Close()

	reader, err := gzip.NewReader(zreader)
	defer reader.Close()
	if err != nil {
		panic(err)
	}

	tarReader := tar.NewReader(reader)

	for {
		inkey := false
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		for _, value := range onlyFiles {
			if value == "*" || value == path.Base(header.Name) {
				inkey = true
				break
			}
		}

		if !inkey && len(onlyFiles) > 0 {
			continue
		}

		path := filepath.Join(target, path.Base(header.Name))
		info := header.FileInfo()
		if info.IsDir() {
			if err = os.MkdirAll(path, info.Mode()); err != nil {
				return err
			}
			continue
		}

		file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode())
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = io.Copy(file, tarReader)
		if err != nil {
			return err
		}
	}
	return nil
}

// untargz a tarball to a target including any folders inside the tarball
// http://blog.ralch.com/tutorial/golang-working-with-tar-and-gzipf
func UnTargzAll(tarball, target string) error {
	zreader, err := os.Open(tarball)
	if err != nil {
		return err
	}
	defer zreader.Close()

	reader, err := gzip.NewReader(zreader)
	defer reader.Close()
	if err != nil {
		panic(err)
	}

	tarReader := tar.NewReader(reader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		path := filepath.Join(target, header.Name)
		info := header.FileInfo()
		if info.IsDir() {
			if err = os.MkdirAll(path, info.Mode()); err != nil {
				return err
			}
			continue
		}

		file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode())
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = io.Copy(file, tarReader)
		if err != nil {
			return err
		}
	}
	return nil
}
