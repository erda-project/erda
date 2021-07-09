# encoding: utf8

from django.db import connection
import feature


if __name__ == "__main__":
    print("Running Erda migration in Python")
    for task in feature.entries:
        print("run task: 20210701-my-feature.py.%s" % (task.__name__))
        task()
    [print(query) for query in connection.queries]

