package models

import "io"

// ImageUpload carries an in-flight file upload from the handler to the service layer,
// deliberately excluding mime/multipart or net/http types so the service layer stays transport-agnostic.
type ImageUpload struct {
	Reader      io.Reader
	Filename    string
	ContentType string
	Size        int64
}
