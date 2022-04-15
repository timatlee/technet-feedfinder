# OPML File goes here.

I think this directory needs to exist for Go to write to it..  and eventually, I want this file to have a text version of the content anyways, soo..

{{ range $category, $blogs := . }}
## {{ $category }}

    {{range $element := $blogs}}
- [{{$element.Name}}]({{$element.Url}}) - [RSS]({{$element.FeedUrl}})  {{end}}

{{ end }}
