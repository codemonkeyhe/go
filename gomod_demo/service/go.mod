module demo.com/gomod_demo/service

go 1.12

//require github.com/robteix/testmod v1.0.1

require local.com/xyz/testmod v1.0.0

//replace local/gomod_demo/testmod => /e/gitPro/go/gomod_demo/testmod
//replace local/gomod_demo/testmod => E:\gitPro\go\gomod_demo\testmod
//replace local/gomod_demo/testmod => e:\gitPro\go\gomod_demo\testmod

replace local.com/xyz/testmod => e:/gitPro/go/gomod_demo/testmod

//replace demo.com => e:/gitPro/go

//require github.com/robteix/testmod v1.0.1
