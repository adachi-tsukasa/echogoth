package main

import "go.uber.org/zap"

var logger *zap.SugaredLogger

func setupLog() {
	// Console Mode
	l, _ := zap.NewDevelopment()
	// Output Mode
	// config := zap.NewDevelopmentConfig()
	// config.OutputPaths = []string{"app.log"}
	// l, _ := config.Build()
	logger = l.Sugar()
}
