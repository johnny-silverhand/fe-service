package model

import (
	"encoding/json"
)

type ProductMetadata struct {


	// Files holds information about the file attachments on the post.
	Files []*FileInfo `json:"files,omitempty"`

	// Images holds the dimensions of all external images in the post as a map of the image URL to its diemsnions.
	// This includes image embeds (when the message contains a plaintext link to an image), Markdown images, images
	// contained in the OpenGraph metadata, and images contained in message attachments. It does not contain
	// the dimensions of any file attachments as those are stored in FileInfos.
	Images map[string]*PostImage `json:"images,omitempty"`

}

type ProductImage struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

func (o *ProductImage) ToJson() string {
	b, _ := json.Marshal(o)
	return string(b)
}
