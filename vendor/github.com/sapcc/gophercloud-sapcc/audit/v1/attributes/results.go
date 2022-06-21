// Copyright 2019 SAP SE
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
