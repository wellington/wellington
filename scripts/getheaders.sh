#!/bin/sh

DIR=${1:-libsass}

# Fetch headers from libsass project, assumes ./libsass
FILES="sass.h sass_context.h sass_functions.h sass_interface.h sass_values.h"

for F in $FILES
do
	cp $DIR/$F context/$F
done
