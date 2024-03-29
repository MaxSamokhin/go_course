package main

import (
    "io"
    "fmt"
    "os"
    "bufio"
    "encoding/json"
    "github.com/mailru/easyjson"
    "github.com/mailru/easyjson/jlexer"
    "github.com/mailru/easyjson/jwriter"
    "strings"
)

type User struct {
    Browsers []string `json:"browsers"`
    Email    string   `json:"email"`
    Name     string   `json:"name"`
}

func getCountUniqueBrowser(file io.Reader, out io.Writer) (int, error) {
    scanner := bufio.NewScanner(file)
    uniqueBrowsers := make(map[string]bool, 150)
    user := new(User)
    fmt.Fprintln(out, "found users:")
    idx := 0
    isAndroid, isMSIE := false, false

    for scanner.Scan() {
        line := scanner.Bytes()

        err := user.UnmarshalJSON(line)
        if err != nil {
            return -1, fmt.Errorf("[ERROR]: Error unmarshal JSON")
        }
        isAndroid, isMSIE = false, false

        for _, browserRaw := range user.Browsers {
            if strings.Contains(browserRaw, "MSIE") {
                isMSIE = true
                uniqueBrowsers[browserRaw] = true
            }

            if strings.Contains(browserRaw, "Android") {
                isAndroid = true
                uniqueBrowsers[browserRaw] = true
            }
        }

        if isAndroid && isMSIE {
            fmt.Fprintf(out, "[%d] %s <%s>\n", idx, user.Name,
                strings.Replace(user.Email, "@", " [at] ", 1))
        }

        idx++
    }

    if err := scanner.Err(); err != nil {
        return -1, fmt.Errorf("[ERROR]: Error scanner")
    }

    return len(uniqueBrowsers), nil
}

func FastSearch(out io.Writer) {
    file, err := os.Open(filePath)
    if err != nil {
        panic(err)
    }

    defer file.Close()

    countUniqueBrowser, err := getCountUniqueBrowser(file, out)
    if err != nil {
        panic(err)
    }

    fmt.Fprintln(out, "\nTotal unique browsers", countUniqueBrowser)
}

// suppress unused package warning
var (
    _ *json.RawMessage
    _ *jlexer.Lexer
    _ *jwriter.Writer
    _ easyjson.Marshaler
)

func easyjson9f2eff5fDecodeMypackage(in *jlexer.Lexer, out *User) {
    isTopLevel := in.IsStart()
    if in.IsNull() {
        if isTopLevel {
            in.Consumed()
        }
        in.Skip()
        return
    }
    in.Delim('{')
    for !in.IsDelim('}') {
        key := in.UnsafeString()
        in.WantColon()
        if in.IsNull() {
            in.Skip()
            in.WantComma()
            continue
        }
        switch key {
        case "browsers":
            if in.IsNull() {
                in.Skip()
                out.Browsers = nil
            } else {
                in.Delim('[')
                if out.Browsers == nil {
                    if !in.IsDelim(']') {
                        out.Browsers = make([]string, 0, 4)
                    } else {
                        out.Browsers = []string{}
                    }
                } else {
                    out.Browsers = (out.Browsers)[:0]
                }
                for !in.IsDelim(']') {
                    var v1 string
                    v1 = string(in.String())
                    out.Browsers = append(out.Browsers, v1)
                    in.WantComma()
                }
                in.Delim(']')
            }
        case "email":
            out.Email = string(in.String())
        case "name":
            out.Name = string(in.String())
        default:
            in.SkipRecursive()
        }
        in.WantComma()
    }
    in.Delim('}')
    if isTopLevel {
        in.Consumed()
    }
}
func easyjson9f2eff5fEncodeMypackage(out *jwriter.Writer, in User) {
    out.RawByte('{')
    first := true
    _ = first
    {
        const prefix string = ",\"browsers\":"
        if first {
            first = false
            out.RawString(prefix[1:])
        } else {
            out.RawString(prefix)
        }
        if in.Browsers == nil && (out.Flags&jwriter.NilSliceAsEmpty) == 0 {
            out.RawString("null")
        } else {
            out.RawByte('[')
            for v2, v3 := range in.Browsers {
                if v2 > 0 {
                    out.RawByte(',')
                }
                out.String(string(v3))
            }
            out.RawByte(']')
        }
    }
    {
        const prefix string = ",\"email\":"
        if first {
            first = false
            out.RawString(prefix[1:])
        } else {
            out.RawString(prefix)
        }
        out.String(string(in.Email))
    }
    {
        const prefix string = ",\"name\":"
        if first {
            first = false
            out.RawString(prefix[1:])
        } else {
            out.RawString(prefix)
        }
        out.String(string(in.Name))
    }
    out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v User) MarshalJSON() ([]byte, error) {
    w := jwriter.Writer{}
    easyjson9f2eff5fEncodeMypackage(&w, v)
    return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v User) MarshalEasyJSON(w *jwriter.Writer) {
    easyjson9f2eff5fEncodeMypackage(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *User) UnmarshalJSON(data []byte) error {
    r := jlexer.Lexer{Data: data}
    easyjson9f2eff5fDecodeMypackage(&r, v)
    return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *User) UnmarshalEasyJSON(l *jlexer.Lexer) {
    easyjson9f2eff5fDecodeMypackage(l, v)
}
