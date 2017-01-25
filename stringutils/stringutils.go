package stringutils

import (
    "strings"
)

func IsEmpty(s string) bool {
    return strings.TrimSpace(s) == ""
}

func IsEmptyPtr(s *string) bool {
    return s == nil || IsEmpty(*s)
}