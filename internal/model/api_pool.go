package model

import (
	"compress/gzip"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
	"zbxtable/pkg/logger"

	zabbix "github.com/canghai908/zabbix-go"
)

// APIPool API 连接池，管理多个 Zabbix 实例的 API 连接
type APIPool struct {
	mu    sync.RWMutex
	pools map[int]*APIInstance // key: ZabbixInstance.ZID

	// building 为每个 ZID 维护一把独立的构建锁，保证：
	// 1）同一实例的并发请求只建连一次（避免重复握手）；
	// 2）不同实例可并行建连；
	// 3）建连期间的网络 I/O 不持有 mu，避免某个实例慢/宕机拖垮整个连接池。
	buildMu  sync.Mutex
	building map[int]*sync.Mutex
}

// APIInstance 单个 Zabbix 实例的 API 连接信息
type APIInstance struct {
	ZID          int
	Instance     string
	Name         string
	WebURL       string
	API          *zabbix.API
	JAR          *Jar // 用于 Web 登录的 Cookie
	Version      string
	IsV54OrLater bool // 是否为 5.4 或更高版本
	LastUpdate   time.Time
}

var (
	apiPool     *APIPool
	apiPoolOnce sync.Once
)

// GetAPIPool 获取全局 API 连接池单例
func GetAPIPool() *APIPool {
	apiPoolOnce.Do(func() {
		apiPool = &APIPool{
			pools:    make(map[int]*APIInstance),
			building: make(map[int]*sync.Mutex),
		}
	})
	return apiPool
}

// getFresh 返回缓存中未过期（30 分钟内）的连接
func (p *APIPool) getFresh(zid int) (*APIInstance, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if inst, ok := p.pools[zid]; ok && time.Since(inst.LastUpdate) < 30*time.Minute {
		return inst, true
	}
	return nil, false
}

// zidLock 返回指定 ZID 的构建锁（惰性创建）
func (p *APIPool) zidLock(zid int) *sync.Mutex {
	p.buildMu.Lock()
	defer p.buildMu.Unlock()
	if p.building == nil {
		p.building = make(map[int]*sync.Mutex)
	}
	l, ok := p.building[zid]
	if !ok {
		l = &sync.Mutex{}
		p.building[zid] = l
	}
	return l
}

// GetOrCreateAPI 获取或创建指定实例的 API 连接
func (p *APIPool) GetOrCreateAPI(zid int) (*APIInstance, error) {
	// 快路径：命中未过期的缓存连接
	if inst, ok := p.getFresh(zid); ok {
		return inst, nil
	}

	// 慢路径：按 ZID 加构建锁。网络握手全程不持有 p.mu，
	// 不同实例可并行建连，同一实例的并发请求只建连一次。
	lock := p.zidLock(zid)
	lock.Lock()
	defer lock.Unlock()

	// 双重检查：可能已被其他 goroutine 建好
	if inst, ok := p.getFresh(zid); ok {
		return inst, nil
	}

	// 从数据库加载实例配置
	instance, err := GetZabbixInstanceByZID(zid)
	if err != nil {
		return nil, fmt.Errorf("获取实例配置失败: %w", err)
	}

	if !instance.Enabled {
		return nil, errors.New("实例已禁用")
	}

	// 解密密码和Token
	decryptedPass, decryptedToken := DecryptInstanceCredentials(instance)

	// 创建 API 连接
	webURL := strings.TrimRight(strings.TrimSpace(instance.URL), "/")
	apiURL := webURL + "/api_jsonrpc.php"

	// 测试连接
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := http.Client{Transport: transport, Timeout: 5 * time.Second}
	resp, err := client.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("连接失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusPreconditionFailed {
		return nil, fmt.Errorf("API 地址不正确，状态码: %d", resp.StatusCode)
	}

	// 创建 API 对象
	api := zabbix.NewAPI(apiURL)
	if strings.TrimSpace(decryptedToken) != "" {
		api.SetAuth(strings.TrimSpace(decryptedToken))
	} else if strings.TrimSpace(instance.User) != "" && strings.TrimSpace(decryptedPass) != "" {
		_, err := api.Login(strings.TrimSpace(instance.User), strings.TrimSpace(decryptedPass))
		if err != nil {
			return nil, fmt.Errorf("登录失败: %w", err)
		}
	} else {
		return nil, errors.New("未配置认证信息（Token 或 用户名/密码）")
	}

	// 获取版本信息
	version, err := api.Version()
	if err != nil {
		return nil, fmt.Errorf("获取版本失败: %w", err)
	}

	// 判断版本
	isV54OrLater := false
	verArr := strings.Split(version, ".")
	if len(verArr) >= 2 {
		major, _ := strconv.ParseInt(verArr[0], 10, 64)
		minor, _ := strconv.ParseInt(verArr[1], 10, 64)
		if major >= 6 || (major == 5 && minor >= 4) {
			isV54OrLater = true
		}
	}

	// 创建 Web 登录 JAR（用于图形查看）
	jar := new(Jar)
	if strings.TrimSpace(instance.User) != "" && strings.TrimSpace(decryptedPass) != "" {
		// 执行 Web 登录
		logger.Log.Infof("尝试登录 Zabbix Web 界面: %s (实例: %s)", webURL, instance.Name)
		err := loginToZabbixWeb(webURL, instance.User, decryptedPass, jar)
		if err != nil {
			// Web 登录失败不影响 API 连接的创建，仅记录警告日志
			logger.Log.Warnf("Zabbix Web 登录失败 (实例: %s, URL: %s): %v，API 连接仍可正常使用，但图形查看功能可能受限", instance.Name, webURL, err)
		} else {
			logger.Log.Infof("Zabbix Web 登录成功: %s (实例: %s)", webURL, instance.Name)
		}
	} else {
		logger.Log.Debugf("实例 %s 未配置用户名/密码，跳过 Web 登录", instance.Name)
	}

	inst := &APIInstance{
		ZID:          zid,
		Instance:     instance.Instance,
		Name:         instance.Name,
		WebURL:       webURL,
		API:          api,
		JAR:          jar,
		Version:      version,
		IsV54OrLater: isV54OrLater,
		LastUpdate:   time.Now(),
	}

	p.mu.Lock()
	p.pools[zid] = inst
	p.mu.Unlock()

	logger.Log.Infof("创建 API 连接成功: %s (ZID=%d, Version=%s)", instance.Name, zid, version)
	return inst, nil
}

