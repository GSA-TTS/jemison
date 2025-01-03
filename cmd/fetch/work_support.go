package main

import (
	"bufio"
	"crypto/sha1"
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

func host_and_path(job *river.Job[common.FetchArgs]) string {
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
	// FIXME: make this a param in the config.
	chunkSize := 4 * 1024
	bytesRead := 0
	buf := make([]byte, chunkSize)
	for {
		n, err := reader.Read(buf)
		bytesRead += n

		if err != nil {
			if err != io.EOF {
				zap.L().Error("chunk error reading")
			}
			break
		}
		chunk := buf[0:n]
		// https://pkg.go.dev/crypto/sha1#example-New
		io.Writer.Write(h, chunk)
	}

	return h.Sum(nil)
}

func getUrlToFile(u url.URL) (string, int64, []byte, error) {
	getResponse, err := RetryClient.Get(u.String())
	if err != nil {
		zap.L().Error("cannot GET content",
			zap.String("url", u.String()),
		)
		return "", 0, nil, err
	}
	zap.L().Debug("successful GET response")
	// Create a temporary file to download the HTML to.
	temporaryFilename := uuid.NewString()
	outFile, err := os.Create(temporaryFilename)
	if err != nil {
		zap.L().Error("cannot create temporary file", zap.String("filename", temporaryFilename))
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
		return "", 0, nil, err
	}
	getResponse.Body.Close()
	// Now, it is in a file.
	// Compute the SHA1
	theSHA := chunkwiseSHA1(temporaryFilename)
	return temporaryFilename, bytesRead, theSHA, nil
}

func fetch_page_content(job *river.Job[common.FetchArgs]) (map[string]string, error) {
	u := url.URL{
		Scheme: job.Args.Scheme,
		Host:   job.Args.Host,
		Path:   job.Args.Path,
	}

	headResp, err := RetryClient.Head(u.String())
	if err != nil {
		return nil, err
	}

	// Get a clean mime type right away
	contentType := util.CleanMimeType(headResp.Header.Get("content-type"))
	log.Debug("checking HEAD MIME type", zap.String("content-type", contentType))
	if !util.IsSearchableMimeType(contentType) {
		return nil, fmt.Errorf(
			common.NonIndexableContentType.String()+
				" non-indexable MIME type: %s", u.String())
	}

	// Make sure we don't fetch things that are too big.
	size_string := headResp.Header.Get("content-length")
	size, err := strconv.Atoi(size_string)
	if err != nil {
		// Could not extract a size header...
	} else {
		// FIXME: Make this a constant
		if int64(size) > MaxFilesize {
			return nil, fmt.Errorf(
				common.FileTooLargeToFetch.String()+
					" file too large to fetch: %s%s", job.Args.Host, job.Args.Path)
		}
	}

	// Write the raw content to a file.
	tempFilename, bytesRead, theSHA, err := getUrlToFile(u)
	if err != nil {
		return nil, err
	}
	key := util.CreateS3Key(util.ToScheme(job.Args.Scheme), job.Args.Host, job.Args.Path, util.Raw)

	if bytesRead > MaxFilesize {
		zap.L().Warn("file too large",
			zap.String("host", job.Args.Host), zap.String("path", job.Args.Path))
		err := os.Remove(tempFilename)
		if err != nil {
			zap.L().Error("could not delete temp file that is too big...")
		}
		return nil, fmt.Errorf(
			common.FileTooLargeToFetch.String()+
				" file is too large: %d %s%s", bytesRead, job.Args.Host, job.Args.Path)
	}

	// Don't bother in case it came in at zero length
	if bytesRead < 100 {
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
	s3.FileToS3(key, tempFilename, util.GetMimeType(contentType))

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

	// FIXME
	// There is a texinfo standard library for normalizing content types.
	// Consider using it. I want a simplified string, not utf-8 etc.
	response["content-type"] = contentType

	zap.L().Debug("content read",
		zap.String("content-length", response["content-length"]),
	)

	return response, nil
}
