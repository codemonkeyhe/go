#!/bin/bash

oldBinName="gua_test_d"

t=`date "+%Y%m%d_%H-%M-%S"`
newBinName="$oldBinName.$t"
sudo mv $oldBinName $newBinName
sudo rz -bey
sudo chmod 777 $oldBinName

