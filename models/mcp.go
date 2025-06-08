package models

type MCPConfig struct {
	Name        string     `json:"name"`
	Version     string     `json:"version"`
	Description string     `json:"description"`
	Author      string     `json:"author"`
	License     string     `json:"license"`
	Keywords    []string   `json:"keywords"`
	Repository  Repository `json:"repository"`
	Run         RunConfig  `json:"run"`
}

type Repository struct {
	Type string `json:"type"`
	URL  string `json:"url"`
}

type RunConfig struct {
	Command string   `json:"command"`
	Args    []string `json:"args"`
	Port    int      `json:"port"`
}

type DockerfileRequest struct {
	ZipFile []byte `json:"zip_file"`
}

type DockerfileResponse struct {
	ExtractedPath  string    `json:"extracted_path"`
	DockerfilePath string    `json:"dockerfile_path"`
	ImageName      string    `json:"image_name"`
	TarFilePath    string    `json:"tar_file_path"`
	Config         MCPConfig `json:"config"`
	Success        bool      `json:"success"`
	Message        string    `json:"message,omitempty"`
}
