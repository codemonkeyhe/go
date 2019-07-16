module main_mod

go 1.12

//require github.com/robteix/testmod v1.0.1

require local.com/xyz/testmod v1.0.0

//replace local/gomod_demo/testmod => /e/gitPro/go/gomod_demo/testmod
//replace local/gomod_demo/testmod => E:\gitPro\go\gomod_demo\testmod
//replace local/gomod_demo/testmod => e:\gitPro\go\gomod_demo\testmod

//replace local.com/xyz/testmod => e:/gitPro/go/gomod_demo/testmod

//require github.com/robteix/testmod v1.0.1
