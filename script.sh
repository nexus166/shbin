#!/usr/bin/env bash

# this is an example script.
# replace this file (keeping the name) with the shell script you want to embed.

for i in id ls pwd w whoami; do 
    echo "running $i"
    $i
    echo -e "\\n\\n"
done

echo -e "printing \$0"
echo $0

echo -e "\\n\\n\\nshowing args"
echo $*