#!/usr/bin/env bash

printf "\n#1: Check Service\n"
curl -s -H "Authorization: Bearer Token" http://localhost:9090 | jq

printf "\n#2: Upload One File\n"
file=`curl -s -H "Authorization: Bearer Token" http://localhost:9090/tenants/ns:01/uploads -F hello.png=@test/hello.png`
echo $file | jq
url=`echo $file | jq ".url"`

printf "\n#2: Upload Multiple Files\n"
curl -s -H "Authorization: Bearer Token" http://localhost:9090/tenants/ns:01/uploads -F hello.png=@test/hello.png -F world.png=@test/world.png | jq

printf "\n#2: Get Uploaded File\n"
curl -s -H "Authorization: Bearer Token" http://localhost:9090/$url \
  -o /dev/null -w "%{http_code} | %{content_type} | %{size_download} bytes\n"

printf "\n#3: Upload an invalid File\n"
curl -s -H "Authorization: Bearer Token" http://localhost:9090/tenants/ns:01/uploads -d '{ "name": "Roller Skates"}'
