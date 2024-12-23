# MyStrom Switch Exporter

Prometheus Exporter for myStrom Switch data.
The exporter collects data via the [HTTP API](https://api.mystrom.ch/).

## Running the Exporter

Run the docker image using:

```sh
docker run --network host ghcr.io/maxbischoff/mystrom-switch-exporter:latest
```

Running in the host network ensure the exporter can access
You can override the addres to listen on using the `-listen-addr` flag (default: `:8000`), e.g.:

```sh
docker run --network host ghcr.io/maxbischoff/mystrom-switch-exporter:latest -listen-addr 127.0.0.1:8000
```

## Configuring Prometheus to scrape the exporter

Following config-snippet can be used to configure a scrape job for scraping this exporter:

```yaml
- job_name: mystrom-switch
  metrics_path: /collect
  scrape_interval: 1m # collect data every minute
  static_configs:
    - targets:
      - http://192.168.0.27 # IP of first myStrom switch
      - http://192.168.0.28 # IP of second myStrom switch
  relabel_configs:
    - source_labels: [__address__]
      target_label: __param_target
    - source_labels: [__param_target]
      target_label: instance
    - target_label: __address__
      replacement: localhost:8000 # The address that the myStrom switch exporter can be reached on
```
