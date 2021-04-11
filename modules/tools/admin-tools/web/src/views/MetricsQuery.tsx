import React from 'react';
import { MonitorApi } from '../actions/Api';
import { Input, Table, Divider, DatePicker, message } from 'antd';
import moment from 'moment';
import 'moment/locale/zh-cn';

const { Search } = Input;
const { RangePicker } = DatePicker;

interface MetricQueryContentState {
    columns:any[],
    values:any[],
    timeRange:any,
}

export default class MetricQueryContent extends React.Component {
    
    api:MonitorApi;
    state:MetricQueryContentState;
    start:number;
    end:number;
    constructor(props:any) {
        super(props);
        let now = new Date();
        this.end = now.getTime();
        this.start = this.end - 60*60*1000;
        this.api = new MonitorApi();
        this.state = {
            columns: [], values: [], 
            timeRange: [moment(new Date(this.start)), moment(now)],
        };
    }

    onTimeChange = (value:any) => {
        this.start = value[0].toDate().getTime();
        this.end = value[1].toDate().getTime();
        this.setState({ timeRange: [moment(value[0].toDate()), moment(value[1].toDate())] });
    }

    onSearch = (value:string) => {
        this.api.QueryMetric({
            Query: value,
            Start: this.start,
            End: this.end,
            Success: (data:any) => {
                console.log("Data:",data);
                let columns:any = [];
                let values:any = [];
                if(!data || 
                    !data.results || data.results.length <= 0 || 
                    !data.results[0].series || data.results[0].series.length <=0 ||
                    !data.results[0].series[0].values || data.results[0].series[0].values.length <=0) {
                    this.setState({columns: columns, values: values});
                    return;
                }
                let series = data.results[0].series[0];
                for(let i=0; i<series.columns.length; i++) {
                    let col = series.columns[i];
                    let key = i+'';
                    columns.push({
                        title: col,
                        dataIndex: key,
                        key: key,
                        width: 256,
                    });
                }
                for(let i=0; i<series.values.length; i++) {
                    let val = series.values[i];
                    let data:any = { key: i+'' };
                    for(let j=0; j<val.length; j++) {
                        data[j+''] = val[j]
                    }
                    values.push(data);
                }
                this.setState({columns: columns, values: values});
            },
            Failture: (error:any) => {
                console.error('Error:', error);
                this.setState({columns: [], values: []});
                message.error(error+'');
            },
        })
    }
      
    render() {
    return (
        <>
        <RangePicker showTime defaultValue={this.state.timeRange} onChange={this.onTimeChange} />
        <Divider orientation="left">查询语句</Divider>
        <Search
            placeholder="SELECT * FROM host_summary WHERE host_ip='192.168.1.100'"
            onSearch={this.onSearch}
            style={{ width: '100%' }}
            size="large"
            defaultValue='SELECT * FROM host_summary;'
            enterButton
        />
        <Divider orientation="left">查询结果</Divider>
        <Table dataSource={this.state.values} columns={this.state.columns} scroll={{ y: 600 }} />
        </>
    );
    }
}