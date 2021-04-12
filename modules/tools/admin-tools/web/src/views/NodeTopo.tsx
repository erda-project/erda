import React from 'react';
import { MonitorApi, AdminApi } from '../actions/Api';
import { Input, Button, DatePicker, Divider, message } from 'antd';
import ReactEcharts from "echarts-for-react";
import moment from 'moment';
import 'moment/locale/zh-cn';
import { SearchOutlined } from '@ant-design/icons';

const { RangePicker } = DatePicker;

interface State {
    option:any,
    timeRange:any,
    clusterName:string,
}

export default class NodeTopoContent extends React.Component {
    
    api:MonitorApi;
    adminApi:AdminApi;
    state:State;
    start:number;
    end:number;
    init:boolean;
    constructor(props:any) {
        super(props);
        this.api = new MonitorApi();
        this.adminApi = new AdminApi();
        this.init = false;
        let now = new Date();
        this.end = now.getTime();
        this.start = this.end - 60*60*1000;
        this.state = {
            option: {},
            clusterName: '',
            timeRange: [moment(new Date(this.start)), moment(now)],
        }
        this.getDefaultCluster();
    }

    getDefaultCluster = () => {
        this.adminApi.Envs({
            Key: "DICE_CLUSTER_NAME",
            Success: (data:any) => {
                this.init = true;
                console.log("Load Envs:",data);
                if(!data || !data.data) {
                    return;
                }
                this.setState({
                    clusterName: data.data,
                });
                this.loadTopo();
            },
            Failture: (error:any) => {
                this.init = true;
                console.error('Error:', error);
                message.error(error+'');
            },
        })
    }

    onTimeChange = (value:any) => {
        this.start = value[0].toDate().getTime();
        this.end = value[1].toDate().getTime();
        this.setState({ timeRange: [moment(value[0].toDate()), moment(value[1].toDate())] });
    }

    onClusterChange = (e:any) => {
        if (!this.init || !e || !e.target) return;
        this.setState({ clusterName: e.target.value });
    }

    onSearch = () => {
        this.loadTopo();
    }

    loadTopo = () => {
        this.api.NodeTopo({
            Start: this.start,
            End: this.end,
            Cluster: this.state.clusterName,
            Success: (data:any) => {
                console.log("Topo Data:", data);
                if (!data || !data.data) {
                    this.setState({
                        option: this.getOption({nodes:[], links:[]}),
                    });
                    return;
                }
                this.setState({
                    option: this.getOption(data.data),
                });
            },
            Failture: (error:any) => {
                console.error('Error:', error);
                message.error(error+'');
            },
        })
    }
    units = ["B", "KB", "MB", "GB", "TB", "PB", "EB"]
    formatBytes = (n:any) => {
        let i = 0;
        for(;i<this.units.length;) {
            if (n >= 1024) {
                i++
                n = n/1024
            } else {
                break;
            }
        }
        return (Math.floor(n*100)/100) + ' ' + this.units[i];
    }

    isBothway = (nm:any, n1:any, n2: any):boolean => {
        let dm = nm[n1];
        if(!dm) {
            return false;
        }
        if(!dm[n2]) {
            return false;
        }
        dm = nm[n2];
        if(!dm) {
            return false;
        }
        if(!dm[n1]) {
            return false;
        }
        return true;
    }

