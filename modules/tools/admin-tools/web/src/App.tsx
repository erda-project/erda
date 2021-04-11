import React from 'react';
import { 
  HashRouter as Router,
  Switch,
  Route,
  Link,
} from 'react-router-dom';
import 'antd/dist/antd.css';
import './assets/css/index.css';
import { Layout, Menu, Popover, message } from 'antd';
import {
  MenuUnfoldOutlined,
  MenuFoldOutlined,
  FileSearchOutlined,
  PieChartOutlined,
  AliyunOutlined,
  DatabaseOutlined,
  AppstoreOutlined,
  ProfileOutlined,
  InfoCircleOutlined,
  ApartmentOutlined,
} from '@ant-design/icons';
import MetricQueryContent from './views/MetricsQuery';
import KafkaPushContent from './views/KafkaPush';
import OfflineMachineContent from './views/OfflineMachine';
import ESIndicesContent from './views/ESIndices';
import EnvsContent from './views/Envs';
import NodeTopoContent from './views/NodeTopo';
import { AdminApi } from './actions/Api';

const { Header, Sider, Content } = Layout;
const { SubMenu } = Menu;

export default class App extends React.Component {
  adminApi:AdminApi;
  state = {
    collapsed: false,
    version: {
      version: '',
		  commitID: '',
		  goVersion: '',
		  buildTime: '',
		  dockerImage: '',
    }
  };
  constructor(props:any) {
    super(props);
    this.adminApi = new AdminApi();
    this.adminApi.Version({
      Success: (data:any) => {
          console.log("version:",data);
          if(!data || !data.data) {
            return;
          }
          let v = data.data;
          this.setState({
            version: {
              version: v.version,
              commitID: v.commit_id,
              goVersion: v.go_version,
              buildTime: v.build_time,
              dockerImage: v.docker_image,
            }
          });
      },
      Failture: (error:any) => {
          console.error('Error:', error);
          message.error(error+'');
      },
    })
  }

  toggle = () => {
    this.setState({
      collapsed: !this.state.collapsed,
    });
  };

  render() {
    const versionContent = (
      <div>
        <p><span className='line-list-title'>Version:</span>{this.state.version.version}</p>
        <p><span className='line-list-title'>Commit:</span>{this.state.version.commitID}</p>
        <p><span className='line-list-title'>Build:</span>{this.state.version.buildTime}</p>
        <p><span className='line-list-title'>GoVersion:</span>{this.state.version.goVersion}</p>
        <p><span className='line-list-title'>DockerImage:</span>{this.state.version.dockerImage}</p>
      </div>
    );
    return (
      <Router>
        <Layout style={{height:'100%'}} >
          <Sider trigger={null} collapsible collapsed={this.state.collapsed}>
            <div className="logo"><AliyunOutlined style={{ color: 'white', fontSize: '32px' }}/></div>
            <Menu theme="dark" mode="inline">
              <Menu.Item key="metric-query" icon={<PieChartOutlined />}><Link to="/">指标</Link></Menu.Item>
              <Menu.Item key="log-query" icon={<ProfileOutlined />}>日志</Menu.Item>
              <Menu.Item key="node-topo" icon={<ApartmentOutlined />}><Link to="/node/topo">网络拓扑</Link></Menu.Item>
              <SubMenu key="Monitor" icon={<AppstoreOutlined />} title="Monitor">
                <Menu.Item key="monitor-envs"><Link to="/envs">环境变量</Link></Menu.Item>
                <Menu.Item key="monitor-offilne-machine"><Link to="/machine/offline">机器下线</Link></Menu.Item>
              </SubMenu>
              <SubMenu key="ElasticSearch" icon={<FileSearchOutlined />} title="ElasticSearch">
                <Menu.Item key="3"><Link to="/es/indices">索引</Link></Menu.Item>
              </SubMenu>
              <SubMenu key="Cassandra" icon={<DatabaseOutlined />} title="Cassandra">
                <Menu.Item key="5">清理</Menu.Item>
              </SubMenu>
              <SubMenu key="Kafka" icon={<DatabaseOutlined />} title="Kafka">
                <Menu.Item key="6"><Link to="/kafka/push">推送数据</Link></Menu.Item>
              </SubMenu>
            </Menu>
          </Sider>
          <Layout className="site-layout">
            <Header className="site-layout-background" style={{ padding: 0, overflow: 'hidden' }}>
              {React.createElement(this.state.collapsed ? MenuUnfoldOutlined : MenuFoldOutlined, {
                className: 'trigger',
                onClick: this.toggle,
              })}
              <div className='line-list'>
                <div className='line-list-item'>
            <span className='line-list-title'>Version:</span><span>{this.state.version.version}</span>
                </div>
                <div className='line-list-split'/>
                <div className='line-list-item'>
                  <span className='line-list-title'>Commit:</span><span>{this.state.version.commitID.substring(0,8)}</span>
                </div>
                <div className='line-list-split'/>
                <div className='line-list-item'>
                  <span className='line-list-title'>Build:</span><span>{this.state.version.buildTime}</span>
                </div>
                &nbsp;&nbsp;
                <Popover content={versionContent} title="Version Info" trigger="hover" placement="bottomRight">
                  <InfoCircleOutlined />
                </Popover>
              </div>
            </Header>
            <Content
              className="site-layout-background"
              style={{
                margin: '24px 16px',
                padding: 24,
                minHeight: 280,
                overflow: 'scroll',
                borderRadius: 5,
              }}
            >
              <Switch>
                <Route exact path="/">
                  <MetricQueryContent />
                </Route>
                <Route exact path="/node/topo">
                  <NodeTopoContent />
                </Route>
                <Route exact path="/envs">
                  <EnvsContent />
                </Route>
                <Route exact path="/machine/offline">
                  <OfflineMachineContent />
                </Route>
                <Route exact path="/es/indices">
                  <ESIndicesContent />
                </Route>
                <Route path="/kafka/push">
                  <KafkaPushContent />
                </Route>
              </Switch>
            </Content>
          </Layout>
        </Layout>
      </Router>
    );
  }
}