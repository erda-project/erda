# encoding: utf8

from django.db import connection
import feature
import datetime


collector_filename = ""


if __name__ == "__main__":
    print("Running Erda migration in Python")
    for task in feature.entries:
        print("run task: 20210701-my-feature.py.%s" % (task.__name__))
        task()

    [print(query) for query in connection.queries]
    if len(collector_filename) > 0:
        with open(collector_filename, "a+") as collector:
            for query in connection.queries:
                try:
                    collector.write('/*-Python BEGIN: {}-*/\n'.format(datetime.datetime.now(datetime.timezone.utc).isoformat()))
                    collector.write(query["sql"])
                    collector.write("  /*-LINE END-*/\n")
                except Exception as e:
                    print(e)

