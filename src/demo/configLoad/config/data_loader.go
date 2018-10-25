package config

import (
	"common/applog"

	//"gua.yy.com/def"
	"fmt"
	"strconv"
	"sync"
	"time"
)

//单例模式的数据定时加载类
type dataLoader struct {
	data map[string]string
	sync.RWMutex
}

//全局对象的指针
var instance *dataLoader
var once sync.Once

func GetInstance() *dataLoader {
	once.Do(func() {
		instance = &dataLoader{}
	})
	return instance
}

func (p *dataLoader) Init() {
	applog.Info("Init Config")
	p.data = make(map[string]string)
	for i := 0; i < 10000; i++ {
		val := strconv.FormatInt(int64(i), 10)
		p.data[val] = val
	}
}

func (p *dataLoader) Start() {
	fmt.Println("start...")
	for {
		now := time.Now().Format("2006-01-02 15:04:05")
		applog.Info("%v , Running...", now)
		p.run()
		time.Sleep(time.Second * 5)
	}
}

//让单例类永驻内存，从而生命周期持续为整个应用周期
func (p *dataLoader) run() {
	tdata := make(map[string]string)
	p.loadData(tdata)
	{
		p.Lock()
		//swap
		applog.Info("BEFORE: origin size=%d, tmp size=%d", len(p.data), len(tdata))
		applog.Info("BEFORE: origin ADDR=%p, tmp ADDR=%p", p.data, tdata)
		p.data, tdata = tdata, p.data
		//p.data = tdata
		applog.Info("END: origin size=%d, tmp size=%d", len(p.data), len(tdata))
		applog.Info("END: origin ADDR=%p, tmp ADDR=%p", p.data, tdata)
		p.Unlock()
	}
}

func (p *dataLoader) loadData(tdata map[string]string) {
	for i := 0; i < 9000; i++ {
		val := strconv.FormatInt(int64(i+10000), 10)
		tdata[val] = val
	}
}

func (p *dataLoader) Stat() {
	size := 0
	{
		p.RLock()
		size = len(p.data)
		p.RUnlock()

	}
	fmt.Println("Stat: size=%d", size)
	applog.Info("Stat: size=%d", size)
}

func (p *dataLoader) readC(key string) string {
	fmt.Println("read data!key=%s", key)
	p.RLock()
	value, ok := p.data[key]
	p.RUnlock()
	//ok == false
	applog.Info("read ok=%v v=%v", ok, value)
	return value
}

func (p *dataLoader) writeC(key string, value string) {
	fmt.Println("write data,key=%s value=%s", key, value)
	p.Lock()
	p.data[key] = value
	p.Unlock()
	//fmt.Println("write data Success,key=%s", key)
}
