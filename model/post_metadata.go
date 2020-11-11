package model

import (
	"encoding/json"
)

type PostMetadata struct {
	// Embeds holds information required to render content embedded in the post. This includes the OpenGraph metadata
	// for links in the post.
	Embeds []*PostEmbed `json:"embeds,omitempty"`

	// Order holds information required to render content order in the post.
	Order *Order `json:"order,omitempty"`

	// Files holds information about the file attachments on the post.
	Files []*FileInfo `json:"files,omitempty"`

	// Images holds the dimensions of all external images in the post as a map of the image URL to its diemsnions.
	// This includes image embeds (when the message contains a plaintext link to an image), Markdown images, images
	// contained in the OpenGraph metadata, and images contained in message attachments. It does not contain
	// the dimensions of any file attachments as those are stored in FileInfos.
	Images map[string]*PostImage `json:"images,omitempty"`
}

type PostImage struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

func (o *PostImage) ToJson() string {
	b, _ := json.Marshal(o)
	return string(b)
}
