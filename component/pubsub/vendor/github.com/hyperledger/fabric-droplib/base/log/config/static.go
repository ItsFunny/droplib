/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/5/14 11:23 上午
# @File : static.go
# @Description :
# @Attention :
*/
package config

//
//
//
// func GetLogConfig() *LogConfiguration {
// 	return logConfig
// }
//
// // FIXME FILTER
// // 只有比配置大的时候打印
// func IsLogLevelDisabled(level common.Level, moduleName string) bool {
// 	if logConfig.LogLevel > level {
// 		return true
// 	}
// 	_, exist := logConfig.blackModuleSet[moduleName]
// 	return exist
// }
// func FindCaller(skip int) (string, bool) {
// 	return common.FindCaller(skip, logConfig.blackList)
// }
//
// func GetModuleLevel(m string) common.Level {
// 	return logConfig.GetModuleLevel(m)
// }
//
// func RegisterBlackList(pathes ...string) {
// 	logConfig.lock.Lock()
// 	defer logConfig.lock.Unlock()
// 	for _, path := range pathes {
// 		logConfig.blackList = append(logConfig.blackList, path)
// 	}
// }
// func RegisterBlackModule(modules ...string) {
// 	logConfig.lock.Lock()
// 	defer logConfig.lock.Unlock()
// 	for _, module := range modules {
// 		module = strings.ToUpper(module)
// 		logConfig.blackModuleSet[module] = struct{}{}
// 	}
// }
// func RegisterModuleLevel(m map[string]common.Level) {
// 	logConfig.lock.Lock()
// 	defer logConfig.lock.Unlock()
// 	for module, l := range m {
// 		module = strings.ToUpper(module)
// 		logConfig.moduleLevel[module] = l
// 	}
// }
//
// func SetGlobalLogLevel(l common.Level) {
// 	logConfig.LogLevel = l
// 	logrus.SetLevel(l.GetLogrusLevel())
// }
//
// func NotifyAsReady() {
// 	logConfig.status = common.READY
// }
