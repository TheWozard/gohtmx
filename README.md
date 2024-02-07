# gohtmx

A component library for building composable golang APIs using HTMX and html/template.

## Goals
The end goal is to have an easy to compose UI component system that allows full stack development from pure golang. The UI should be easy to pre-populate with data and support deep linking functionality as well as leave the door open for more complex functionality.

## Status
Currently the package is highly unstable and going through many rewrites. Once the interface is more solid, unit testing will follow. For now its just free form experimentation.

# Architecture
gohtmx is in the end generating html/templates. It provides ergonomic creation of templates through compositional structs and enabling inline usage of golang functions for template data interaction.
