package main

import (
	"fmt"
	"github.com/containous/yaegi/interp"
	"github.com/containous/yaegi/stdlib"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

func runFileAsLambda(ctx *Context) error {
	i := interp.New(interp.Options{})
	i.Use(stdlib.Symbols)

	f, _ := os.Open(ctx.path)
	fS, _ := ioutil.ReadAll(f)

	//fmt.Println(string(fS))
	log.Printf("Running lambda at %v", ctx.path)
	_, err := i.Eval(string(fS))
	if err != nil {
		return err
	}
	lambdaDef, err := i.Eval("lambda.Handler")
	if err != nil {
		return err
	}

	lambda, ok := lambdaDef.Interface().(func(http.ResponseWriter, *http.Request))
	if !ok {
		return fmt.Errorf("not correctly formatted lambda")
	}
	lambda(ctx.w, ctx.r)
	return nil
}
