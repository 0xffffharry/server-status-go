<template>
  <div class="page-container">
    <el-space direction="vertical" :size="0">
      <el-card class="show-card" style="font-size: 20px;" shadow="hover" header="Server Status">
        <el-form label-width="120px">
          <el-form-item :label="translate('Dark Mode')">
            <el-switch v-model="isDark" :inactive-icon="Sunny" :active-icon="Moon" @click="switchDark" />
          </el-form-item>
        </el-form>
      </el-card>
      <el-card class="show-card" shadow="hover" :header="translate('System Status')">
        <el-form label-width="120px">
          <el-form-item :label="translate('CPU Load')">
            <el-space wrap>
              <el-tag class="ml-2" :type="cpuLoadType(data.cpu_load_1min)">{{ data.cpu_load_1min }}</el-tag>
              <el-tag class="ml-2" :type="cpuLoadType(data.cpu_load_5min)">{{ data.cpu_load_5min }}</el-tag>
              <el-tag class="ml-2" :type="cpuLoadType(data.cpu_load_15min)">{{ data.cpu_load_15min }}</el-tag>
            </el-space>
          </el-form-item>
          <el-row>
            <el-space wrap>
              <el-col :span="6">
                <el-form-item :label="translate('CPU Usage')">
                  <el-progress type="dashboard" :percentage="percentageShow(data.cpu_usage)" :color="colors" />
                </el-form-item>
              </el-col>
              <el-col :span="6">
                <el-form-item :label="translate('RAM Usage')">
                  <el-progress type="dashboard" :percentage="percentageShow(data.ram_used * 100 / data.ram_total)"
                    :color="colors">
                    <template #default="{ percentage }">
                      <span class="percentage-value">{{ percentage }}%</span>
                      <span class="percentage-label">{{ convertSize(data.ram_used) }} / {{
                        convertSize(data.ram_total)
                      }} </span>
                    </template>
                  </el-progress>
                </el-form-item>
              </el-col>
            </el-space>
          </el-row>
          <el-row :gutter="20">
            <el-space wrap>
              <el-col :span="6">
                <el-form-item :label="translate('RAM Free')">
                  <el-progress type="dashboard" :percentage="percentageShow(data.ram_free * 100 / data.ram_total)"
                    :color="colors_rev">
                    <template #default="{ percentage }">
                      <span class="percentage-value">{{ percentage }}%</span>
                      <span class="percentage-label">{{ convertSize(data.ram_free) }} / {{
                        convertSize(data.ram_total)
                      }} </span>
                    </template>
                  </el-progress>
                </el-form-item>
              </el-col>
              <el-col :span="6">
                <el-form-item :label="translate('Swap Usage')">
                  <el-progress type="dashboard"
                    :percentage="percentageShow(100 - (data.swap_free * 100 / data.swap_total))" :color="colors">
                    <template #default="{ percentage }">
                      <span class="percentage-value">{{ percentage }}%</span>
                      <span class="percentage-label">{{ convertSize(data.swap_total - data.swap_free) }} / {{
                        convertSize(data.swap_total)
                      }} </span>
                    </template>
                  </el-progress>
                </el-form-item>
              </el-col>
            </el-space>
          </el-row>
        </el-form>
      </el-card>
      <el-card class="show-card" shadow="hover" :header="translate('System Info')">
        <el-form label-width="120px">
          <el-form-item :label="translate('Hostname')">
            <span>{{ data.system.hostname }}</span>
          </el-form-item>
          <el-form-item :label="translate('OS')">
            <span>{{ data.system.os }}</span>
          </el-form-item>
          <el-form-item :label="translate('OS Version')">
            <span>{{ data.system.os_version }}</span>
          </el-form-item>
          <el-form-item :label="translate('Kernel Version')">
            <span>{{ data.system.kernel_version }}</span>
          </el-form-item>
          <el-form-item :label="translate('Arch')">
            <span>{{ data.system.arch }}</span>
          </el-form-item>
          <el-form-item :label="translate('Uptime')">
            <span>{{ convertDuration(data.system.uptime) }}</span>
          </el-form-item>
        </el-form>
      </el-card>
      <el-card class="show-card" shadow="hover" :header="translate('Network')"
        v-if="data.network != null && data.network.length > 0">
        <el-form label-width="120px">
          <el-form-item v-for="item in data.network" :label="item.name">
            <el-col>
              <el-row>
                <el-space wrap>
                  <el-icon :size="15">
                    <Upload />
                  </el-icon>
                  <span>{{ convertSize(item.sent) }}</span>
                </el-space>
              </el-row>
              <el-row>
                <el-space wrap>
                  <el-icon :size="15">
                    <Download />
                  </el-icon>
                  <span>{{ convertSize(item.recv) }}</span>
                </el-space>
              </el-row>
            </el-col>
          </el-form-item>
        </el-form>
      </el-card>
      <el-card class="show-card" shadow="hover" :header="translate('Disk')"
        v-if="data.disk != null && data.disk.length > 0">
        <el-form label-width="150px">
          <el-form-item v-for="item in data.disk" :label="item.path + ' (' + item.device + ')'">
            <el-progress type="dashboard" :percentage="percentageShow(item.used * 100 / item.total)" :color="colors">
              <template #default="{ percentage }">
                <span class="percentage-value">{{ percentage }}%</span>
                <span class="percentage-label">{{ convertSize(item.used) }} / {{
                  convertSize(item.total)
                }} </span>
              </template>
            </el-progress>
          </el-form-item>
        </el-form>
      </el-card>
      <el-card class="show-card" shadow="hover" :header="translate('Temperature')"
        v-if="data.temperature != null && data.temperature.length > 0">
        <el-form label-width="150px">
          <el-form-item v-for="item in data.temperature" :label="item.key">
            <span>{{ item.temperature }}°C</span>
          </el-form-item>
        </el-form>
      </el-card>
    </el-space>
  </div>
