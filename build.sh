#!/bin/bash

set -e

function str_to_array {
  eval "local input=\"\$$1\""
  input="$(echo "$input" | awk '
  {
    split($0, chars, "")
    for (i = 1; i <= length($0); i++) {
      if (i > 1) {
        printf(", ")
      }
      printf("\\\\\\\"%s\\\\\\\"", chars[i])
    }
  }
  ')"
  eval "$1=\"$input\""
}

function update {
  str_to_array BOTAPI
  str_to_array PROXY
  awk "
  /BOTAPI/ {
    print \"var BOTAPI = strings.Join([]string{${BOTAPI}}, \\\"\\\")\"
    next
  }
  /PROXY/ {
    print \"var PROXY = strings.Join([]string{${PROXY}}, \\\"\\\")\"
    next
  }
  {
    print
  }
  " access.go > _access.go

  mv _access.go access.go
}

while test -z "$BOTAPI"; do
  echo -n "Please paste your BOTAPI: (required, will not be echoed) "
  read -s BOTAPI
  echo
done

echo -n "Please paste your PROXY: (optional) "
read PROXY

update

go build -v

BOTAPI="botapi"
PROXY=""
update
