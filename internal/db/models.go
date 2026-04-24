package db

type Game struct {
	ID               string        `json:"id"`
	Platform         string        `json:"platform"`
	Category         string        `json:"category,omitempty"`
	Titles           []Title       `json:"titles,omitempty"`
	ROMs             []ROM         `json:"roms,omitempty"`
	Media            []Media       `json:"media,omitempty"`
	Descriptions     []Description `json:"descriptions,omitempty"`
	FirstReleaseDate string        `json:"first_release_date,omitempty"`
	ExternalIDs     map[string]any `json:"external_ids,omitempty"`
}

type Title struct {
	Text     string `json:"text"`
	Lang     string `json:"lang,omitempty"`
	Script   string `json:"script,omitempty"`
	Region   string `json:"region,omitempty"`
	Form     string `json:"form,omitempty"`
	Source   string `json:"source,omitempty"`
	Verified bool   `json:"verified,omitempty"`
}

type ROM struct {
	Name   string `json:"name,omitempty"`
	Region string `json:"region,omitempty"`
	Serial string `json:"serial,omitempty"`
	Size   int64  `json:"size,omitempty"`
	CRC32  string `json:"crc32,omitempty"`
	MD5    string `json:"md5,omitempty"`
	SHA1   string `json:"sha1,omitempty"`
	SHA256 string `json:"sha256,omitempty"`
	Source string `json:"source,omitempty"`
}

type Media struct {
	Kind     string `json:"kind"`
	URL      string `json:"url"`
	Region   string `json:"region,omitempty"`
	Source   string `json:"source,omitempty"`
	Verified bool   `json:"verified,omitempty"`
	Note     string `json:"note,omitempty"`
}

type Description struct {
	Lang   string `json:"lang"`
	Text   string `json:"text"`
	Source string `json:"source,omitempty"`
}
