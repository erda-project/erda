import React from 'react';
import { AdminApi } from '../actions/Api';
import Highlighter from 'react-highlight-words';
import { Space, Input, Button, Divider, Table, Popconfirm, message } from 'antd';
import {
    SearchOutlined,
} from '@ant-design/icons';

interface State {
    filteredInfo:any,
    sortedInfo:any,
    data:any[],
    searchText:string,
    searchedColumn:string,
    loading:boolean,
    selectedRowKeys:string[],
}

export default class ESIndicesContent extends React.Component {
    
    api:AdminApi;
    state:State = {
        filteredInfo: {},
        sortedInfo: {},
        data: [],
        searchText: '',
        searchedColumn: '',
        loading: true,
        selectedRowKeys: [],
    };
    searchInput:any
    indices:any = [];
    LoadIndices:(init:boolean) => void;
    constructor(props:any) {
        super(props);
        this.api = new AdminApi();
        this.LoadIndices = (init:boolean) => {
            let state = {
                filteredInfo: {},
                sortedInfo: {},
                data: [],
                searchText: '',
                searchedColumn: '',
                loading: true,
                selectedRowKeys: [],
            }
            if (init) {
                this.state = state
            } else {
                this.setState(state);
            }
            this.api.ESIndices({
                Success: (data:any) => {
                    console.log("Load Indices:",data);
                    if(!data || !data.data) {
                        return;
                    }
                    let list = data.data;
                    let row = []
                    for (let i=0; i<list.length; i++) {
                        let item = list[i];
                        for (let j=0; j<item.indices.length; j++) {
                            let index = item.indices[j];
                            index.key = index.index,
                            index.time = index['timestamp']?new Date(index['timestamp']).toLocaleString():'';
                            row.push(index);
                        }
                    }
                    this.indices = row
                    this.setState({
                        data: row,
                        loading: false,
                    });
                },
                Failture: (error:any) => {
                    console.error('Error:', error);
                    message.error(error+'');
                },
            });
        };
        this.LoadIndices(true);
    }

    handleChange = (pagination:any, filters:any, sorter:any) => {
        this.setState({
          filteredInfo: filters,
          sortedInfo: sorter,
        });
    };

    handleSearch = (selectedKeys:any, confirm:any, dataIndex:any) => {
        let data:any[] = []
        let value = selectedKeys[0];
        for(let i=0; i<this.indices.length; i++) {
            let record = this.indices[i];
            if(record[dataIndex]?record[dataIndex].toString().toLowerCase().includes(value.toLowerCase()):'') {
                data.push(record);
            }
        }
        this.setState({
            searchText: value,
            searchedColumn: dataIndex,
            data: data,
        });
        confirm();
    };
    handleReset = (clearFilters:any) => {
        clearFilters();
        this.setState({ searchText: '' });
    };

    getColumnSearchProps = (dataIndex:any) => ({
        filterDropdown: ({ setSelectedKeys, selectedKeys, confirm, clearFilters }:any) => (
            <div style={{ padding: 8 }}>
            <Input
                ref={node => {
                    this.searchInput = node;
                }}
                placeholder={`Search ${dataIndex}`}
                value={selectedKeys[0]}
                onChange={e => setSelectedKeys(e.target.value ? [e.target.value] : [])}
                onPressEnter={() => this.handleSearch(selectedKeys, confirm, dataIndex)}
                style={{ width: 188, marginBottom: 8, display: 'block' }}
            />
            <Space>
                <Button
                    type="primary"
                    onClick={() => this.handleSearch(selectedKeys, confirm, dataIndex)}
                    icon={<SearchOutlined />}
                    size="small"
                    style={{ width: 90 }}
                >
                    Search
                </Button>
                <Button onClick={() => this.handleReset(clearFilters)} size="small" style={{ width: 90 }}>
                    Reset
                </Button>
            </Space>
            </div>
        ),
        filterIcon: (filtered:any) => <SearchOutlined style={{ color: filtered ? '#1890ff' : undefined }} />,
        onFilterDropdownVisibleChange: (visible:any) => {
            if (visible) {
                setTimeout(() => this.searchInput.select(), 100);
            }
        },
        render: (text:any) => this.state.searchedColumn === dataIndex ? (
            <Highlighter
                highlightStyle={{ backgroundColor: '#ffc069', padding: 0 }}
                searchWords={[this.state.searchText]}
                autoEscape
                textToHighlight={text ? text.toString() : ''}
            />
        ) : (text),
    });

