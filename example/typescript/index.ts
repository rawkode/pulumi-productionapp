import * as pulumi from "@pulumi/pulumi";
import * as rawkode from "@rawkode/pulumi-productionapp";

const r = new rawkode.Deployment("example", {
  image: "nginx",
  port: 80,
});
