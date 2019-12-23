import React from "react";
import axios from "axios";
import {
  Spin,
  message,
  Table,
  Icon,
  Row,
  Col,
  Select,
  Button,
  Card
} from "antd";

import "./App.sass";

const { Option } = Select;
const oneProxyURL = window.location.origin + "/proxies/one";

// copy 将内容复制至粘贴板
export function copy(value, parent) {
  if (!document.execCommand) {
    return new Error("The browser isn't support copy.");
  }
  const input = document.createElement("input");
  input.value = value;
  if (parent) {
    parent.appendChild(input);
  } else {
    document.body.appendChild(input);
  }
  input.focus();
  input.select();
  document.execCommand("Copy", false, null);
  input.remove();
}

class App extends React.Component {
  state = {
    pageSize: 10,
    total: 0,
    loading: true,
    originalProxies: null,
    oneProxyURL,
    selectCategory: "",
    selectSpeed: "",
    proxies: null
  };
  async componentDidMount() {
    this.setState({
      loading: true
    });
    try {
      const { data } = await axios.get("/proxies");
      const proxies = data.proxies || [];
      proxies.forEach(item => {
        item.speed = item.speed || 0;
      });
      proxies.sort((a, b) => {
        return b.detectedAt - a.detectedAt;
      });
      this.setState({
        proxies,
        originalProxies: proxies
      });
    } catch (err) {
      message.error(err.message);
    } finally {
      this.setState({
        loading: false
      });
    }
  }
  handleChange(pagination, filters, sorter) {
    const { originalProxies } = this.state;
    if (!originalProxies) {
      return;
    }
    let arr = [];
    const filterKeys = Object.keys(filters);
    originalProxies.forEach(item => {
      if (filterKeys.length === 0) {
        arr.push(item);
        return;
      }
      let matched = true;
      filterKeys.forEach(key => {
        const values = filters[key];
        if (!values || values.length === 0) {
          return;
        }
        if (!filters[key].includes(item[key])) {
          matched = false;
        }
      });
      if (matched) {
        arr.push(item);
      }
    });
    const { field, order } = sorter;
    arr.sort((a, b) => {
      return a[field] - b[field];
    });
    if (order === "descend") {
      arr = arr.reverse();
    }

    this.setState({
      proxies: arr
    });
  }
  renderAvailablePorxySelector() {
    const { loading, oneProxyURL, selectCategory, selectSpeed } = this.state;
    if (loading) {
      return;
    }
    let requestUrl = oneProxyURL;
    const arr = [];
    if (selectCategory) {
      arr.push(`category=${selectCategory}`);
    }
    if (selectSpeed) {
      arr.push(`spped=${selectSpeed}`);
    }
    if (arr.length !== 0) {
      requestUrl += `?${arr.join("&")}`;
    }
    return (
      <Card className="proxySelector" title="Get available proxy" size="small">
        <p>Select cateogry and speed to generate the request.</p>
        <Row gutter={8}>
          <Col span={16}>
            <Button
              onClick={e => {
                try {
                  copy(requestUrl, e.target);
                  message.info("The URL was copied successfully.");
                } catch (err) {
                  message.error(err.message);
                }
              }}
              type="dashed"
              style={{
                width: "100%",
                textAlign: "left"
              }}
            >
              <Icon type="api" />
              <span>{requestUrl}</span>
            </Button>
          </Col>
          <Col span={4}>
            <Select
              defaultValue=""
              style={{
                width: "100%"
              }}
              onChange={value => {
                this.setState({
                  selectCategory: value
                });
              }}
            >
              <Option value="http">http</Option>
              <Option value="https">https</Option>
              <Option value="">http(s)</Option>
            </Select>
          </Col>
          <Col span={4}>
            <Select
              defaultValue=""
              style={{
                width: "100%"
              }}
              onChange={value => {
                this.setState({
                  selectSpeed: value,
                })
              }}
            >
              <Option value="">all</Option>
              <Option value="0">{"<750ms"}</Option>
              <Option value="1">{"<150ms"}</Option>
              <Option value="2">{">=150ms"}</Option>
            </Select>
          </Col>
        </Row>
      </Card>
    );
  }
  renderProxyList() {
    const { loading, proxies, pageSize } = this.state;
    if (loading) {
      return;
    }
    const columns = [
      {
        title: "IP",
        dataIndex: "ip",
        key: "ip"
      },
      {
        title: "Port",
        dataIndex: "port",
        key: "port"
      },
      {
        title: "Speed",
        dataIndex: "speed",
        key: "speed",
        sorter: true
      },
      {
        title: "Type",
        dataIndex: "category",
        key: "category",
        filters: [
          { text: "http", value: "http" },
          { text: "https", value: "https" }
        ]
      },
      {
        title: "Anonymous",
        dataIndex: "anonymous",
        key: "anonymous",
        render: v => {
          if (v) {
            return "YES";
          }
          return "NO";
        }
      },
      {
        title: "DetectedAt",
        dataIndex: "detectedAt",
        key: "detectedAt",
        sorter: true,
        render: v => {
          const date = new Date(v * 1000);
          return `${date.toLocaleDateString()} ${date.toLocaleTimeString()}`;
        }
      }
    ];
    return (
      <Table
        tableLayout="fixed"
        pagination={{
          total: proxies && proxies.length,
          pageSize,
          showSizeChanger: true,
          onShowSizeChange: (current, size) => {
            this.setState({
              pageSize: size
            });
          }
        }}
        columns={columns}
        dataSource={proxies}
        onChange={this.handleChange.bind(this)}
      />
    );
  }
  render() {
    const { loading } = this.state;

    return (
      <div className="App">
        <header className="header">
          <Icon type="api" />
          Free Proxy
        </header>
        <Spin spinning={loading} tip="Loading..." />
        <div className="contentWrapper">
          {this.renderAvailablePorxySelector()}
          {this.renderProxyList()}
        </div>
      </div>
    );
  }
}

export default App;
