# Security boundaries

Report suspected problems privately to the repository owner; do not include
access codes, sessions, personal projects or printer serials in public issues.

Private per-user resources, OIDC issuer/subject identity, CSRF/origin checks,
protected NTFS app data and expected-account guards come from the shared app host.
Printer codes use Windows DPAPI, with no browser persistence. Printer connections
are limited to private IPv4 LAN destinations, bound to the resolved IP and
explicit device certificate pins; passive FTP cannot redirect elsewhere.

Only typed printer operations exist. No raw MQTT, arbitrary executable, shell
script, host file path or uploaded preset execution route is provided. Start
requests are durable and idempotent, and ambiguous outcomes require reconciliation.
Archive/entity/numeric/size checks and a bounded Windows job reduce exposure;
they do not constitute an independent native-code sandbox audit. Use trusted
projects and inspect every print. Keep the host and firmware maintained.

Remote access uses One Workspace HTTPS, authentication and tunnel policy. Never
forward printer MQTT/FTPS ports to the Internet. Review device certificate changes
and physically verify the target before trust or printing.
