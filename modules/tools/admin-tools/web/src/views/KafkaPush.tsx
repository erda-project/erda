import React from 'react';
import { AdminApi } from '../actions/Api';
import { Input, Form, Button, Divider, Select, message } from 'antd';

const { TextArea } = Input;
const { Option } = Select;
interface KafkaPushContentState {
    topics:string[],
    payload:string,
}

export default class KafkaPushContent extends React.Component {
    api:AdminApi;
    state:KafkaPushContentState;
    topicOptions:any[]
    defaultTopics = ['spot-metrics']
    defaultPayload = `{
    "fields":{
        "value":1
    },
    "name":"test_data",
    "tags":{
        "tagkey":"tagvalue"
    },
    "timestamp":1600829940000000000
}`    
    constructor(props:any) {
        super(props);
        this.api = new AdminApi();
        this.state = {
            topics: this.defaultTopics,
            payload: this.defaultPayload,
        };
        let topics:string[] = [
            "spot-metrics",
            "spot-alert-event",
            "spot-analytics",
            "spot-container-log",
            "spot-error",
            "spot-job-log",
            "spot-metrics-temp",
            "spot-trace",
        ]
        this.topicOptions = []
        for (let i = 0; i < topics.length; i++) {
            let topic = topics[i];
            this.topicOptions.push(<Option key={topic} value={topic}>{topic}</Option>);
        }
    }
    
    layout = {
        labelCol: { span: 4 },
        wrapperCol: { span: 20 },
    };
    tailLayout = {
        wrapperCol: { offset: 4, span: 20 },
    };

    onPushButtonClick = (value:any) => {
        console.log(this.state.topics, this.state.payload)
        let topics:string = this.state.topics.join(",")
        this.api.KafkaPush({
            Topics: topics,
            Payload: this.state.payload,
            Success: (data:any) => {
                message.info("Success");
            },
            Failture: (error:any) => {
                console.log("Error:", error);
                message.error(error+'');
            }
        })
    }
    onTopicsChange = (value:string[]) => {
        console.log('topics:', value);
        this.setState({
            topics: value,
        })
    };
    onPlayloadChange = ({ target: { value } }:any) => {
        let payload = { value }.value
        console.log('payload:', payload);
        this.setState({
            payload: payload,
        })
    };

    render() {
    return (    
        <Form
            {...this.layout}
            name="basic"
            initialValues={{
                topics: this.defaultTopics,
                payload: this.defaultPayload,
            }}
        >
            <Divider orientation="left">数据推送</Divider>
            <Form.Item
                label="Topics"
                name="topics"
                rules={[{ required: true, message: '输入 Topics !' }]}
            >
                <Select
                    mode="tags"
                    placeholder="输入 Topics"
                    onChange={this.onTopicsChange}
                    value={this.state.topics}
                >
                    {this.topicOptions}
                </Select>
            </Form.Item>

            <Form.Item
                label="JSON数据"
                name="payload"
                rules={[{ required: true, message: '输入 JSON 数据!' }]}
            >
                <TextArea rows={15} value={this.state.payload} onChange={this.onPlayloadChange} />
            </Form.Item>
            <Form.Item {...this.tailLayout}>
                <Button type="primary" htmlType="button" onClick={this.onPushButtonClick}>推送</Button>
            </Form.Item>
        </Form>
    );
    }
}