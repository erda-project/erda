import React from 'react';
import { MonitorApi } from '../actions/Api';
import { Form, Input, Button, Divider, message } from 'antd';
import { MinusCircleOutlined, PlusOutlined } from '@ant-design/icons';

export default class OfflineMachineContent extends React.Component {
    
    api:MonitorApi;
    constructor(props:any) {
        super(props);
        this.api = new MonitorApi();
    }

    formItemLayout = {
        labelCol: {
            xs: { span: 24 },
            sm: { span: 4 },
        },
        wrapperCol: {
            xs: { span: 24 },
            sm: { span: 20 },
        },
    };
    formItemLayoutWithOutLabel = {
        wrapperCol: {
            xs: { span: 24, offset: 0 },
            sm: { span: 20, offset: 4 },
        },
    };

    onFinish = (form:any) => {
        if(!form.hosts || !form.cluster ||form.hosts.length<= 0) {
            return
        }
        console.log('offline machine:', form.cluster, form.hosts);
        this.api.OfflineMachine({
            Cluster: form.cluster,
            Hosts: form.hosts,
            Success: (data:any) => {
                console.log("Data:",data);
                message.info("Success");
            },
            Failture: (error:any) => {
                console.error('Error:', error);
                message.error(error+'');
            },
        });
    };

    render() {
    return (
    <>
        <Divider orientation="left">机器下线</Divider>
        <Form {...this.formItemLayoutWithOutLabel} onFinish={this.onFinish}>
            <Form.Item name="cluster" label="Cluster Name" rules={[{ required: true }]} {...this.formItemLayout} >
                <Input style={{ width: '60%' }} />
            </Form.Item>
            <Form.List name="hosts">
            {(fields, { add, remove }) => {
                return (
                <div>
                    {fields.map((field, index) => (
                    <Form.Item
                        {...(index === 0 ? this.formItemLayout : this.formItemLayoutWithOutLabel)}
                        label={index === 0 ? 'Host IP' : ''}
                        required={false}
                        key={field.key}
                    >
                        <Form.Item
                        {...field}
                        validateTrigger={['onChange', 'onBlur']}
                        rules={[
                            {
                            required: true,
                            whitespace: true,
                            message: "host ip 不能为空",
                            },
                        ]}
                        noStyle
                        >
                        <Input placeholder="输入 host ip" style={{ width: '60%' }} />
                        </Form.Item>
                        {fields.length > 0 ? (
                        <MinusCircleOutlined
                            className="dynamic-delete-button"
                            style={{ margin: '0 8px' }}
                            onClick={() => {
                                remove(field.name);
                            }}
                        />
                        ) : null}
                    </Form.Item>
                    ))}
                    <Form.Item>
                        <Button
                            type="dashed"
                            onClick={() => {
                                add();
                            }}
                            style={{ width: '60%' }}
                        >
                            <PlusOutlined /> 添加 Host
                        </Button>
                    </Form.Item>
                </div>
                );
            }}
            </Form.List>

            <Form.Item>
                <Button type="primary" htmlType="submit">提交</Button>
            </Form.Item>
        </Form>
    </>
    );
    }
}