    deleteIndices = () => {
        let indices = this.state.selectedRowKeys.join(",")
        if (indices.length <= 0) {
            return
        }
        this.api.ESDeleteIndices({
            Wildcard: indices,
            Success: (data:any) => {
                console.log("Delete Indices Response:", data);
                this.LoadIndices(false);
            },
            Failture: (error:any) => {
                console.error('Error:', error);
                message.error(error+'');
            },
        });
    };
    
    render() {
    let columns:any = [
        {
            title: 'Health',
            dataIndex: 'health',
            key: 'health',
            sorter: (a:any, b:any) => a.health<b.health?-1:(a.health>b.health?1:0),
            sortOrder: this.state.sortedInfo.columnKey === 'health' && this.state.sortedInfo.order,
            ellipsis: true,
            width: 75,
        },
        {
            title: 'Status',
            dataIndex: 'status',
            key: 'status',
            sorter: (a:any, b:any) => a.status<b.status?-1:(a.status>b.status?1:0),
            sortOrder: this.state.sortedInfo.columnKey === 'status' && this.state.sortedInfo.order,
            ellipsis: true,
            width: 75,
        },
        {
            title: 'Index',
            dataIndex: 'index',
            key: 'index',
            filteredValue: true,
            sorter: (a:any, b:any) => a.index<b.index?-1:(a.index>b.index?1:0),
            sortOrder: this.state.sortedInfo.columnKey === 'index' && this.state.sortedInfo.order,
            ellipsis: true,
            ...this.getColumnSearchProps("index")
        },
        {
            title: 'Pri',
            dataIndex: 'pri',
            key: 'pri',
            filteredValue: true,
            sorter: (a:any, b:any) => a.pri-b.pri,
            sortOrder: this.state.sortedInfo.columnKey === 'pri' && this.state.sortedInfo.order,
            ellipsis: true,
            width: 55,
        },
        {
            title: 'Rep',
            dataIndex: 'rep',
            key: 'rep',
            filteredValue: true,
            sorter: (a:any, b:any) => a.rep-b.rep,
            sortOrder: this.state.sortedInfo.columnKey === 'rep' && this.state.sortedInfo.order,
            ellipsis: true,
            width: 55,
        },
        {
            title: 'Docs Count',
            dataIndex: 'docs.count',
            key: 'docs.count',
            sorter: (a:any, b:any) => a['docs.count'] - b['docs.count'],
            sortOrder: this.state.sortedInfo.columnKey === 'docs.count' && this.state.sortedInfo.order,
            ellipsis: true,
            width: 150,
        },
        {
            title: 'Store Size',
            dataIndex: 'store.size',
            key: 'store.size',
            sorter: (a:any, b:any) => a.store_size_value - b.store_size_value,
            sortOrder: this.state.sortedInfo.columnKey === 'store.size' && this.state.sortedInfo.order,
            ellipsis: true,
            width: 150,
        },
        {
            title: 'Time',
            dataIndex: 'time',
            key: 'time',
            sorter: (a:any, b:any) => a.timestamp - b.timestamp,
            sortOrder: this.state.sortedInfo.columnKey === 'time' && this.state.sortedInfo.order,
            ellipsis: true,
            width: 200,
        }
    ];

    const rowSelection = {
        onChange: (selectedRowKeys:any, selectedRows:any) => {
            if (selectedRowKeys.length > 15) {
                message.warn("一次最多只能删除15个索引");
                selectedRowKeys.list.slice
            }
            this.setState({
                selectedRowKeys: selectedRowKeys,
            });
        },
        getCheckboxProps: (record:any) => ({
          disabled: record.timestamp <= 0, 
          name: record.index,
        }),
    };

    return (
    <>
        <Divider orientation="left">索引查询</Divider>
        <Table 
            style={{
                marginTop: '10px',
            }}
            rowSelection={{
                type: 'checkbox',
                hideSelectAll: true,
                selectedRowKeys: this.state.selectedRowKeys,
                ...rowSelection,
            }}
            columns={columns}
            loading={this.state.loading}
            dataSource={this.state.data} onChange={this.handleChange} 
            pagination={false} 
        />
        
        <Popconfirm
            title={() =>  {
                return (
                    <div>
                        确定要删除如下索引吗:<br/>
                        {this.state.selectedRowKeys.map((field:any, index:any) => (
                            <div key={field}>{field}</div>
                        ))}
                    </div>
                );
            }}
            onConfirm={this.deleteIndices}
            okText="Yes"
            cancelText="No"
            disabled={this.state.selectedRowKeys.length<=0}
        >
            <Button danger
                disabled={this.state.selectedRowKeys.length<=0}
                style={{
                    margin: '20px 0 15px 0',
                }}
            >删除选择索引</Button>
        </Popconfirm>
    </>
    );
    }
}