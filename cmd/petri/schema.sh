#!/bin/bash

filename=".schema" # name of the file containing types list
while read -r type; do
    typescript-json-schema tsconfig.json "$type" -o "app/forms/$type.schema.json"
done < "$filename"