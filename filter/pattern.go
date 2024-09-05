package filter

import "strings"

func WildcardPattern2RegexpPattern(pattern string) string {
	var regPattern string  = pattern

	//将通配符的.转换为\.,注意先后顺序
	regPattern = strings.ReplaceAll(regPattern,".","\\.")

	//将通配符的*转换为.*
	regPattern = strings.ReplaceAll(regPattern,"*",".*")

	return regPattern

}