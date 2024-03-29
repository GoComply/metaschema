# Metaschema modeling language ![Build](https://github.com/GoComply/metaschema/workflows/Metaschema%20Build%20CI/badge.svg) ![Process NIST OSCAL](https://github.com/GoComply/metaschema/workflows/Process%20NIST%20OSCAL/badge.svg) [![PkgGoDev](https://pkg.go.dev/badge/github.com/gocomply/metaschema)](https://pkg.go.dev/github.com/gocomply/metaschema)

## Introduction

Metaschema framework is s an information modeling methodology. Metaschema
framework is developed by [NIST](https://pages.nist.gov/metaschema/). An
information model developed using this framework can be used to automatically:

 * Generate associated XML and JSON schema
 * Produce model documentation
 * Create content converters capable of converting between XML and JSON formats
 * Data APIs for use in application code

## Golang Extension

This project extends metaschema beyond xml/json/yaml. This project allows users
to generate golang code for processing those xml/json/yaml files out of NIST's
metaschema.

## Usage

```
# Acquire latest OSCAL metaschema (OSCAL is the most evolved appliacation of the metaschema)
git clone --depth 1 https://github.com/usnistgov/OSCAL
# Parse metaschema and generate golang structs
./gocomply_metaschema generate ./OSCAL/src/metaschema github.com/gocomply/oscalkit types/oscal
```

## Installation

```
go get -u -v github.com/gocomply/metaschema/cli/gocomply_metaschema
```

## Projects using GoComply/metaschema

 * [GoComply/oscalkit](https://github.com/GoComply/oscalkit/tree/master/types/oscal) - OSCAL implementation, oscalkit uses CI to re-generate models periodically from [usnistgov/OSCAL](https://github.com/usnistgov/OSCAL)
