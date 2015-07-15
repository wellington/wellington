#!/bin/bash

export rmnpath=$RMN_BASE_PATH
export guipath=$rmnpath/www/gui

FILES=$(find $rmnpath/www/gui/sass -name "[^_]*\.scss")
echo wt compile --gen $guipath/build/im --font $guipath/font-face --destination $guipath/build/css/ -p $guipath/sass --images-dir $guipath/im/sass $FILES

time wt compile --gen $guipath/build/im --font $guipath/font-face --destination $guipath/build/css/ -p $guipath/sass --images-dir $guipath/im/sass $FILES