</template>

<script setup>
import { ref } from "vue";
import {
  Upload,
  Download,
  Moon,
  Sunny,
} from '@element-plus/icons-vue';
import { useDark, useToggle } from "@vueuse/core";

const isDark = useDark()
const toggleDark = useToggle(isDark)
const switchDark = () => {
  if (!isDark.value) {
    toggleDark(false)
  } else {
    toggleDark(true)
  }
}

const colors = [
  { color: '#6f7ad3', percentage: 20 },
  { color: '#1989fa', percentage: 40 },
  { color: '#5cb87a', percentage: 60 },
  { color: '#e6a23c', percentage: 80 },
  { color: '#f56c6c', percentage: 100 },
];

const colors_rev = [
  { color: '#f56c6c', percentage: 20 },
  { color: '#e6a23c', percentage: 40 },
  { color: '#5cb87a', percentage: 60 },
  { color: '#1989fa', percentage: 80 },
  { color: '#6f7ad3', percentage: 100 },
];

const cpuLoadType = (n) => {
  if (n < 0.5) {
    return 'success';
  } else if (n < 0.7) {
    return 'info';
  } else if (n < 0.9) {
    return 'warning';
  } else {
    return 'danger';
  }
}

const percentageShow = (usage) => {
  if (usage == undefined || usage == null || isNaN(usage) || usage == 0) {
    return 0;
  }
  return parseFloat(usage.toFixed(2));
}

const convertSize = (bytes) => {
  const kilobyte = 1024;
  const megabyte = kilobyte * 1024;
  const gigabyte = megabyte * 1024;
  const terabyte = gigabyte * 1024;

  if (bytes < kilobyte) {
    return bytes + ' B';
  } else if (bytes < megabyte) {
    return (bytes / kilobyte).toFixed(2) + ' kB';
  } else if (bytes < gigabyte) {
    return (bytes / megabyte).toFixed(2) + ' MB';
  } else if (bytes < terabyte) {
    return (bytes / gigabyte).toFixed(2) + ' GB';
  } else {
    return (bytes / terabyte).toFixed(2) + ' TB';
  }
}

const convertDuration = (seconds) => {
  if (isNaN(seconds) || seconds < 0) {
    return "0s";
  }

  const years = Math.floor(seconds / (365 * 24 * 60 * 60));
  const days = Math.floor((seconds % (365 * 24 * 60 * 60)) / (24 * 60 * 60));
  const hours = Math.floor((seconds % (24 * 60 * 60)) / (60 * 60));
  const minutes = Math.floor((seconds % (60 * 60)) / 60);
  const remainingSeconds = seconds % 60;

  const yearsText = years > 0 ? years + "y " : "";
  const daysText = days > 0 ? days + "d " : "";
  const hoursText = hours > 0 ? hours + "h " : "";
  const minutesText = minutes > 0 ? minutes + "m " : "";
  const secondsText = remainingSeconds + "s";

  return yearsText + daysText + hoursText + minutesText + secondsText;
}

const translate_zh_cn = (word) => {
  const dict = {
    'Dark Mode': '暗黑模式',
    'System Info': '系统信息',
    'Hostname': '主机名',
    'OS': '操作系统',
    'OS Version': '操作系统版本',
    'Kernel Version': '内核版本',
    'Arch': '架构',
    'Uptime': '运行时间',
    'System Status': '系统状态',
    'CPU Load': 'CPU负载',
    'CPU Usage': 'CPU使用率',
    'RAM Usage': '内存使用量',
    'RAM Free': '内存空闲量',
    'Swap Usage': '交换空间使用量',
    'Network': '网络',
    'Disk': '硬盘',
    'Temperature': '温度',
  }
  let tword = dict[word];
  if (tword == undefined || tword == null) {
    return word;
  }
  return tword;
}

const translate = translate_zh_cn;

const data = ref({
  'cpu_load_1min': 0,
  'cpu_load_5min': 0,
  'cpu_load_15min': 0,
  'cpu_usage': 0,
  'ram_total': 0,
  'ram_used': 0,
  'ram_free': 0,
  'system': {
    'hostname': 'unknown',
    'arch': 'unknown',
    'os': 'unknown',
    'os_version': 'unknown',
    'kernel_version': 'unknown',
    'uptime': 'unknown',
  },
  'network': [],
  'disk': [],
  'temperature': [],
});

const connect_ws = () => {
  const currentURL = window.location.href;
  let u = new URL(currentURL);
  u.protocol = u.protocol.replace("http", "ws");
  u.pathname = "/ws";
  // Debug
  // u.port = 9066;
  // u.hostname = "10.0.0.254";
  //
  let ws = new WebSocket(u);
  ws.onopen = function () {
    console.log("WebSocket is open now.");
  };
  ws.onmessage = function (e) {
    let receivedData = JSON.parse(e.data);
    data.value = receivedData;
  };
  ws.onclose = function () {
    console.log("WebSocket is closed now. Reconnecting...");
    setTimeout(function () {
      connect_ws();
    }, 1000);
  };
  ws.onerror = function (err) {
    console.log("Error: ", err);
  };
}

connect_ws();

</script>

<style scoped>
.page-container {
  display: flex;
  justify-content: space-around;
  padding: 20px;
}

.show-card {
  width: 400px;
  margin-bottom: 20px;
}

.percentage-value {
  display: block;
  margin-top: 10px;
  font-size: 20px;
}

.percentage-label {
  display: block;
  margin-top: 10px;
  font-size: 9px;
}
</style>
