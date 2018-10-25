#!/bin/sh

/bin/rm -f */*.exe  
/bin/rm -f  */*/*.exe
/bin/rm  -f */test

dirlist=`ls | egrep -v "clean"`
for pro in $dirlist
do 
    proDir="$pro"
    binName="$proDir/$proDir"
    echo "rm -rf $binName"
    rm -rf $binName 
done

