package main

import (
    "fmt"
    "time"
    "unicode"
)

func generateCallID() string {
    return fmt.Sprintf("call_%s", time.Now().UTC().Format("20060102T150405Z"))
}

func sanitize(s string) string {
    out := make([]rune, 0, len(s))
    for _, r := range s {
        if unicode.IsLetter(r) || unicode.IsDigit(r) {
            out = append(out, r)
        } else {
            out = append(out, '_')
        }
    }
    return string(out)
}

func timeNowDate() string {
    return time.Now().Format("2006-01-02")
}
