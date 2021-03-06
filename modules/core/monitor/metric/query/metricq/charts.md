## Curve graph, Area graph
```json
{
    "success":true,
    "data":{
        "results":[
            {
                "data":[
                    {
                        "max.mem_used":{
                            "agg":"max",
                            "data":[
                                24204627968,
                                24210165760
                            ],
                            "name":"mem_usedMaximum",
                            "tag":"10.167.0.10",
                        }
                    }
                ],
                "name":"host_summary"
            }
        ],
        "time":[
            1599464700000,
            1599464760000
        ],
        "title":"",
        "total":6002
    }
}
```

## Histogram
```json
{
    "success":true,
    "data":{
        "results":[
            {
                "data":[
                    {
                        "max.mem_used":{
                            "agg":"max",
                            "axisIndex":null,
                            "chartType":null,
                            "data":[
                                24204627968,
                                24210165760
                            ],
                            "name":"mem_usedMaximum",
                            "tag":"10.167.0.10",
                            "unit":null,
                            "unitType":null
                        }
                    }
                ],
                "name":"host_summary"
            }
        ],
        "xData":[
            "192.168.0.1",
            "192.168.0.2"
        ],
        "title":"",
        "total":6002
    }
}
```

## Pie chart
```json
{
    "success":true,
    "data":{
        "metricData":[
            {
                "title":"10.167.0.10",
                "value":63
            }
        ]
    }
}
```

## Card map
```json
{
    "success":true,
    "data":{
        "metricData":[
            {
                "name":"containersMaximum",
                "value":86
            }
        ]
    }
}
```

## Form
```json
{
    "success":true,
    "data":{
        "cols":[
            {
                "dataIndex":"max.containers",
                "title":"containersMaximum"
            },
            {
                "dataIndex":"last.tags.host_ip",
                "title":"host_ip"
            }
        ],
        "metricData":[
            {
                "last.tags.host_ip":"10.167.0.27",
                "max.containers":86
            }
        ]
    }
}
```