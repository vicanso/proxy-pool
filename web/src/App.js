import React from "react";
import axios from "axios";
import { Spin, message, Table } from "antd";

import "./App.sass";

class App extends React.Component {
  state = {
    pageSize: 10,
    total: 0,
    loading: true,
    originalProxies: null,
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
    console.dir(pagination);
    const { originalProxies } = this.state;
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
  render() {
    const { loading, proxies, pageSize } = this.state;
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
      <div className="App">
        <Spin spinning={loading} tip="Loading..." />
        {!loading && (
          <Table
            // scroll={{ y: 240 }}
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
            style={{
              margin: "30px",
              width: "100%",
              backgroundColor: "#f0f0f0"
            }}
          />
        )}
      </div>
    );
  }
}

export default App;
