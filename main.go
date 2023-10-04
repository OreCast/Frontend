package main

// Go implementation of OreCast Frontend service
//
// Copyright (c) 2023 - Valentin Kuznetsov <vkuznet@gmail.com>
//

import (
	_ "expvar"         // to be used for monitoring, see https://github.com/divan/expvarmon
	_ "net/http/pprof" // profiler, see https://golang.org/pkg/net/http/pprof/

	oreConfig "github.com/OreCast/common/config"
)

func main() {
	oreConfig.Init()
	Server()
}
