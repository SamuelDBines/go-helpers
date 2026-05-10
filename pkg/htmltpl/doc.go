// Package htmltpl wraps [html/template] for small Go web apps: embed FS parsing,
// a shared [template.FuncMap], and HTML responses with a consistent Content-Type.
//
// Use [DefaultFuncMap] (or [MergeFuncs] with your own funcs) before [ParseFS].
// Template files should use {{define "Name"}}…{{end}}; render with [Set.ExecuteTemplate].
//
// This package does not add a template language beyond [html/template]; it only
// reduces boilerplate across services (e.g. RepKit, BrandForge-style servers).
package htmltpl
