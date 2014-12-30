#!/bin/bash

export rmnpath=$RMN_BASE_PATH
export guipath=$rmnpath/www/gui

FILES=$(find $rmnpath/www/gui/sass -name "[^_]*\.scss")

time wt -gen $guipath/build/im -font $guipath/font-face -b $guipath/build/css/ -p $guipath/sass -d $guipath/im/sass $FILES
