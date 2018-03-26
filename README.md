Imagestream Import Smoketest
============================

Building
--------

To build the binary, run

```
$ make
```

To build the image, run

```
$ docker build -t docker.io/bparees/openshift-monitor-imagestream-import .
```

(You can tag it as any name you want, but you'll need to update `install/manifests/monitor-imagestream-smoketest.yaml` to reference
the appropriate name if you change it).


To build the RPM and RPM-based image, run

```
$ OS_BUILD_ENV_PRESERVE=_output/local/bin hack/env make build-images
```

Running
-------

Create a namespace named `imagestream-smoketest`

```
$ oc create project imagestream-smoketest
```

Deploy the smoketest in that project
```
$ oc new-app -f install/manifests/monitor-imagestream-smoketest.yaml 
```

To verify the metrics, you can make a request to the `/metrics` path of the service:

```
$ curl -sk https://$(oc get svc monitor-imagestream-import -o jsonpath={.spec.clusterIP}:{.spec.ports[0].port})/metrics | grep imagestream
```

Alerting
--------

A few potentially interesting things to query on once you have this smoketest running and Prometheus is scraping it:

Check if the last success is more than 5 minutes old:

```
count(time() - imagestream_import_last_run{result="success"} > 300)
```


Count the number of failures in the last 30 minutes:

```
count_over_time(imagestream_import_last_run{result="failed"}[30m])
```


Updating Vendored Dependencies
-------------------

This project uses Glide, to update the dependencies run:

```
$ glide up --strip-vendor
```
