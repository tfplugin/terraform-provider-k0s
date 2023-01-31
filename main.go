package main

import (
	"context"
	"flag"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/tfplugin/terraform-provider-k0s/k0s"
	"log"
)

func main() {
	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := providerserver.ServeOpts{
		Address: "registry.terraform.io/tfplugin/k0s",
		Debug:   debug,
	}

	err := providerserver.Serve(context.Background(), k0s.New(), opts)

	if err != nil {
		log.Fatal(err.Error())
	}
}
