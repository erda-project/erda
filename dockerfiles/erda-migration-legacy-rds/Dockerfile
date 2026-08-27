FROM registry.erda.cloud/erda/erda-migratoin:57-20260504230057-9dafe6754dc644c375514a9a0818b650e9791ef0

COPY dbmigration-legacy /usr/bin/erda-migrate

RUN chmod 0755 /usr/bin/erda-migrate

ENTRYPOINT ["/usr/bin/erda-migrate"]
