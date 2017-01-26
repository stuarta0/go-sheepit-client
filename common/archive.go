package common

type FileArchive interface {
	// MD5 hash that the archive should match
	GetExpectedHash() string

    // Usually RootPath/<hash>.zip
    GetArchivePath() string

    // Usually RootPath/<hash/*.*
    GetContentPath() string
}