package model

type Document struct {
	ID   string
	Body string
}

type File struct {
	Path string
	Body string
}

type Skill struct {
	ID    string
	Files []File
}

type Output struct {
	Path string
	Body string
}

type Section struct {
	Path string
	ID   string
	Body string
}

type Desired struct {
	Files    []Output
	Sections []Section
}

type Resolved struct {
	Sections []Document
	Commands []Document
	Agents   []Document
	Skills   []Skill
}
