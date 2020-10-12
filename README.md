# mqtt2exporter

```mqtt2exporter``` listens to a mqtt broker and expose a prometheus-compatible exporter.

It tries to guess as much as it can the device type, id and data to expose. It can be tuned with a config file.

Default supported devices are:
- shellies
- zigbee2mqtt
- somfy
- teleinfo