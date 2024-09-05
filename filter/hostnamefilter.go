package filter

import (
	log "github.com/kataras/golog"
	"regexp"
	"strings"
)

var (

	AllowedHostNameReg  *regexp.Regexp = nil
	DisallowedHostNameReg *regexp.Regexp = nil

	HostNameACL map[string]bool
)

func Init() {
	InitWithFilter(DefaultAllowHostname,DefaultDisallowHostname)
	HostNameACL = make(map[string]bool)
}

func InitWithFilter(allowedHostName []string,disallowedHostName []string) {
	if len(allowedHostName) > 0 {
		var regPattern string
		for index,v := range allowedHostName {
			//将通配符转换为正则表达式
			regPattern += WildcardPattern2RegexpPattern(v)
			if index != (len(allowedHostName) -1 ) {
				regPattern +=  "|"
			}
		}
		AllowedHostNameReg = regexp.MustCompile(regPattern)
	}

	if len(disallowedHostName) > 0 {
		var regPattern string
		for index,v := range disallowedHostName {
			regPattern += WildcardPattern2RegexpPattern(v)
			if index != (len(disallowedHostName) -1 ) {
				regPattern += "|"
			}
		}
		DisallowedHostNameReg = regexp.MustCompile(regPattern)
	}
	log.Infof("[Filter] : Allowed Hostname : %s",strings.Join(allowedHostName,", "))
	log.Infof("[Filter] : Disallowed Hostname : %s",strings.Join(disallowedHostName,", "))

}

func IsAllowHostName(hostname string) bool {
	// 默认不允许
	var isAllow bool = false
	if AllowedHostNameReg == nil {
		//允许的主机名为空,也就是当前是黑名单模式，默认允许
		isAllow = true
		return isAllow
	}
	//允许的主机名不为空,也就是当前是白名单模式
	if status,ok := HostNameACL[hostname]; ok {
		//在表中找到了相关的条目,直接返回
		isAllow = status
	} else if AllowedHostNameReg.MatchString(hostname) {
		//只有开启了白名单模式才匹配域名允许
		isAllow = true
		HostNameACL[hostname] = true
	}

	return isAllow
}

func IsDisallowHostName(hostname string) bool {
	//默认所有的域名都是允许的,后面根据域名的匹配情况来决定是否不允许该域名
	var isDisallow bool = false
	if AllowedHostNameReg != nil {
		//允许的主机名不为空,也就是当前是白名单模式，默认所有的域名都是不允许的
		//判断是否当前域名为允许的域名,判断后直接返回，不需要进行下一步
		isDisallow = !(AllowedHostNameReg.MatchString(hostname))
		return isDisallow
	}
	//允许的主机名为空,也就是当前是黑名单模式
	if status,ok := HostNameACL[hostname]; ok {
		//在表中找到了相关的条目,直接返回
		isDisallow = !status
	} else if !isDisallow && DisallowedHostNameReg != nil && DisallowedHostNameReg.MatchString(hostname) {
		//默认所有域名都允许，但是匹配了就不允许当前域名
		isDisallow = true
		HostNameACL[hostname] = false
	}
	
	return isDisallow
}

func IsNeedProcessHostname(hostname string) bool {
	if (IsAllowHostName(hostname) && !IsDisallowHostName(hostname)) {
		//当前域名是允许处理的域名或者是不在不允许处理的域名里面，处理该域名
		return true
	} else {
		log.Debugf("[Filter] : skipping hostname %s",hostname)
		return false
	}
}