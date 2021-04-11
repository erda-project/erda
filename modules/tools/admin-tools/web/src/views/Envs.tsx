import React from 'react';
import { AdminApi } from '../actions/Api';
import { List, message } from 'antd';

interface State {
    envs:string[],
    loading:boolean,
}

export default class EnvsContent extends React.Component {
    
    api:AdminApi;
    state:State = {
        envs: [],
        loading: true,
    };
    constructor(props:any) {
        super(props);
        this.api = new AdminApi();
        this.api.Envs({
            Success: (data:any) => {
                console.log("Load Envs:",data);
                if(!data || !data.data) {
                    return;
                }
                let list = data.data;
                this.setState({
                    envs: list,
                    loading: false,
                });
            },
            Failture: (error:any) => {
                console.error('Error:', error);
                message.error(error+'');
            },
        })
    }

   
    
    render() {
    return (
    <>
        <List
            header={<div>环境变量</div>}
            bordered
            loading={this.state.loading}
            dataSource={this.state.envs}
            renderItem={item => (
                <List.Item>
                    {item}
                </List.Item>
            )}
        />
    </>
    );
    }
}