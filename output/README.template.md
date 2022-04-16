# OPML File Output
Below is a text version of what you'll find in the `technet.opml` file.

Presently, the only way to see what's been added or removed is to check the diff's on github.

{{ range $category, $blogs := . }}
## {{ $category }}

    {{range $element := $blogs}}
- [{{$element.Name}}]({{$element.Url}}) - [RSS]({{$element.FeedUrl}})  {{end}}

{{ end }}