    getOption = (data:any):any => {
        let ns = data.nodes || [];
        let ls = data.links || [];
        let categories = [];
        let nodes = [];
        let links = [];
        let names = [];
        for (let i=0; i<ns.length; i++) {
            let node = ns[i];
            names.push(node.name);
            categories.push({
                name: node.name,
            });
            nodes.push({
                id: node.name,
                name: node.name,
                value: node.metrics.bytes,
                category: node.name,
                metrics: node.metrics,
            });
        }

        let nm:any = {}
        for (let i=0; i<ls.length; i++) {
            let line = ls[i];
            let sm = nm[line.source];
            if (!sm) {
                sm = {}
                nm[line.source] = sm;
            }
            let dm = sm[line.target]
            if (!dm) {
                sm[line.target] = true;
            }
        }
        for (let i=0; i<ls.length; i++) {
            let line = ls[i];
            let color = '#555';
            if (this.isBothway(nm, line.source,line.target)) {
                if (line.source > line.target) {
                    color = '#999';
                } else {
                    color = '#777';
                }
            }
            links.push({
                source: line.source,
                target: line.target,
                lineStyle: {
                    width: 2,
                    color: color,
                    curveness: 0.05,
                },
                label: {
                    normal: {
                        show: true,
                    },
                },
                metrics: line.metrics,
            });
        }
        return {
            legend: {
                data: names,
            },
            tooltip: {
                formatter: (params:any, ticket:any, callback:any) => {
                    let html = '';
                    if (params.dataType == 'edge') {
                        html += '<span style="color:#faad14">' + params.data.source + 
                            '</span> -> <span style="color:#52c41a">' + params.data.target + '</span><br>';
                    } else if (params.dataType == 'node') {
                        html += '<span style="color:#faad14">' + params.data.name + '</span><br>';
                    } else {
                        return '';
                    }
                    let m = params.data.metrics;
                    html += '<span style="color:#52c41a">Bytes: </span><span style="font-weight:bolder;color:#52c41a">' + this.formatBytes(m.bytes) + '</span><br>';
                    html += '<span style="color:#52c41a">Packets: </span><span style="font-weight:bolder;color:#52c41a">' + m.packets + '</span><br>';
                    html += '<span style="color:#40a9ff">TCP Bytes: </span><span style="font-weight:bolder;color:#40a9ff">' + this.formatBytes(m.tcp_bytes) + '</span><br>';
                    html += '<span style="color:#40a9ff">TCP Packets: </span><span style="font-weight:bolder;color:#40a9ff">' + m.tcp_packets + '</span><br>';
                    html += '<span style="color:#fafafa">UDP Bytes: </span><span style="font-weight:bolder;">' + this.formatBytes(m.udp_bytes) + '</span><br>';
                    html += '<span style="color:#fafafa">UDP Packets: </span><span style="font-weight:bolder;">' + m.udp_packets + '</span><br>';
                    return html;
                }
            },
            series: [{
                type: 'graph',
                layout: 'force',
                symbol: 'roundRect',
                symbolSize: [72, 38],
                focusNodeAdjacency: true,
                animation: false,
                roam: true,
                label: {
                    normal: {
                        show: true,
                        formatter: (params:any) => {
                            return params.data.name + '\n' + this.formatBytes(params.data.value) +'';
                        },
                        textStyle: {
                            fontSize: 12,
                        }
                    },
                },
                itemStyle: {
                    borderColor: '#fff',
                    borderWidth: 1,
                    shadowBlur: 5,
                    shadowColor: 'rgba(0, 0, 0, 0.3)'
                },
                categories: categories,
                edgeSymbol: ['circle', 'arrow'],
                edgeSymbolSize: [2, 10],
                edgeLabel: {
                    formatter: (params:any) => {
                        return this.formatBytes(params.data.metrics.bytes);
                    },
                },
                draggable: true,
                data: nodes,
                force: {
                    edgeLength: [200, 600],
                    repulsion: 5000,
                    gravity: 1,
                },
                edges: links,
            }]
        }
    }

    render() {
        return (
        <div>
            <Input.Group compact>
                <Input style={{ width: '30%', borderColor: this.state.clusterName?'':'red' }} placeholder="输入集群名" value={this.state.clusterName} onChange={this.onClusterChange} />
                <RangePicker showTime defaultValue={this.state.timeRange} style={{ width: '55%' }}  onChange={this.onTimeChange} />
                <Button style={{ width: '15%' }} type="primary" icon={<SearchOutlined />} onClick={this.onSearch}>查询</Button>
            </Input.Group>
            <Divider orientation="left">流量拓扑图</Divider>
            <ReactEcharts
                option={this.state.option}
                notMerge={true}
                lazyUpdate={true}
                style={{height: '700px', width: '100%'}}
            />
        </div>
        );
    }
}