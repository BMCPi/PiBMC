#!/bin/bash
# SPDX-License-Identifier: BSD-3-Clause

# Find the schema document name by going here:
#
#     https://www.dmtf.org/standards/redfish
#
# Inspect the url for the schema you want - for example, the 2020.1 update
# document is "DSP8010_2020.1.zip" on this page. The base name then is:
schemadoc="DSP8010_2023.3"

# Check if filename provided on the command line
if [[ "$#" -eq 1 ]]; then
    schemadoc="${1}"
fi

# See if we already have this locally or if we need to fetch it
if [[ ! -d $schemadoc ]]; then
    if [[ ! -f "${schemadoc}.zip" ]]; then
        # Use curl instead of wget because it is more likely to be present
        echo "Fetching schema document $schemadoc"
        curl -G -L "https://www.dmtf.org/sites/default/files/standards/documents/${schemadoc}.zip" > "${schemadoc}.zip"
    fi

    echo "Extracting schema files..."
    unzip -q "${schemadoc}.zip" -d "${schemadoc}"
fi

schema_objects=$(find "${schemadoc}/openapi" -name "*.yaml" | cut -d '/' -f 3 | cut -d '.' -f 1 | sort | uniq )
for schema in ${schema_objects}; do
    latest_schema=$(find "${schemadoc}/openapi" -name "${schema}.*.yaml" | cut -d '/' -f 3 | cut -d '.' -f 2 | sort | uniq | tail -n1)
    if [[ -z "${latest_schema}" ]]; then
        schema_collection=("${schema_collection[@]}" "${schemadoc}/openapi/${schema}.yaml")
    else
        schema_collection=("${schema_collection[@]}" "${schemadoc}/openapi/${schema}.yaml" "${schemadoc}/openapi/${schema}.${latest_schema}.yaml")
    fi
done

unset 'schema_collection[${#schema_collection[@]}-1]'

printf '%s\n' "${schema_collection[@]}" | grep -v "${schemadoc}/openapi/odata.yaml" | sort | uniq | jq -Rn '{ inputs: [inputs | { inputFile: "\(.)" }], output: "./output.swagger.json" }' > openapi-merge.json

npx openapi-merge-cli

go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen -package redfish -o redfish/server.gen.go -generate std-http-server,models \
    --import-mapping http://redfish.dmtf.org/schemas/v1/ServiceRoot.v1_16_1.yaml \
    output.swagger.json

exit 0

# Generates a list of schema object from a given schema zip or directory then
# generates go files based on the provided generate_from_schema.py tool and
# accompanying source.tmpl file.

# Set correct name for python3 executable. Some platforms just call it python
# while others call it python3.
PYTHON="python3"

# Make sure we're running in a virtual environment
if [[ -z "$VIRTUAL_ENV" ]]; then
    if [[ ! -d ./env ]]; then
        $PYTHON -m venv env
    fi

    if [[ -d ./env/Scripts ]]; then
        # Windows
        source ./env/Scripts/activate
    else
        source ./env/bin/activate
    fi
    $PYTHON -m pip install -r requirements.txt
fi



find "${schemadoc}/openapi" -name "*.yaml" | jq -Rn '{ inputs: [inputs | { inputFile: "\(.)" }], output: "./output.swagger.json" }' > openapi-merge.json

npx openapi-merge-cli

go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen -package redfish -o api/redfish/server.gen.go -generate std-http-server,models tools/output.swagger.json


# Generate the object list. Not elegant, but it works well enough for now.
#
# SerialInterface and Switch still have some identifier issues to be worked out
# Collection files, Schedule, are "common" and included differently
# redfish-schema, odata and all the versions of each object aren't needed
#
# General process is get a list of the json-schema objects from the zip, drop
# things we don't need/want, and clip the column we want generating a file of
# object names we can use later.


exit 0

# Now we're ready to generate the go source based on these files
if [[ ! -d gofiles ]]; then
    mkdir gofiles
fi

# Loop through each one and generate the raw source file
for object in $schema_objects; do
    echo "Processing object ${object}..."
    $PYTHON generate_from_schema.py -l "${schemadoc}/json-schema" -t redfish "${object}" -o "gofiles/${object}.go"
done

# Then clean them up
echo "Running go fmt on generated files"
go fmt ./gofiles/*.go

echo "Processing Complete"
echo "(Ready for manual cleanup)"