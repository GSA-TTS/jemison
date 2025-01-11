//nolint:gosec
package main

import (
	"bufio"
	"crypto/sha1"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"strconv"
	"strings"

	common "github.com/GSA-TTS/jemison/internal/common"
	kv "github.com/GSA-TTS/jemison/internal/kv"
	"github.com/GSA-TTS/jemison/internal/util"
	"github.com/google/uuid"
	"github.com/pingcap/log"
	"github.com/riverqueue/river"
	"go.uber.org/zap"
)

// This could become a constant in the config.
// But, it is not likely something we want to change.
const CHUNKSIZE = 4 * 1024

func hostAndPath(job *river.Job[common.FetchArgs]) string {
	var u url.URL
	u.Scheme = job.Args.Scheme
	u.Host = job.Args.Host
	u.Path = job.Args.Path

	return u.String()
}

func chunkwiseSHA1(filename string) []byte {
	// Open the file for reading.
	tFile, err := os.Open(filename)
	if err != nil {
		zap.L().Error("could not open temp file for encoding to B64")
	}
	defer tFile.Close()
	// Compute the SHA1 going chunk-by-chunk
	h := sha1.New()
	reader := bufio.NewReader(tFile)

	chunkSize := CHUNKSIZE
	bytesRead := 0
	buf := make([]byte, chunkSize)

	for {
		n, err := reader.Read(buf)
		bytesRead += n

		if err != nil {
			if !errors.Is(err, io.EOF) {
				zap.L().Error("chunk error reading")
			}

			break
		}

		chunk := buf[0:n]

		// https://pkg.go.dev/crypto/sha1#example-New
		_, err = io.Writer.Write(h, chunk)
		if err != nil {
			zap.L().Error("did not write SHA bytes successfully",
				zap.Int("h.Size", h.Size()), zap.String("chunk", string(chunk)))
		}
	}

	return h.Sum(nil)
}

func getURLToFile(u url.URL) (string, int64, []byte, error) {
	getResponse, err := RetryClient.Get(u.String())
	if err != nil {
		zap.L().Error("cannot GET content",
			zap.String("url", u.String()),
		)

		//nolint:wrapcheck
		return "", 0, nil, err
	}

	zap.L().Debug("successful GET response")

	// Create a temporary file to download the HTML to.
	temporaryFilename := uuid.NewString()

	outFile, err := os.Create(temporaryFilename)
	if err != nil {
		zap.L().Error("cannot create temporary file", zap.String("filename", temporaryFilename))

		//nolint:wrapcheck
		return "", 0, nil, err
	}

	defer outFile.Close()

	// Copy the Get Reader to a file Writer
	// Should consume little/no RAM.
	// Destination, Source
	bytesRead, err := io.Copy(outFile, getResponse.Body)
	if err != nil {
		zap.L().Error("could not copy GET to file",
			zap.String("url", u.String()),
			zap.String("filename", temporaryFilename))

		//nolint:wrapcheck
		return "", 0, nil, err
	}

	getResponse.Body.Close()

	// Now, it is in a file.
	// Compute the SHA1
	theSHA := chunkwiseSHA1(temporaryFilename)

	return temporaryFilename, bytesRead, theSHA, nil
}

const TooShort = 20

//nolint:cyclop,funlen
func fetchPageContent(job *river.Job[common.FetchArgs]) (
	map[string]string,
	error,
) {
	u := url.URL{
		Scheme: job.Args.Scheme,
		Host:   job.Args.Host,
		Path:   job.Args.Path,
	}

	headResp, err := RetryClient.Head(u.String())
	if err != nil {
		//nolint:wrapcheck
		return nil, err
	}

	defer headResp.Body.Close()

	// Get a clean mime type right away
	contentType := util.CleanMimeType(headResp.Header.Get("content-type"))
	log.Debug("checking HEAD MIME type", zap.String("content-type", contentType))

	if !util.IsSearchableMimeType(contentType) {
		return nil, fmt.Errorf(
			common.NonIndexableContentType.String()+
				" non-indexable MIME type: %s", u.String())
	}

	// Make sure we don't fetch things that are too big.
	sizeString := headResp.Header.Get("content-length")

	size, err := strconv.Atoi(sizeString)
	if err == nil {
		if int64(size) > MaxFilesize {
			return nil, fmt.Errorf(
				common.FileTooLargeToFetch.String()+
					" file too large to fetch: %s%s", job.Args.Host, job.Args.Path)
		}
	}

	// Write the raw content to a file.
	tempFilename, bytesRead, theSHA, err := getURLToFile(u)
	if err != nil {
		return nil, err
	}

	key := util.CreateS3Key(util.ToScheme(job.Args.Scheme), job.Args.Host, job.Args.Path, util.Raw)

	if bytesRead > MaxFilesize {
		zap.L().Warn("file too large",
			zap.String("host", job.Args.Host),
			zap.String("path", job.Args.Path))

		err := os.Remove(tempFilename)
		if err != nil {
			zap.L().Error("could not delete temp file that is too big...")
		}

		return nil, fmt.Errorf(
			common.FileTooLargeToFetch.String()+
				" file is too large: %d %s%s", bytesRead, job.Args.Host, job.Args.Path)
	}

	// Don't bother in case it came in at zero length
	if bytesRead < TooShort {
		return nil, fmt.Errorf(
			common.FileTooSmallToProcess.String()+
				" file is too small: %d %s%s", bytesRead, job.Args.Host, job.Args.Path)
	}

	defer func(u url.URL, key *util.Key) {
		err := os.Remove(tempFilename)
		if err != nil {
			zap.L().Error("could not remove temp file",
				zap.String("url", u.String()),
				zap.String("key", key.Render()))
		}
	}(u, key)

	// Stream that file over to S3
	s3 := kv.NewS3(ThisServiceName)

	err = s3.FileToS3(key, tempFilename, util.GetMimeType(contentType))
	if err != nil {
		zap.L().Error("could not send file to S3",
			zap.String("key", key.Render()),
			zap.String("tempFilename", tempFilename))
	}

	response := make(map[string]string)
	// Copy in all of the response headers.
	// Doing this first, so we can overwrite some things.
	for k := range headResp.Header {
		response[strings.ToLower(k)] = headResp.Header.Get(k)
	}

	for k, v := range map[string]string{
		"raw":            key.Render(),
		"sha1":           fmt.Sprintf("%x", theSHA),
		"content-length": fmt.Sprintf("%d", bytesRead),
		"scheme":         job.Args.Scheme,
		"host":           job.Args.Host,
		"path":           job.Args.Path,
	} {
		response[k] = v
	}

	response["content-type"] = contentType

	zap.L().Debug("content read",
		zap.String("content-length", response["content-length"]),
	)

	return response, nil
}
