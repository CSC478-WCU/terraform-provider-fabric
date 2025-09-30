package main

import (
	"context"
	"flag"
	"log"

	"github.com/csc478-wcu/terraform-provider-fabric/internal/provider"
	"github.com/csc478-wcu/terraform-provider-fabric/internal/version"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

//go:generate terraform fmt -recursive ../../examples/
//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs generate -rendered-website-dir ../../docs

var buildVersion = "v0.1.0" // overridden by goreleaser

func main() {
	var debug bool
	flag.BoolVar(&debug, "debug", false, "enable delve-compatible debug mode")
	flag.Parse()

	version.Set(buildVersion)

	opts := providerserver.ServeOpts{
		Address: "registry.terraform.io/csc478-wcu/fabric",
		Debug:   debug,
	}

	if err := providerserver.Serve(context.Background(), provider.New(version.Get()), opts); err != nil {
		log.Fatal(err)
	}
}