// GetAllEnabledAPIs 获取所有启用实例的 API 连接
func (p *APIPool) GetAllEnabledAPIs() ([]*APIInstance, error) {
	// 查询所有启用的实例
	var instances []ZabbixInstance
	err := DB.Where("enabled = ?", true).Find(&instances).Error
	if err != nil {
		return nil, err
	}

	if len(instances) == 0 {
		return nil, errors.New("没有启用的 Zabbix 实例")
	}

	var apiInstances []*APIInstance
	var errs []string

	for _, instance := range instances {
		inst, err := p.GetOrCreateAPI(instance.ID)
		if err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", instance.Name, err))
			logger.Log.Errorf("获取实例 %s 的 API 连接失败: %v", instance.Name, err)
			continue
		}
		apiInstances = append(apiInstances, inst)
	}

	if len(apiInstances) == 0 {
		return nil, fmt.Errorf("所有实例连接失败: %s", strings.Join(errs, "; "))
	}

	return apiInstances, nil
}

// RemoveAPI 移除指定实例的 API 连接（用于实例删除或禁用时）
func (p *APIPool) RemoveAPI(zid int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.pools, zid)
	logger.Log.Infof("移除 API 连接: ZID=%d", zid)
}

// ClearAll 清空所有连接（用于重启或重新加载配置）
func (p *APIPool) ClearAll() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.pools = make(map[int]*APIInstance)
	logger.Log.Info("清空所有 API 连接")
}

// loginToZabbixWeb 登录 Zabbix Web 界面（用于图形查看）
func loginToZabbixWeb(webURL, user, pass string, jar *Jar) error {
	v := url.Values{}
	v.Set("name", user)
	v.Add("password", pass)
	v.Add("autologin", "1")
	v.Add("enter", "Sign in")

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Transport: tr,
		Jar:       jar,
		Timeout:   10 * time.Second,
	}

	request, err := http.NewRequest("POST", webURL+"/index.php", strings.NewReader(v.Encode()))
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	request.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
	request.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	request.Header.Add("Accept-Encoding", "gzip, deflate")
	request.Header.Add("Accept-Language", "zh-cn,zh;q=0.8,en-us;q=0.5,en;q=0.3")
	request.Header.Add("Connection", "keep-alive")
	request.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 6.1; WOW64; rv:12.0) Gecko/20100101 Firefox/12.0")

	response, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("登录请求失败: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		return fmt.Errorf("登录失败，状态码: %d", response.StatusCode)
	}

	// 检查响应内容
	var reader io.Reader
	switch response.Header.Get("Content-Encoding") {
	case "gzip":
		reader, err = gzip.NewReader(response.Body)
		if err != nil {
			return fmt.Errorf("解压响应失败: %w", err)
		}
	default:
		reader = response.Body
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("读取响应失败: %w", err)
	}

	if strings.Contains(string(data), "blocked") {
		return errors.New("登录被阻止，请检查用户名和密码")
	}
	return nil
}

// GetAPIByZID 便捷函数：根据实例 ZID 获取 API 连接
func GetAPIByZID(id int) (*APIInstance, error) {
	return GetAPIPool().GetOrCreateAPI(id)
}

// GetAllEnabledAPIInstances 便捷函数：获取所有启用的 API 实例
func GetAllEnabledAPIInstances() ([]*APIInstance, error) {
	return GetAPIPool().GetAllEnabledAPIs()
}

// GetZabbixInstanceAPI 根据 instance 字符串获取 API 实例
// func GetZabbixInstanceAPI(id string) (*APIInstance, error) {
// 	if id == "" {
// 		return nil, errors.New("id 不能为空")
// 	}
// 	zid, _ := strconv.Atoi(id)
// 	// 根据 instance 查询实例
// 	instance, err := GetZabbixInstanceByZID(zid)
// 	if err != nil {
// 		return nil, fmt.Errorf("未找到启用的实例 (zid=%v): %w", zid, err)
// 	}

// 	if !instance.Enabled {
// 		return nil, fmt.Errorf("实例已禁用 (id=%s)", id)
// 	}

// 	// 获取或创建 API 连接
// 	return GetAPIPool().GetOrCreateAPI(instance.ID)
// }

// GetZabbixInstanceAPIByZID 根据 ZID 获取 API 实例（新增便捷函数）
// func GetZabbixInstanceAPIByZID(zid int) (*APIInstance, error) {
// 	return GetAPIByZID(zid)
// }
