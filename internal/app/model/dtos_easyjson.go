// Code generated by easyjson for marshaling/unmarshaling. DO NOT EDIT.

package model

import (
	json "encoding/json"
	easyjson "github.com/mailru/easyjson"
	jlexer "github.com/mailru/easyjson/jlexer"
	jwriter "github.com/mailru/easyjson/jwriter"
)

// suppress unused package warning
var (
	_ *json.RawMessage
	_ *jlexer.Lexer
	_ *jwriter.Writer
	_ easyjson.Marshaler
)

func easyjson782a897aDecodeGithubComUjweghShortenerInternalAppModel(in *jlexer.Lexer, out *ShortenedURL) {
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
		key := in.UnsafeFieldName(false)
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "uuid":
			if data := in.UnsafeBytes(); in.Ok() {
				in.AddError((out.UUID).UnmarshalText(data))
			}
		case "short_url":
			out.ShortURL = string(in.String())
		case "original_url":
			out.OriginalURL = string(in.String())
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
func easyjson782a897aEncodeGithubComUjweghShortenerInternalAppModel(out *jwriter.Writer, in ShortenedURL) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"uuid\":"
		out.RawString(prefix[1:])
		out.RawText((in.UUID).MarshalText())
	}
	{
		const prefix string = ",\"short_url\":"
		out.RawString(prefix)
		out.String(string(in.ShortURL))
	}
	{
		const prefix string = ",\"original_url\":"
		out.RawString(prefix)
		out.String(string(in.OriginalURL))
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v ShortenedURL) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjson782a897aEncodeGithubComUjweghShortenerInternalAppModel(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v ShortenedURL) MarshalEasyJSON(w *jwriter.Writer) {
	easyjson782a897aEncodeGithubComUjweghShortenerInternalAppModel(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *ShortenedURL) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson782a897aDecodeGithubComUjweghShortenerInternalAppModel(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *ShortenedURL) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson782a897aDecodeGithubComUjweghShortenerInternalAppModel(l, v)
}
func easyjson782a897aDecodeGithubComUjweghShortenerInternalAppModel1(in *jlexer.Lexer, out *ShortenResponseDto) {
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
		key := in.UnsafeFieldName(false)
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "result":
			out.Result = string(in.String())
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
func easyjson782a897aEncodeGithubComUjweghShortenerInternalAppModel1(out *jwriter.Writer, in ShortenResponseDto) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"result\":"
		out.RawString(prefix[1:])
		out.String(string(in.Result))
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v ShortenResponseDto) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjson782a897aEncodeGithubComUjweghShortenerInternalAppModel1(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v ShortenResponseDto) MarshalEasyJSON(w *jwriter.Writer) {
	easyjson782a897aEncodeGithubComUjweghShortenerInternalAppModel1(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *ShortenResponseDto) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson782a897aDecodeGithubComUjweghShortenerInternalAppModel1(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *ShortenResponseDto) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson782a897aDecodeGithubComUjweghShortenerInternalAppModel1(l, v)
}
func easyjson782a897aDecodeGithubComUjweghShortenerInternalAppModel2(in *jlexer.Lexer, out *ShortenRequestDto) {
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
		key := in.UnsafeFieldName(false)
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "url":
			out.URL = string(in.String())
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
func easyjson782a897aEncodeGithubComUjweghShortenerInternalAppModel2(out *jwriter.Writer, in ShortenRequestDto) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"url\":"
		out.RawString(prefix[1:])
		out.String(string(in.URL))
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v ShortenRequestDto) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjson782a897aEncodeGithubComUjweghShortenerInternalAppModel2(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v ShortenRequestDto) MarshalEasyJSON(w *jwriter.Writer) {
	easyjson782a897aEncodeGithubComUjweghShortenerInternalAppModel2(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *ShortenRequestDto) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson782a897aDecodeGithubComUjweghShortenerInternalAppModel2(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *ShortenRequestDto) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson782a897aDecodeGithubComUjweghShortenerInternalAppModel2(l, v)
}
