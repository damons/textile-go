package schema

import (
	"errors"

	ipld "gx/ipfs/QmR7TcHkR9nxkUorfi8XMTAMLUK7GiP64TWWBzY3aacc1o/go-ipld-format"
)

// ErrFileValidationFailed indicates dag schema validation failed
var ErrFileValidationFailed = errors.New("file failed schema validation")

// ErrEmptySchema indicates a schema is empty
var ErrEmptySchema = errors.New("schema is does not create any files")

// ErrLinkOrderNotSolvable
var ErrLinkOrderNotSolvable = errors.New("link order is not solvable")

// ErrSchemaInvalidMill indicates a schema has an invalid mill entry
var ErrSchemaInvalidMill = errors.New("schema contains an invalid mill")

// ErrMissingJsonSchema indicates json schema is missing
var ErrMissingJsonSchema = errors.New("json mill requires a json schema")

// ErrBadJsonSchema indicates json schema is invalid
var ErrBadJsonSchema = errors.New("json schema is not valid")

// FileTag indicates the link should "use" the input file as source
const FileTag = ":file"

// SingleFileTag is a magic key indicating that a directory is actually a single file
const SingleFileTag = ":single"

// Node describes a DAG node
type Node struct {
	Name       string                 `json:"name,omitempty"`
	Pin        bool                   `json:"pin"`
	Plaintext  bool                   `json:"plaintext"`
	Mill       string                 `json:"mill,omitempty"`
	Opts       map[string]string      `json:"opts,omitempty"`
	JsonSchema map[string]interface{} `json:"json_schema,omitempty"`
	Links      map[string]*Link       `json:"links,omitempty"`
}

// Link is a sub-node which can "use" input from other sub-nodes
type Link struct {
	Use        string                 `json:"use,omitempty"`
	Pin        bool                   `json:"pin"`
	Plaintext  bool                   `json:"plaintext"`
	Mill       string                 `json:"mill,omitempty"`
	Opts       map[string]string      `json:"opts,omitempty"`
	JsonSchema map[string]interface{} `json:"json_schema,omitempty"`
}

// Step is an ordered name-link pair
type Step struct {
	Name string
	Link *Link
}

// ValidateMill is false if mill is not one of the built in tags
func ValidateMill(mill string) bool {
	switch mill {
	case
		"/schema",
		"/blob",
		"/image/resize",
		"/image/exif",
		"/json":
		return true
	}
	return false
}

// LinkByName find a link w/ the given name in the provided list
func LinkByName(links []*ipld.Link, name string) *ipld.Link {
	for _, l := range links {
		if l.Name == name {
			return l
		}
	}
	return nil
}

// Steps returns link steps in the order they should be processed
func Steps(links map[string]*Link) ([]Step, error) {
	var steps []Step
	run := links
	i := 0
	for {
		if i > len(links) {
			return nil, ErrLinkOrderNotSolvable
		}
		next := orderLinks(run, &steps)
		if len(next) == 0 {
			break
		}
		run = next
		i++
	}
	return steps, nil
}

// orderLinks attempts to place all links in steps, returning any unused
// whose source is not yet in steps
func orderLinks(links map[string]*Link, steps *[]Step) map[string]*Link {
	unused := make(map[string]*Link)
	for name, link := range links {
		if link.Use == FileTag {
			*steps = append([]Step{{Name: name, Link: link}}, *steps...)
		} else {
			useAt := -1
			for i, s := range *steps {
				if link.Use == s.Name {
					useAt = i
					break
				}
			}
			if useAt >= 0 {
				*steps = append(*steps, Step{Name: name, Link: link})
			} else {
				unused[name] = link
			}
		}
	}
	return unused
}
