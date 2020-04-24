package attributes

import (
	"github.com/gophercloud/gophercloud/pagination"
)

// AttributePage is a single page of attributes results.
type AttributePage struct {
	pagination.SinglePageBase
}

// IsEmpty determines whether or not a page of attributes contains any results.
func (r AttributePage) IsEmpty() (bool, error) {
	attributes, err := ExtractAttributes(r)
	return len(attributes) == 0, err
}

// NextPageURL extracts the "next" link from the links section of the result.
func (r AttributePage) NextPageURL() (string, error) {
	return "", nil
}

// ExtractAttributes accepts a Page struct, specifically an AttributePage struct,
// and extracts the elements into a slice of Attribute structs. In other words,
// a generic collection is mapped into a relevant slice.
func ExtractAttributes(r pagination.Page) ([]string, error) {
	var s []string
	err := ExtractAttributesInto(r, &s)
	return s, err
}

func ExtractAttributesInto(r pagination.Page, v interface{}) error {
	return r.(AttributePage).Result.ExtractIntoSlicePtr(v, "")
}
